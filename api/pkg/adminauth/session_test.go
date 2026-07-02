package adminauth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIssueAndValidateSessionToken(t *testing.T) {
	token, err := IssueSessionToken("secret", "admin", time.Hour)
	require.NoError(t, err)

	username, err := ValidateSessionToken("secret", token)
	require.NoError(t, err)
	require.Equal(t, "admin", username)
}

func TestValidateSessionToken_RejectsTamperedToken(t *testing.T) {
	token, err := IssueSessionToken("secret", "admin", time.Hour)
	require.NoError(t, err)

	tampered := token[:len(token)-1] + "x"
	_, err = ValidateSessionToken("secret", tampered)
	require.ErrorIs(t, err, ErrInvalidSession)
}

func TestValidateSessionToken_RejectsExpiredToken(t *testing.T) {
	token, err := IssueSessionToken("secret", "admin", -time.Second)
	require.NoError(t, err)

	_, err = ValidateSessionToken("secret", token)
	require.ErrorIs(t, err, ErrExpiredSession)
}

func TestCredentialsMatch(t *testing.T) {
	require.True(t, CredentialsMatch("admin", "secret", "admin", "secret"))
	require.False(t, CredentialsMatch("admin", "secret", "admin", "wrong"))
	require.False(t, CredentialsMatch("admin", "secret", "wrong", "secret"))
}

func TestSessionTokenFromCookie(t *testing.T) {
	require.Equal(
		t,
		"abc.def",
		SessionTokenFromCookie("foo=bar; gishathfetch_admin_session=abc.def; baz=qux"),
	)
}
