package affiliatelinks

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func newImageObjectID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", b[0])
	}
	return hex.EncodeToString(b[:])
}
