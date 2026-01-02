package cache

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/99designs/keyring"
	_ "modernc.org/sqlite"
)

type Manager struct {
	ring       keyring.Keyring
	db         *sql.DB
	keyName    string
	ttlMinutes int
}

type Options struct {
	KeyringConfig keyring.Config
	Path          string
	KeyName       string
	TTLMinutes    int
}

func New(opts Options) (*Manager, error) {
	if opts.Path == "" {
		return nil, errors.New("cache path required")
	}
	p := opts.Path
	if p[0] == '~' {
		home, _ := os.UserHomeDir()
		p = filepath.Join(home, p[1:])
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", p)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS public_cache (key TEXT PRIMARY KEY, payload_json TEXT, updated_at INTEGER, payload BLOB, meta TEXT)`); err != nil {
		_ = db.Close()
		return nil, err
	}
	r, err := keyring.Open(opts.KeyringConfig)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	keyName := opts.KeyName
	if keyName == "" {
		keyName = "cache_key"
	}
	m := &Manager{db: db, ttlMinutes: opts.TTLMinutes, ring: r, keyName: keyName}
	return m, nil
}

func (m *Manager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

func (m *Manager) put(key string, payloadJSON []byte, payload []byte, meta string, updatedAt time.Time) error {
	enc, err := m.encrypt(payload)
	if err != nil {
		return err
	}
	_, err = m.db.Exec(`INSERT INTO public_cache(key, payload_json, updated_at, payload, meta) VALUES(?,?,?,?,?) ON CONFLICT(key) DO UPDATE SET payload_json=excluded.payload_json, updated_at=excluded.updated_at, payload=excluded.payload, meta=excluded.meta`, key, string(payloadJSON), updatedAt.Unix(), enc, meta)
	return err
}

func (m *Manager) Put(key string, payloadJSON []byte, payload []byte, meta string) error {
	return m.put(key, payloadJSON, payload, meta, time.Now())
}

func (m *Manager) PutWithTimestamp(key string, payloadJSON []byte, payload []byte, meta string, ts time.Time) error {
	return m.put(key, payloadJSON, payload, meta, ts)
}

func (m *Manager) Get(key string) (payloadJSON []byte, payload []byte, updatedAt time.Time, meta string, err error) {
	row := m.db.QueryRow(`SELECT payload_json, updated_at, payload, meta FROM public_cache WHERE key=?`, key)
	var pj string
	var ts int64
	var enc []byte
	if err = row.Scan(&pj, &ts, &enc, &meta); err != nil {
		return nil, nil, time.Time{}, "", err
	}
	dec, derr := m.decrypt(enc)
	if derr != nil {
		return nil, nil, time.Time{}, "", derr
	}
	return []byte(pj), dec, time.Unix(ts, 0), meta, nil
}

func (m *Manager) Delete(key string) error {
	_, err := m.db.Exec(`DELETE FROM public_cache WHERE key=?`, key)
	return err
}

func (m *Manager) DeletePrefix(prefix string) error {
	if prefix == "" {
		return errors.New("empty prefix")
	}
	_, err := m.db.Exec(`DELETE FROM public_cache WHERE key LIKE ?`, prefix+"%")
	return err
}

func (m *Manager) IsFresh(ts time.Time) bool {
	if m.ttlMinutes <= 0 {
		return false
	}
	return time.Since(ts) <= time.Duration(m.ttlMinutes)*time.Minute
}

func (m *Manager) encrypt(plain []byte) ([]byte, error) {
	key, err := m.loadOrGenerateKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, plain, nil)
	return ciphertext, nil
}

func (m *Manager) decrypt(ciphertext []byte) ([]byte, error) {
	key, err := m.loadOrGenerateKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ns := gcm.NonceSize()
	if len(ciphertext) < ns {
		return nil, errors.New("invalid ciphertext")
	}
	nonce := ciphertext[:ns]
	data := ciphertext[ns:]
	plain, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, err
	}
	return plain, nil
}

func (m *Manager) loadOrGenerateKey() ([]byte, error) {
	it, err := m.ring.Get(m.keyName)
	if err == nil && len(it.Data) == 32 {
		return it.Data, nil
	}
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	if err := m.ring.Set(keyring.Item{Key: m.keyName, Data: key}); err != nil {
		return nil, err
	}
	return key, nil
}

func MarshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}
