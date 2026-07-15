package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/cardaffinity"
	"mtg-price-checker-sg/gateway/cardscentral"
	"mtg-price-checker-sg/gateway/cardscitadel"
	"mtg-price-checker-sg/gateway/hideout"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestInitAndMapShops_FiltersByRequestedLGS(t *testing.T) {
	shops := initAndMapShops([]string{cardscentral.StoreName, hideout.StoreName})

	if len(shops) != 2 {
		t.Fatalf("expected 2 shops after filtering, got %d", len(shops))
	}
	if _, ok := shops[cardscentral.StoreName]; !ok {
		t.Fatalf("expected %q to be included", cardscentral.StoreName)
	}
	if _, ok := shops[hideout.StoreName]; !ok {
		t.Fatalf("expected %q to be included", hideout.StoreName)
	}
	if _, ok := shops[cardaffinity.StoreName]; ok {
		t.Fatalf("did not expect %q to be included", cardaffinity.StoreName)
	}
}

// MockLGS is a mock implementation of gateway.LGS
type MockLGS struct {
	SearchFunc func(ctx context.Context, searchStr string) ([]gateway.Card, error)
}

func (m *MockLGS) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, searchStr)
	}
	return nil, nil
}

func TestIsJapanese(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"Contains Japanese":             {"This is a Japanese card", true},
		"Contains japanese (lowercase)": {"this is a japanese card", true},
		"Does not contain Japanese":     {"This is an English card", false},
		"Empty string":                  {"", false},
		"Mixed case":                    {"JaPaNeSe", true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := isJapanese(tt.input); got != tt.expected {
				t.Errorf("isJapanese(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCleanName(t *testing.T) {
	tests := map[string]struct {
		name              string
		quality           string
		expectedName      string
		expectedExtraInfo []string
	}{
		"Name [Tag]": {
			name:              "Name [Tag]",
			quality:           "",
			expectedName:      "Name",
			expectedExtraInfo: []string{"[Tag]"},
		},
		"Name[Tag]": {
			name:              "Name[Tag]",
			quality:           "",
			expectedName:      "Name",
			expectedExtraInfo: []string{"[Tag]"},
		},
		"Name (Tag)": {
			name:              "Name (Tag)",
			quality:           "",
			expectedName:      "Name",
			expectedExtraInfo: []string{"(Tag)"},
		},
		"Name [Tag1] (Tag2)": {
			name:              "Name [Tag1] (Tag2)",
			quality:           "",
			expectedName:      "Name",
			expectedExtraInfo: []string{"[Tag1] (Tag2)"},
		},
		"Name (Tag1) [Tag2]": {
			name:              "Name (Tag1) [Tag2]",
			quality:           "",
			expectedName:      "Name",
			expectedExtraInfo: []string{"[Tag2]", "(Tag1)"},
		},
		"Name - Quality": {
			name:              "Name - Quality",
			quality:           "Quality",
			expectedName:      "Name",
			expectedExtraInfo: nil,
		},
		"Empty bracket extra info": {
			name:              "Kratos, God of War",
			quality:           "",
			expectedName:      "Kratos, God of War",
			expectedExtraInfo: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var initialExtra []string
			if name == "Empty bracket extra info" {
				initialExtra = []string{"[]"}
			}
			gotName, gotExtra := cleanName(tt.name, tt.quality, initialExtra)
			if gotName != tt.expectedName {
				t.Errorf("cleanName() gotName = %v, want %v", gotName, tt.expectedName)
			}
			if len(gotExtra) != len(tt.expectedExtraInfo) {
				t.Errorf("cleanName() gotExtra len = %v, want %v", len(gotExtra), len(tt.expectedExtraInfo))
			} else {
				for i := range gotExtra {
					if gotExtra[i] != tt.expectedExtraInfo[i] {
						t.Errorf("cleanName() gotExtra[%d] = %v, want %v", i, gotExtra[i], tt.expectedExtraInfo[i])
					}
				}
			}
		})
	}
}

func TestIsArtCard(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"Art Card":     {"Some Name Art Card", true},
		"Art Series":   {"Some Name Art Series", true},
		"Normal Card":  {"Some Name", false},
		"Empty String": {"", false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := isArtCard(tt.input); got != tt.expected {
				t.Errorf("isArtCard(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSearchShops(t *testing.T) {
	tests := map[string]struct {
		input         SearchInput
		lgsResponses  map[string][]gateway.Card
		expectedCount int
		verifyFunc    func(t *testing.T, cards []Card)
	}{
		"Basic Search - One Shop, One Result": {
			input: SearchInput{SearchString: "Card A"},
			lgsResponses: map[string][]gateway.Card{
				"Shop1": {
					{Name: "Card A", Price: 10.0, InStock: true, Source: "Shop1"},
				},
			},
			expectedCount: 1,
			verifyFunc: func(t *testing.T, cards []Card) {
				if cards[0].Name != "Card A" {
					t.Errorf("Expected card name 'Card A', got '%s'", cards[0].Name)
				}
			},
		},
		"Filtering - Skip Out of Stock": {
			input: SearchInput{SearchString: "Card A"},
			lgsResponses: map[string][]gateway.Card{
				"Shop1": {
					{Name: "Card A", Price: 10.0, InStock: false, Source: "Shop1"},
				},
			},
			expectedCount: 0,
			verifyFunc:    nil,
		},
		"Filtering - Skip Price Zero": {
			input: SearchInput{SearchString: "Card A"},
			lgsResponses: map[string][]gateway.Card{
				"Shop1": {
					{Name: "Card A", Price: 0, InStock: true, Source: "Shop1"},
				},
			},
			expectedCount: 0,
			verifyFunc:    nil,
		},
		"Filtering - Skip Art Card": {
			input: SearchInput{SearchString: "Card A"},
			lgsResponses: map[string][]gateway.Card{
				"Shop1": {
					{Name: "Card A Art Card", Price: 10.0, InStock: true, Source: "Shop1"},
				},
			},
			expectedCount: 0,
			verifyFunc:    nil,
		},
		"Filtering - Skip Japanese Card": {
			input: SearchInput{SearchString: "Card A"},
			lgsResponses: map[string][]gateway.Card{
				"Shop1": {
					{Name: "Card A (Japanese)", Price: 10.0, InStock: true, Source: "Shop1"},
				},
			},
			expectedCount: 0,
			verifyFunc:    nil,
		},
		"Sorting - Price Ascending": {
			input: SearchInput{SearchString: "Card A"},
			lgsResponses: map[string][]gateway.Card{
				"Shop1": {
					{Name: "Card A", Price: 20.0, InStock: true, Source: "Shop1"},
				},
				"Shop2": {
					{Name: "Card A", Price: 10.0, InStock: true, Source: "Shop2"},
				},
			},
			expectedCount: 2,
			verifyFunc: func(t *testing.T, cards []Card) {
				if cards[0].Price != 10.0 {
					t.Errorf("Expected first card price 10.0, got %f", cards[0].Price)
				}
				if cards[1].Price != 20.0 {
					t.Errorf("Expected second card price 20.0, got %f", cards[1].Price)
				}
			},
		},
		"Match Priority - Exact > Prefix > Partial": {
			input: SearchInput{SearchString: "Jace"},
			lgsResponses: map[string][]gateway.Card{
				"Shop1": {
					{Name: "Jace, the Mind Sculptor", Price: 50.0, InStock: true, Source: "Shop1"}, // Prefix
					{Name: "Jace", Price: 50.0, InStock: true, Source: "Shop1"},                    // Exact
					{Name: "Agent of Jace", Price: 50.0, InStock: true, Source: "Shop1"},           // Partial
				},
			},
			expectedCount: 3,
			verifyFunc: func(t *testing.T, cards []Card) {
				if cards[0].Name != "Jace" {
					t.Errorf("Expected first card to be Exact Match 'Jace', got '%s'", cards[0].Name)
				}
				if cards[1].Name != "Jace, the Mind Sculptor" {
					t.Errorf("Expected second card to be Prefix Match 'Jace, the Mind Sculptor', got '%s'", cards[1].Name)
				}
				if cards[2].Name != "Agent of Jace" {
					t.Errorf("Expected third card to be Partial Match 'Agent of Jace', got '%s'", cards[2].Name)
				}
			},
		},

		"Extra Info Parsing": {
			input: SearchInput{SearchString: "Card A"},
			lgsResponses: map[string][]gateway.Card{
				"Shop1": {
					{Name: "Card A [Foil]", Price: 10.0, InStock: true, Source: "Shop1"},
				},
			},
			expectedCount: 1,
			verifyFunc: func(t *testing.T, cards []Card) {
				if cards[0].Name != "Card A" {
					t.Errorf("Expected clean name 'Card A', got '%s'", cards[0].Name)
				}
				if cards[0].ExtraInfo != "[Foil]" {
					t.Errorf("Expected ExtraInfo '[Foil]', got '%s'", cards[0].ExtraInfo)
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockMap := make(map[string]gateway.LGS)
			for shopName, storedCards := range tt.lgsResponses {
				// Avoid closure capture issues
				cardsToReturn := storedCards
				mockMap[shopName] = &MockLGS{
					SearchFunc: func(ctx context.Context, searchStr string) ([]gateway.Card, error) {
						if cardsToReturn == nil {
							return nil, fmt.Errorf("simulated error")
						}
						return cardsToReturn, nil
					},
				}
			}

			results, storeErrors, err := searchShops(context.Background(), tt.input, mockMap)
			if err != nil {
				t.Fatalf("searchShops returned unexpected error: %v", err)
			}
			if len(storeErrors) != 0 {
				t.Fatalf("expected no store errors, got %d", len(storeErrors))
			}

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}

			if tt.verifyFunc != nil && len(results) > 0 {
				tt.verifyFunc(t, results)
			}
		})
	}
}

func TestSearchShops_IncludesStoreErrors(t *testing.T) {
	shops := map[string]gateway.LGS{
		"Failing Shop": &MockLGS{
			SearchFunc: func(ctx context.Context, searchStr string) ([]gateway.Card, error) {
				return nil, fmt.Errorf("simulated failure")
			},
		},
		"Good Shop": &MockLGS{
			SearchFunc: func(ctx context.Context, searchStr string) ([]gateway.Card, error) {
				return []gateway.Card{
					{Name: searchStr, Price: 1.0, InStock: true, Source: "Good Shop"},
				}, nil
			},
		},
	}

	results, storeErrors, err := searchShops(context.Background(), SearchInput{SearchString: "Card A"}, shops)
	if err != nil {
		t.Fatalf("searchShops returned unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result card, got %d", len(results))
	}
	if len(storeErrors) != 1 {
		t.Fatalf("expected 1 store error, got %d", len(storeErrors))
	}
	if storeErrors[0].Store != "Failing Shop" {
		t.Fatalf("expected store error for 'Failing Shop', got %q", storeErrors[0].Store)
	}
	if storeErrors[0].Error != "simulated failure" {
		t.Fatalf("expected 'simulated failure', got %q", storeErrors[0].Error)
	}
}

func TestBuildStoreErrors_IncludesHTTPStatusCode(t *testing.T) {
	storeErrors := buildStoreErrors(map[string]error{
		"Hideout": fmt.Errorf(
			"attempt 2 (scrap-direct): Service Unavailable (proxy_mode=direct proxy=none)",
		),
		"Cards Central": fmt.Errorf("unexpected status for Cards Central: 429 Too Many Requests"),
	})

	if len(storeErrors) != 2 {
		t.Fatalf("expected 2 store errors, got %d", len(storeErrors))
	}

	byStore := make(map[string]StoreError, len(storeErrors))
	for _, storeError := range storeErrors {
		byStore[storeError.Store] = storeError
	}

	hideout := byStore["Hideout"]
	if hideout.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 for Hideout, got %d", hideout.StatusCode)
	}
	if !strings.Contains(hideout.Error, "503 Service Unavailable") {
		t.Fatalf("expected enriched error message, got %q", hideout.Error)
	}

	cardsCentral := byStore["Cards Central"]
	if cardsCentral.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected status 429 for Cards Central, got %d", cardsCentral.StatusCode)
	}
}

func TestIsBinderposStore(t *testing.T) {
	tests := map[string]bool{
		cardscitadel.StoreName: true,
		hideout.StoreName:      true,
		cardscentral.StoreName: false,
		"Unknown Shop":         false,
	}

	for shop, expected := range tests {
		t.Run(shop, func(t *testing.T) {
			if got := isBinderposStore(shop); got != expected {
				t.Fatalf("isBinderposStore(%q) = %v, want %v", shop, got, expected)
			}
		})
	}
}

func TestFetchCardsConcurrently_ConcurrentStoresGetDistinctDedicatedProxies(t *testing.T) {
	for i := 1; i <= 7; i++ {
		t.Setenv(fmt.Sprintf("DEDICATED_PROXY_%d", i), "")
	}
	for i := 1; i <= 3; i++ {
		t.Setenv(fmt.Sprintf("DEDICATED_PROXY_%d", i), fmt.Sprintf("10.0.0.%d|8080|user|pass", i))
	}

	started := make(chan struct{}, gateway.DedicatedProxySearchMaxConcurrent)
	release := make(chan struct{})
	var mu sync.Mutex
	pinned := make(map[string]string, gateway.DedicatedProxySearchMaxConcurrent)

	makeShop := func(name string) gateway.LGS {
		return &MockLGS{
			SearchFunc: func(ctx context.Context, searchStr string) ([]gateway.Card, error) {
				proxyURL, ok := gateway.RequestDedicatedProxyURL(ctx)
				if !ok || proxyURL == "" {
					t.Errorf("shop %s: expected pinned dedicated proxy", name)
				}
				mu.Lock()
				pinned[name] = proxyURL
				mu.Unlock()
				started <- struct{}{}
				<-release
				return nil, nil
			},
		}
	}

	shops := map[string]gateway.LGS{
		"Shop1": makeShop("Shop1"),
		"Shop2": makeShop("Shop2"),
		"Shop3": makeShop("Shop3"),
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, siteErrors := fetchCardsConcurrently(context.Background(), "Abrade", shops)
		if len(siteErrors) != 0 {
			t.Errorf("unexpected site errors: %v", siteErrors)
		}
	}()

	for range gateway.DedicatedProxySearchMaxConcurrent {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for stores to start")
		}
	}
	close(release)
	<-done

	mu.Lock()
	defer mu.Unlock()
	if len(pinned) != gateway.DedicatedProxySearchMaxConcurrent {
		t.Fatalf("expected %d pinned proxies, got %d (%v)", gateway.DedicatedProxySearchMaxConcurrent, len(pinned), pinned)
	}
	seen := make(map[string]struct{}, gateway.DedicatedProxySearchMaxConcurrent)
	for shop, proxyURL := range pinned {
		if proxyURL == "" {
			t.Fatalf("shop %s pinned empty proxy", shop)
		}
		if _, dup := seen[proxyURL]; dup {
			t.Fatalf("expected distinct proxies for concurrent stores, got %v", pinned)
		}
		seen[proxyURL] = struct{}{}
	}
}

func TestFetchCardsConcurrently_ConcurrentSearchesGetDistinctDedicatedProxies(t *testing.T) {
	for i := 1; i <= 7; i++ {
		t.Setenv(fmt.Sprintf("DEDICATED_PROXY_%d", i), "")
	}
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")
	t.Setenv("DEDICATED_PROXY_2", "5.6.7.8|8080|user|pass")
	t.Setenv("DEDICATED_PROXY_3", "9.10.11.12|8080|user|pass")

	started := make(chan struct{}, 3)
	release := make(chan struct{})
	var mu sync.Mutex
	pinnedBySearch := make(map[string]string, 3)

	shops := map[string]gateway.LGS{
		"ShopA": &MockLGS{
			SearchFunc: func(ctx context.Context, searchStr string) ([]gateway.Card, error) {
				if proxyURL, ok := gateway.RequestDedicatedProxyURL(ctx); ok {
					mu.Lock()
					pinnedBySearch[searchStr] = proxyURL
					mu.Unlock()
				}
				started <- struct{}{}
				<-release
				return nil, nil
			},
		},
	}

	done := make(chan struct{}, 3)
	for _, searchStr := range []string{"Search1", "Search2", "Search3"} {
		go func(searchStr string) {
			defer func() { done <- struct{}{} }()
			_, siteErrors := fetchCardsConcurrently(context.Background(), searchStr, shops)
			if len(siteErrors) != 0 {
				t.Errorf("search %q: unexpected site errors: %v", searchStr, siteErrors)
			}
		}(searchStr)
	}

	for range 3 {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for searches to start")
		}
	}
	close(release)

	for range 3 {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for searches to finish")
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(pinnedBySearch) != 3 {
		t.Fatalf("expected 3 pinned proxy URLs, got %d (%v)", len(pinnedBySearch), pinnedBySearch)
	}
	seen := make(map[string]struct{}, 3)
	for searchStr, proxyURL := range pinnedBySearch {
		if proxyURL == "" {
			t.Fatalf("search %q pinned empty proxy URL", searchStr)
		}
		if _, dup := seen[proxyURL]; dup {
			t.Fatalf("expected each concurrent search to pin a different dedicated proxy, got %v", pinnedBySearch)
		}
		seen[proxyURL] = struct{}{}
	}
}

func TestFetchCardsConcurrently_CollatesDiscordErrors(t *testing.T) {
	shops := map[string]gateway.LGS{
		"ErrorShopA": &MockLGS{
			SearchFunc: func(ctx context.Context, searchStr string) ([]gateway.Card, error) {
				return nil, fmt.Errorf("shop A failed")
			},
		},
		"ErrorShopB": &MockLGS{
			SearchFunc: func(ctx context.Context, searchStr string) ([]gateway.Card, error) {
				return nil, fmt.Errorf("shop B failed")
			},
		},
		"PanicShop": &MockLGS{
			SearchFunc: func(ctx context.Context, searchStr string) ([]gateway.Card, error) {
				panic("shop panic")
			},
		},
	}

	var mu sync.Mutex
	alertMessages := make([]string, 0, 1)
	alertDone := make(chan struct{}, 1)

	originalSendDiscordAlert := sendDiscordAlert
	sendDiscordAlert = func(message string) {
		mu.Lock()
		alertMessages = append(alertMessages, message)
		mu.Unlock()
		select {
		case alertDone <- struct{}{}:
		default:
		}
	}
	t.Cleanup(func() {
		sendDiscordAlert = originalSendDiscordAlert
	})

	_, siteErrors := fetchCardsConcurrently(context.Background(), "Abrade", shops)
	if len(siteErrors) != 3 {
		t.Fatalf("expected 3 site errors, got %d", len(siteErrors))
	}

	select {
	case <-alertDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for collated discord alert")
	}

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(alertMessages) != 1 {
		t.Fatalf("expected exactly 1 collated alert, got %d", len(alertMessages))
	}

	got := alertMessages[0]
	if !strings.Contains(got, "Encountered 3 error(s) while searching [Abrade]:") {
		t.Fatalf("expected collated summary header, got: %s", got)
	}
	if !strings.Contains(got, "- [ErrorShopA] shop A failed") {
		t.Fatalf("expected ErrorShopA details in alert, got: %s", got)
	}
	if !strings.Contains(got, "- [ErrorShopB] shop B failed") {
		t.Fatalf("expected ErrorShopB details in alert, got: %s", got)
	}
	if !strings.Contains(got, "- Recovered from panic in shop [PanicShop]: shop panic") {
		t.Fatalf("expected PanicShop details in alert, got: %s", got)
	}
}

func TestFetchCardsConcurrently_ReportsPerSiteTimeoutToDiscord(t *testing.T) {
	shops := map[string]gateway.LGS{
		"Timeout Shop": &MockLGS{
			SearchFunc: func(ctx context.Context, searchStr string) ([]gateway.Card, error) {
				return nil, context.DeadlineExceeded
			},
		},
	}

	var mu sync.Mutex
	alertMessages := make([]string, 0, 1)
	alertDone := make(chan struct{}, 1)

	originalSendDiscordAlert := sendDiscordAlert
	sendDiscordAlert = func(message string) {
		mu.Lock()
		alertMessages = append(alertMessages, message)
		mu.Unlock()
		select {
		case alertDone <- struct{}{}:
		default:
		}
	}
	t.Cleanup(func() {
		sendDiscordAlert = originalSendDiscordAlert
	})

	_, siteErrors := fetchCardsConcurrently(context.Background(), "Abrade", shops)
	if len(siteErrors) != 1 {
		t.Fatalf("expected 1 site error, got %d", len(siteErrors))
	}
	if !errors.Is(siteErrors["Timeout Shop"], context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded site error, got %v", siteErrors["Timeout Shop"])
	}

	select {
	case <-alertDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for timeout discord alert")
	}

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(alertMessages) != 1 {
		t.Fatalf("expected exactly 1 alert, got %d", len(alertMessages))
	}
	if !strings.Contains(alertMessages[0], "- [Timeout Shop] context deadline exceeded") {
		t.Fatalf("expected timeout shop in alert, got: %s", alertMessages[0])
	}
}

func TestFetchCardsConcurrently_SkipsCanceledForDiscord(t *testing.T) {
	shops := map[string]gateway.LGS{
		"Canceled Shop": &MockLGS{
			SearchFunc: func(ctx context.Context, searchStr string) ([]gateway.Card, error) {
				return nil, context.Canceled
			},
		},
	}

	alertSent := make(chan struct{}, 1)
	originalSendDiscordAlert := sendDiscordAlert
	sendDiscordAlert = func(message string) {
		select {
		case alertSent <- struct{}{}:
		default:
		}
	}
	t.Cleanup(func() {
		sendDiscordAlert = originalSendDiscordAlert
	})

	_, siteErrors := fetchCardsConcurrently(context.Background(), "Abrade", shops)
	if len(siteErrors) != 1 {
		t.Fatalf("expected 1 site error, got %d", len(siteErrors))
	}
	if !errors.Is(siteErrors["Canceled Shop"], context.Canceled) {
		t.Fatalf("expected canceled site error, got %v", siteErrors["Canceled Shop"])
	}

	select {
	case <-alertSent:
		t.Fatal("did not expect discord alert for canceled search")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestFetchCardsConcurrently_LimitsConcurrentWorkers(t *testing.T) {
	for i := 1; i <= 7; i++ {
		t.Setenv(fmt.Sprintf("DEDICATED_PROXY_%d", i), "")
	}

	const shopCount = 10
	started := make(chan struct{}, shopCount)
	release := make(chan struct{})
	var inFlight atomic.Int32
	var maxInFlight atomic.Int32

	shops := make(map[string]gateway.LGS, shopCount)
	for i := range shopCount {
		shops[fmt.Sprintf("Shop%d", i)] = &MockLGS{
			SearchFunc: func(ctx context.Context, searchStr string) ([]gateway.Card, error) {
				cur := inFlight.Add(1)
				for {
					prev := maxInFlight.Load()
					if cur <= prev || maxInFlight.CompareAndSwap(prev, cur) {
						break
					}
				}
				started <- struct{}{}
				<-release
				inFlight.Add(-1)
				return nil, nil
			},
		}
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, siteErrors := fetchCardsConcurrently(context.Background(), "Abrade", shops)
		if len(siteErrors) != 0 {
			t.Errorf("unexpected site errors: %v", siteErrors)
		}
	}()

	for i := range 6 {
		select {
		case <-started:
		case <-done:
			t.Fatal("fetch finished before 6 shops started")
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for shop %d to start; max in flight was %d", i, maxInFlight.Load())
		}
	}

	time.Sleep(50 * time.Millisecond)
	if got := maxInFlight.Load(); got > maxConcurrentStoreSearches {
		t.Fatalf("expected at most %d concurrent searches, got %d", maxConcurrentStoreSearches, got)
	}

	close(release)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for fetch to complete")
	}
}

func TestFormatShopSearchSummary(t *testing.T) {
	got := formatShopSearchSummary("Orthion, Hero of Lavabrink", 8*time.Second+240*time.Millisecond, []shopSearchDuration{
		{name: "5 Mana", duration: 8*time.Second + 240*time.Millisecond},
		{name: "Fyendal Hobby", duration: 222 * time.Millisecond},
		{name: "Cards & Collections", duration: 341 * time.Millisecond},
	})

	if !strings.Contains(got, "Checked 3 shops for [Orthion, Hero of Lavabrink] in 8.24s:") {
		t.Fatalf("expected summary header, got: %s", got)
	}
	if !strings.Contains(got, "[5 Mana] 8.24s") {
		t.Fatalf("expected 5 Mana duration, got: %s", got)
	}
	if !strings.Contains(got, "[Cards & Collections] 341ms") {
		t.Fatalf("expected Cards & Collections duration, got: %s", got)
	}
	if !strings.Contains(got, "[Fyendal Hobby] 222ms") {
		t.Fatalf("expected Fyendal Hobby duration, got: %s", got)
	}
	if strings.Index(got, "[5 Mana]") > strings.Index(got, "[Cards & Collections]") {
		t.Fatalf("expected shops to be sorted alphabetically, got: %s", got)
	}
}

func TestFormatDiscordErrorSummary(t *testing.T) {
	got := formatDiscordErrorSummary("Uro, Titan of Nature's Wrath", []string{
		"Error encountered searching [Tefuda] for [Uro, Titan of Nature's Wrath]: attempt 3 (scrap-direct): 503 Service Unavailable (proxy_mode=direct proxy=none)",
		"Error encountered searching [Hideout] for [Uro, Titan of Nature's Wrath]: attempt 2 (scrap-direct): 503 Service Unavailable (proxy_mode=direct proxy=none)",
		"Recovered from panic in shop [ShopPanic]: panic value",
	})

	if !strings.Contains(got, "Encountered 3 error(s) while searching [Uro, Titan of Nature's Wrath]:") {
		t.Fatalf("expected summary header, got: %s", got)
	}
	if !strings.Contains(got, "- [Hideout] attempt 2 (scrap-direct): 503 Service Unavailable (proxy_mode=direct proxy=none)") {
		t.Fatalf("expected Hideout concise line, got: %s", got)
	}
	if !strings.Contains(got, "- [Tefuda] attempt 3 (scrap-direct): 503 Service Unavailable (proxy_mode=direct proxy=none)") {
		t.Fatalf("expected Tefuda concise line, got: %s", got)
	}
	if !strings.Contains(got, "- Recovered from panic in shop [ShopPanic]: panic value") {
		t.Fatalf("expected fallback line for non-search error, got: %s", got)
	}
}
