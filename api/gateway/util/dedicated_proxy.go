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

	parts := strings.SplitN(raw, "|", 4)

	// Be defensive: if there are fewer than 4 segments, fill what we can.
	proxy := DedicatedProxy{}
	if len(parts) > 0 {
		proxy.Host = strings.TrimSpace(parts[0])
	}
	if len(parts) > 1 {
		proxy.Port = strings.TrimSpace(parts[1])
	}
	if len(parts) > 2 {
		proxy.Username = strings.TrimSpace(parts[2])
	}
	if len(parts) > 3 {
		proxy.Password = strings.TrimSpace(parts[3])
	}
	return proxy
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
