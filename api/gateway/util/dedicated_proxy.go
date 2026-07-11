package util

import (
	"fmt"
	"os"
	"strings"
)

type DedicatedProxy struct {
	Host     string
	Port     string
	Username string
	Password string
}

func parseDedicatedProxy(raw string) DedicatedProxy {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DedicatedProxy{}
	}

	parts := strings.Split(raw, "|")
	if len(parts) != 4 {
		return DedicatedProxy{}
	}

	return DedicatedProxy{
		Host:     strings.TrimSpace(parts[0]),
		Port:     strings.TrimSpace(parts[1]),
		Username: strings.TrimSpace(parts[2]),
		Password: strings.TrimSpace(parts[3]),
	}
}

func GetDedicatedProxy() []DedicatedProxy {
	return []DedicatedProxy{
		parseDedicatedProxy(os.Getenv("DEDICATED_PROXY_1")),
		parseDedicatedProxy(os.Getenv("DEDICATED_PROXY_2")),
		parseDedicatedProxy(os.Getenv("DEDICATED_PROXY_3")),
		parseDedicatedProxy(os.Getenv("DEDICATED_PROXY_4")),
		parseDedicatedProxy(os.Getenv("DEDICATED_PROXY_5")),
		parseDedicatedProxy(os.Getenv("DEDICATED_PROXY_6")),
		parseDedicatedProxy(os.Getenv("DEDICATED_PROXY_7")),
	}
}

func BuildDedicatedProxyURL(proxy DedicatedProxy) (string, bool) {
	if proxy.Host == "" || proxy.Port == "" {
		return "", false
	}

	if proxy.Username != "" || proxy.Password != "" {
		return fmt.Sprintf("http://%s:%s@%s:%s", proxy.Username, proxy.Password, proxy.Host, proxy.Port), true
	}

	return fmt.Sprintf("http://%s:%s", proxy.Host, proxy.Port), true
}

// BuildProxyURL supports either a fully-formed proxy URL or the pipe-separated
// host|port|username|password format used by DEDICATED_PROXY_* env vars.
func BuildProxyURL(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	if strings.Contains(raw, "://") {
		return raw, true
	}
	return BuildDedicatedProxyURL(parseDedicatedProxy(raw))
}

func GetDedicatedProxyURLs() []string {
	proxies := GetDedicatedProxy()
	proxyURLs := make([]string, 0, len(proxies))
	for _, proxy := range proxies {
		proxyURL, ok := BuildDedicatedProxyURL(proxy)
		if ok {
			proxyURLs = append(proxyURLs, proxyURL)
		}
	}
	return proxyURLs
}
