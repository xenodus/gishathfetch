package tcgmarketplace

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"mtg-price-checker-sg/pkg/config"
)

// productURLPublicKeyPEM matches the storefront RSA public key used to encode IDs in
// /product/B/{encoded_id}/0 links (see thetcgmarketplace.com static bundle).
const productURLPublicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIGeMA0GCSqGSIb3DQEBAQUAA4GMADCBiAKBgGlempQY/LwZbvzeYl76yMaH/onD
/olkEmMC5rbms3BSAA/TbzPMEVjjXcKjFHcBlKC5KOAyqNF5z7VZc6hyM6GL8l4o
bNBp6LWUmeZUWFm7rsLNXIm+Sv7IOw2z/1frbyKgWagqRstIkEnmqqsgDrLJc9OS
t5FfOO99tterVzVlAgMBAAE=
-----END PUBLIC KEY-----`

var (
	productURLPublicKeyOnce sync.Once
	productURLPublicKey     *rsa.PublicKey
	productURLPublicKeyErr  error
	weakRSA1024Once         sync.Once
)

func allowWeakRSA1024Keys() {
	weakRSA1024Once.Do(func() {
		current := os.Getenv("GODEBUG")
		if strings.Contains(current, "rsa1024min=0") {
			return
		}
		if current == "" {
			os.Setenv("GODEBUG", "rsa1024min=0")
			return
		}
		os.Setenv("GODEBUG", current+",rsa1024min=0")
	})
}

func productURLPublicKeyFromPEM() (*rsa.PublicKey, error) {
	productURLPublicKeyOnce.Do(func() {
		block, _ := pem.Decode([]byte(productURLPublicKeyPEM))
		if block == nil {
			productURLPublicKeyErr = fmt.Errorf("%s: invalid product url public key", StoreName)
			return
		}
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			productURLPublicKeyErr = err
			return
		}
		key, ok := pub.(*rsa.PublicKey)
		if !ok {
			productURLPublicKeyErr = fmt.Errorf("%s: product url key is not RSA", StoreName)
			return
		}
		productURLPublicKey = key
	})
	return productURLPublicKey, productURLPublicKeyErr
}

func encodeProductID(id int64) (string, error) {
	if id <= 0 {
		return "", fmt.Errorf("%s: invalid product id %d", StoreName, id)
	}
	pub, err := productURLPublicKeyFromPEM()
	if err != nil {
		return "", err
	}
	allowWeakRSA1024Keys()
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(strconv.FormatInt(id, 10)))
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(base64.StdEncoding.EncodeToString(encrypted), "/", "_"), nil
}

func productPageURLFromID(id int64) (*url.URL, error) {
	encoded, err := encodeProductID(id)
	if err != nil {
		return nil, err
	}
	pageURL, err := url.Parse(StoreBaseURL + "/product/B/" + encoded + "/0")
	if err != nil {
		return nil, err
	}
	pageURL.RawQuery = url.Values{
		"utm_source": []string{config.UtmSource},
	}.Encode()
	return pageURL, nil
}
