# Sufir's Keeper Client

## Политика TLS
- Клиент использует стандартное хранилище доверенных корней ОС (system trust store).
- TLS завершается на nginx; бекенд общается с nginx по HTTP.
- mTLS не используется; встроенные сертификаты и отключение проверки TLS запрещены.
- Для локальной разработки допускается dev CA, добавляемый в системный trust store контейнера; клиент может дополнительно подключать CA из файла.

## Конфигурация
- Источники: флаги CLI → ENV → конфиг‑файл → значения по умолчанию.
- Флаги:
  - `--config` путь к файлу конфигурации
  - `--server` базовый URL API, например `https://localhost:8443/api/v1`
  - `--log-level` уровень логирования `error|warn|info|debug`
  - `--ca-cert-path` путь к дополнительному CA (для dev)
- ENV:
  - `SUFIR_KEEPER_CONFIG` файл конфигурации
  - `SUFIR_KEEPER_SERVER` базовый URL API
  - `SUFIR_KEEPER_LOG_LEVEL` уровень логирования
  - `SUFIR_KEEPER_CA_CERT` путь к CA файлу
  - `SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE` имя сервиса в keyring
  - `SUFIR_KEEPER_AUTH_BACKEND` backend keyring (`file` для dev)
  - `SUFIR_KEEPER_AUTH_FILE_DIR` директория для file backend
  - `SUFIR_KEEPER_CACHE_PATH` путь к файлу кеша
  - `SUFIR_KEEPER_CACHE_TTL` TTL кеша в минутах
  - `SUFIR_KEEPER_CACHE_ENABLED` включение кеша (`true|false`)
- Конфиг‑ключи:
  - `server.base_url`
  - `tls.ca_cert_path`
  - `log.level`
  - `auth.token_store_service`, `auth.backend`, `auth.file_dir`
  - `cache.path`, `cache.ttl_minutes`, `cache.enabled`
- Значения по умолчанию:
  - `server.base_url`: `https://localhost:8443/api/v1`
  - `tls.ca_cert_path`: `./var/ca.crt`
  - `log.level`: `info`
  - `auth.token_store_service`: `sufir-keeper-client`
  - `cache.path`: `~/.local/share/sufir-keeper-client/cache.db`
  - `cache.ttl_minutes`: `180`
  - `cache.enabled`: `true`

## Команды CLI
- Аутентификация:
  - `keepcli register --login user --password pass`
  - `keepcli login --login user --password pass`
  - `keepcli status`
  - `keepcli logout`
  - Верификация и обновление токенов выполняются прозрачно в фоне при выполнении команд; отдельная команда не требуется.
- Записи:
  - `keepcli list --type TEXT --search x --limit 10 --offset 0`
  - `keepcli get <uuid>`
  - `keepcli create --title t --value v --meta k=v`
  - `keepcli update <uuid> --title t2 --value v2 --meta k=v`
  - `keepcli delete <uuid>`
  - Fallback на кеш: только для `list` и `get` при недоступности сети и валидном TTL; CRUD строго онлайн.
- Файлы:
  - `keepcli upload --path ./a.txt`
  - `keepcli download <uuid> ./out.bin`
  - Загрузка через Presigned POST; перед загрузкой client прозрачно делает presign и сразу начинает отправку; прогресс отображается в stdout.
- Автодополнение:
  - `keepcli completion bash` или `zsh|fish|powershell`
  - Bash: `keepcli completion bash > /etc/bash_completion.d/keepcli` (под root) или в `~/.bashrc`
  - Zsh: `keepcli completion zsh > ~/.zsh/completions/_keepcli`

## Запуск в Docker
- Запуск окружения: `docker compose up -d`
- Выполнение CLI внутри контейнера:
  - `docker compose exec go keepcli --server https://localhost:8443/api/v1 login`
  - Для dev TLS: `--ca-cert-path /workspace/var/ca.crt`
- Запуск проверок: `docker compose exec go bash -lc "bash tools/devcheck.sh"`

## Поведение кеша
- Файл кеша: `~/.local/share/sufir-keeper-client/cache.db` (0600)
- Шифрование AES‑256‑GCM; ключ хранится в OS keyring.
- Обновление кеша при успешных ответах API в `list/get`; инвалидация на `create/update/delete`.

## Логирование
- Уровни: `error|warn|info|debug`.
- Без включения токенов и приватных данных; логируются статусы и метаданные.

## Примеры создания и изменения по типам данных
- TEXT:
  - Создание: `keepcli create --title "Заметка" --type TEXT --value "Привет мир" --meta category=personal,tag=note`
  - Обновление значения: `keepcli update <uuid> --type TEXT --value "Новое значение"`
  - Обновление заголовка: `keepcli update <uuid> --title "Новый заголовок"`
- CREDENTIAL:
  - Создание: `keepcli create --title "Gmail" --type CREDENTIAL --login user@gmail.com --password strongpass --meta category=work`
  - Обновление части данных: `keepcli update <uuid> --type CREDENTIAL --password newpass`
  - Обновление логина: `keepcli update <uuid> --type CREDENTIAL --login newuser@gmail.com`
- CARD:
  - Создание: `keepcli create --title "Visa" --type CARD --card-number 4111111111111111 --card-holder "IVAN IVANOV" --expiry-date 12/25 --cvv 123 --meta category=personal`
  - Обновление владельца карты: `keepcli update <uuid> --type CARD --card-holder "IVAN PETROV"`
  - Обновление срока действия: `keepcli update <uuid> --type CARD --expiry-date 01/27`
