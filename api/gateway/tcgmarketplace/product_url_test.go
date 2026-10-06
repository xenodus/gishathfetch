package tcgmarketplace

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test-only private key: published in the storefront bundle for client-side decode helpers.
const productURLPrivateKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgGlempQY/LwZbvzeYl76yMaH/onD/olkEmMC5rbms3BSAA/TbzPM
EVjjXcKjFHcBlKC5KOAyqNF5z7VZc6hyM6GL8l4obNBp6LWUmeZUWFm7rsLNXIm+
Sv7IOw2z/1frbyKgWagqRstIkEnmqqsgDrLJc9OSt5FfOO99tterVzVlAgMBAAEC
gYAKYT9pB20eOoMsddvK73mH1S3F9IDGmA0Xo9mGewOCNRG8fV+fAqNS1lMOMXJ6
prU1gAf+zf7DY/SKMN2r4lQjnMl4GhGmKUcZnulLyi1rXFEbJdktBnjH02KozaNt
pARxKPCfmJlJ/252MbK7WZBb7fpbCC3M36cuBvujxirooQJBAMq9KZBmiYCpD7fV
ajA3gr9C0/DDUd4hiwSaRe07jntXknK98a8mA9yxi9JFo0bKxWeUuTilTCLVmbSM
SUx7Ov8CQQCFDQqUdzMgjeMqXy7q+t7ZCFiJ8uD4W6QiYStff8VV/z+0/gf1iqEC
N1ecIUL/0dNvlIPjFEpplyJ35AXkdoObAkEAiv9N89aZB0DhqeQDvbMRTRgAPt1q
z6SnrywmLRPcB+xuV4Sqoq6pjzGa7zsXSn3TAcURiCisHmXbz9Eun+UM+QJAbPJg
13FDzERi3y9Rm9gQ4maTGWJoFPX1ULGAcpKNY/2+jNX2bnk1njry4kXaNt/54kQa
OrdCt3OqJcHxkUH6QQJARa37/rqAMsNFPMppQ+BDQorQSy0Pu3PYmE9Jqh7WVsAi
FMdiPGMl2MkaD0kEF2BCzh33jTnejFJD897xgKbEXg==
-----END RSA PRIVATE KEY-----`

func decryptProductID(encoded string) (string, error) {
	block, _ := pem.Decode([]byte(productURLPrivateKeyPEM))
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(encoded, "_", "/"))
	if err != nil {
		return "", err
	}
	plain, err := rsa.DecryptPKCS1v15(rand.Reader, key, ciphertext)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func TestEncodeProductID_roundTrip(t *testing.T) {
	t.Setenv("GODEBUG", "rsa1024min=0")

	encoded, err := encodeProductID(285511)
	require.NoError(t, err)
	require.NotEmpty(t, encoded)

	decoded, err := decryptProductID(encoded)
	require.NoError(t, err)
	require.Equal(t, "285511", decoded)
}

func TestProductPageURLFromID(t *testing.T) {
	t.Setenv("GODEBUG", "rsa1024min=0")

	u, err := productPageURLFromID(285511)
	require.NoError(t, err)
	require.Contains(t, u.String(), StoreBaseURL+"/product/B/")
	require.Contains(t, u.String(), "/0")
	require.Contains(t, u.RawQuery, "utm_source=")
}

func TestResolveProductPageURL_prefersExplicitURL(t *testing.T) {
	raw := "https://thetcgmarketplace.cc/product/B/example/0"
	got, err := resolveProductPageURL(listing{URL: raw})
	require.NoError(t, err)
	require.Contains(t, got, "thetcgmarketplace.com/product/B/example/0")
}

func TestResolveProductPageURL_fromListingID(t *testing.T) {
	t.Setenv("GODEBUG", "rsa1024min=0")

	got, err := resolveProductPageURL(listing{ID: 285511})
	require.NoError(t, err)
	require.Contains(t, got, StoreBaseURL+"/product/B/")
}
