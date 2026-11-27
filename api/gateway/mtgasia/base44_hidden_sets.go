package mtgasia

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"
)

const hiddenSetCacheTTL = 15 * time.Minute

// resetHiddenSetCacheForTest clears the in-process hidden-set cache (tests only).
func resetHiddenSetCacheForTest() {
	hiddenSetCache.Lock()
	hiddenSetCache.ids = nil
	hiddenSetCache.expiry = time.Time{}
	hiddenSetCache.Unlock()
}

var hiddenSetCache struct {
	sync.RWMutex
	ids    map[string]struct{}
	expiry time.Time
}

func getHiddenSetIDs(ctx context.Context) (map[string]struct{}, error) {
	hiddenSetCache.RLock()
	if len(hiddenSetCache.ids) > 0 && time.Now().Before(hiddenSetCache.expiry) {
		cached := make(map[string]struct{}, len(hiddenSetCache.ids))
		for id := range hiddenSetCache.ids {
			cached[id] = struct{}{}
		}
		hiddenSetCache.RUnlock()
		return cached, nil
	}
	hiddenSetCache.RUnlock()

	ids, err := fetchHiddenSetIDs(ctx)
	if err != nil {
		return nil, err
	}

	hiddenSetCache.Lock()
	hiddenSetCache.ids = ids
	hiddenSetCache.expiry = time.Now().Add(hiddenSetCacheTTL)
	hiddenSetCache.Unlock()

	copied := make(map[string]struct{}, len(ids))
	for id := range ids {
		copied[id] = struct{}{}
	}
	return copied, nil
}

func fetchHiddenSetIDs(ctx context.Context) (map[string]struct{}, error) {
	filterJSON, err := json.Marshal(map[string]any{
		"hidden": true,
	})
	if err != nil {
		return nil, err
	}

	apiURL := &url.URL{
		Scheme: "https",
		Host:   strings.TrimPrefix(base44ServerURL, "https://"),
		Path:   "/api/apps/" + base44AppID + "/entities/CardSet",
		RawQuery: url.Values{
			"q":     {string(filterJSON)},
			"limit": {"500"},
		}.Encode(),
	}

	resp, err := gateway.DoOutboundGET(ctx, apiURL.String(), base44OutboundOpts(), config.DedicatedSearchAttemptTimeout)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s", gateway.FormatUnexpectedHTTPStatus(StoreName, resp, body))
	}

	var sets []cardSet
	if err := json.Unmarshal(body, &sets); err != nil {
		return nil, gateway.WrapJSONDecodeError(err, resp, body)
	}

	ids := make(map[string]struct{}, len(sets))
	for _, set := range sets {
		if id := strings.TrimSpace(set.ID); id != "" {
			ids[id] = struct{}{}
		}
	}
	return ids, nil
}

func inventoryInHiddenSet(setID string, hidden map[string]struct{}) bool {
	id := strings.TrimSpace(setID)
	if id == "" {
		return false
	}
	_, found := hidden[id]
	return found
}