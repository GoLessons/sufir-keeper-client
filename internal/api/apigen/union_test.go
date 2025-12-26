package apigen

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestItemCreateUnionOperations(t *testing.T) {
	var u ItemCreate_Data
	require.NoError(t, u.FromCredentialData(CredentialData{Type: "CREDENTIAL", Login: "l", Password: "p"}))
	cd, err := u.AsCredentialData()
	require.NoError(t, err)
	require.Equal(t, "CREDENTIAL", string(cd.Type))
	require.NoError(t, u.MergeCredentialData(CredentialData{Type: "CREDENTIAL", Login: "l2"}))
	require.NoError(t, u.FromCardData(CardData{Type: "CARD", CardNumber: "4111111111111111", CardHolder: "h", ExpiryDate: "12/25", Cvv: "123"}))
	ca, err := u.AsCardData()
	require.NoError(t, err)
	require.Equal(t, "CARD", string(ca.Type))
	require.NoError(t, u.MergeCardData(CardData{Type: "CARD", CardHolder: "h2"}))
	require.NoError(t, u.FromTextData(TextData{Type: "TEXT", Value: "v"}))
	td, err := u.AsTextData()
	require.NoError(t, err)
	require.Equal(t, "TEXT", string(td.Type))
	require.NoError(t, u.MergeTextData(TextData{Type: "TEXT", Value: "v2"}))
	require.NoError(t, u.FromBinaryData(BinaryData{Type: "BINARY", Filename: "f", Id: uuid.Nil}))
	bd, err := u.AsBinaryData()
	require.NoError(t, err)
	require.Equal(t, "BINARY", string(bd.Type))
	d, err := u.Discriminator()
	require.NoError(t, err)
	require.Equal(t, "BINARY", d)
	val, err := u.ValueByDiscriminator()
	require.NoError(t, err)
	_ = val
}

func TestItemResponseUnionOperations(t *testing.T) {
	var u ItemResponse_Data
	require.NoError(t, u.FromCredentialData(CredentialData{Type: "CREDENTIAL", Login: "l", Password: "p"}))
	_, err := u.AsCredentialData()
	require.NoError(t, err)
	require.NoError(t, u.MergeCredentialData(CredentialData{Type: "CREDENTIAL", Login: "l2"}))
	require.NoError(t, u.FromCardData(CardData{Type: "CARD", CardNumber: "4111111111111111", CardHolder: "h", ExpiryDate: "12/25", Cvv: "123"}))
	_, err = u.AsCardData()
	require.NoError(t, err)
	require.NoError(t, u.MergeCardData(CardData{Type: "CARD", CardHolder: "h2"}))
	require.NoError(t, u.FromTextData(TextData{Type: "TEXT", Value: "v"}))
	_, err = u.AsTextData()
	require.NoError(t, err)
	require.NoError(t, u.MergeTextData(TextData{Type: "TEXT", Value: "v2"}))
	require.NoError(t, u.FromBinaryData(BinaryData{Type: "BINARY", Filename: "f", Id: uuid.Nil}))
	_, err = u.AsBinaryData()
	require.NoError(t, err)
	require.NoError(t, u.MergeBinaryData(BinaryData{Type: "BINARY", Filename: "f2", Id: uuid.Nil}))
}

func TestItemUpdateUnionOperations(t *testing.T) {
	var u ItemUpdate_Data
	require.NoError(t, u.FromCredentialData(CredentialData{Type: "CREDENTIAL", Login: "l", Password: "p"}))
	_, err := u.AsCredentialData()
	require.NoError(t, err)
	require.NoError(t, u.MergeCredentialData(CredentialData{Type: "CREDENTIAL", Login: "l2"}))
	require.NoError(t, u.FromCardData(CardData{Type: "CARD", CardNumber: "4111111111111111", CardHolder: "h", ExpiryDate: "12/25", Cvv: "123"}))
	_, err = u.AsCardData()
	require.NoError(t, err)
	require.NoError(t, u.MergeCardData(CardData{Type: "CARD", CardHolder: "h2"}))
	require.NoError(t, u.FromTextData(TextData{Type: "TEXT", Value: "v"}))
	_, err = u.AsTextData()
	require.NoError(t, err)
	require.NoError(t, u.MergeTextData(TextData{Type: "TEXT", Value: "v2"}))
	require.NoError(t, u.FromBinaryData(BinaryData{Type: "BINARY", Filename: "f", Id: uuid.Nil}))
	_, err = u.AsBinaryData()
	require.NoError(t, err)
	require.NoError(t, u.MergeBinaryData(BinaryData{Type: "BINARY", Filename: "f2", Id: uuid.Nil}))
}
