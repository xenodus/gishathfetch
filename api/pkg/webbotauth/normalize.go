package webbotauth

import (
	"encoding/base64"
	"regexp"
	"strings"
)

var singleLinePEMRE = regexp.MustCompile(`-----BEGIN ([^-]+)-----[ \t]*(.+?)[ \t]*-----END [^-]+-----`)

// NormalizePrivateKeyMaterial prepares secret key text copied from env vars or
// secret managers for PEM parsing. It handles common CI formats such as quoted
// values, escaped newlines, and single-line PEM blocks.
func NormalizePrivateKeyMaterial(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "\uFEFF")
	s = strings.Trim(s, `"'`)

	if strings.Contains(s, `\n`) || strings.Contains(s, `\r`) {
		s = strings.ReplaceAll(s, `\r\n`, "\n")
		s = strings.ReplaceAll(s, `\n`, "\n")
		s = strings.ReplaceAll(s, `\r`, "\n")
	}

	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	if !strings.Contains(s, "\n") {
		if match := singleLinePEMRE.FindStringSubmatch(s); match != nil {
			s = "-----BEGIN " + match[1] + "-----\n" +
				strings.TrimSpace(match[2]) + "\n" +
				"-----END " + match[1] + "-----"
		}
	}

	return strings.TrimSpace(s)
}

func decodeBase64KeyMaterial(raw string) ([]byte, error) {
	compact := strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '\t', ' ':
			return -1
		default:
			return r
		}
	}, raw)

	if compact == "" {
		return nil, base64.CorruptInputError(0)
	}

	if decoded, err := base64.StdEncoding.DecodeString(compact); err == nil {
		return decoded, nil
	}
	return base64.RawStdEncoding.DecodeString(compact)
}
