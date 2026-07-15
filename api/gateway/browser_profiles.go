package gateway

import (
	"math/rand/v2"

	"github.com/bogdanfinn/tls-client/profiles"
)

// BrowserFamily identifies which browser TLS/HTTP headers to emulate.
type BrowserFamily string

const (
	BrowserFamilyChrome  BrowserFamily = "chrome"
	BrowserFamilyFirefox BrowserFamily = "firefox"
	BrowserFamilySafari  BrowserFamily = "safari"
	BrowserFamilyEdge    BrowserFamily = "edge"
)

// BrowserEmulationProfile pairs a User-Agent with a matching TLS client profile.
type BrowserEmulationProfile struct {
	Enabled    bool
	Family     BrowserFamily
	UserAgent  string
	TLSProfile profiles.ClientProfile
	Platform   string
}

var browserEmulationProfiles = []BrowserEmulationProfile{
	{
		Enabled:    true,
		Family:     BrowserFamilyChrome,
		UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
		TLSProfile: profiles.Chrome_133,
		Platform:   `"Windows"`,
	},
	{
		Enabled:    true,
		Family:     BrowserFamilyChrome,
		UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
		TLSProfile: profiles.Chrome_133,
		Platform:   `"macOS"`,
	},
	{
		Enabled:    true,
		Family:     BrowserFamilyChrome,
		UserAgent:  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
		TLSProfile: profiles.Chrome_133,
		Platform:   `"Linux"`,
	},
	{
		Enabled:    true,
		Family:     BrowserFamilyFirefox,
		UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:137.0) Gecko/20100101 Firefox/137.0",
		TLSProfile: profiles.Firefox_135,
		Platform:   `"Windows"`,
	},
	{
		Enabled:    true,
		Family:     BrowserFamilyFirefox,
		UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 14.4; rv:137.0) Gecko/20100101 Firefox/137.0",
		TLSProfile: profiles.Firefox_135,
		Platform:   `"macOS"`,
	},
	{
		Enabled:    true,
		Family:     BrowserFamilySafari,
		UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.4 Safari/605.1.15",
		TLSProfile: profiles.Safari_16_0,
		Platform:   `"macOS"`,
	},
	{
		Enabled:    true,
		Family:     BrowserFamilyEdge,
		UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edg/135.0.3179.98 Chrome/135.0.0.0 Safari/537.36",
		TLSProfile: profiles.Chrome_133,
		Platform:   `"Windows"`,
	},
}

// PickBrowserProfile selects a random browser profile with a TLS fingerprint that
// matches the chosen User-Agent family.
func PickBrowserProfile() BrowserEmulationProfile {
	if len(browserEmulationProfiles) == 0 {
		return BrowserEmulationProfile{}
	}
	return browserEmulationProfiles[rand.IntN(len(browserEmulationProfiles))]
}

// RandomBrowserUserAgent returns a User-Agent from the emulation profile list.
func RandomBrowserUserAgent() string {
	return PickBrowserProfile().UserAgent
}
