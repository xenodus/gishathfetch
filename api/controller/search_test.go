package controller

import (
	"context"
	"fmt"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/agora"
	"mtg-price-checker-sg/gateway/cardaffinity"
	"mtg-price-checker-sg/gateway/cardscitadel"
	"mtg-price-checker-sg/gateway/gameshaven"
	"mtg-price-checker-sg/gateway/gog"
	"mtg-price-checker-sg/gateway/hideout"
	"mtg-price-checker-sg/gateway/tefuda"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestInitAndMapShops_FiltersByRequestedLGS(t *testing.T) {
	shops := initAndMapShops([]string{agora.StoreName, tefuda.StoreName})

	if len(shops) != 2 {
		t.Fatalf("expected 2 shops after filtering, got %d", len(shops))
	}
	if _, ok := shops[agora.StoreName]; !ok {
		t.Fatalf("expected %q to be included", agora.StoreName)
	}
	if _, ok := shops[tefuda.StoreName]; !ok {
		t.Fatalf("expected %q to be included", tefuda.StoreName)
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
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotName, gotExtra := cleanName(tt.name, tt.quality, nil)
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

func TestIsBinderposStore(t *testing.T) {
	tests := map[string]bool{
		cardscitadel.StoreName: true,
		tefuda.StoreName:       true,
		agora.StoreName:        false,
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

func TestFetchCardsConcurrently_BinderposGate(t *testing.T) {
	var activeBinderpos int32
	var maxActiveBinderpos int32

	binderposShops := []string{
		cardscitadel.StoreName,
		cardaffinity.StoreName,
		hideout.StoreName,
		gameshaven.StoreName,
		tefuda.StoreName,
		gog.StoreName,
	}
	shops := make(map[string]gateway.LGS, len(binderposShops)+1)
	for _, name := range binderposShops {
		shops[name] = &MockLGS{
			SearchFunc: func(ctx context.Context, searchStr string) ([]gateway.Card, error) {
				active := atomic.AddInt32(&activeBinderpos, 1)
				for {
					currentMax := atomic.LoadInt32(&maxActiveBinderpos)
					if active <= currentMax || atomic.CompareAndSwapInt32(&maxActiveBinderpos, currentMax, active) {
						break
					}
				}
				time.Sleep(80 * time.Millisecond)
				atomic.AddInt32(&activeBinderpos, -1)
				return []gateway.Card{{Name: searchStr, InStock: true, Price: 1, Source: cardscitadel.StoreName}}, nil
			},
		}
	}

	shops[agora.StoreName] = &MockLGS{
		SearchFunc: func(ctx context.Context, searchStr string) ([]gateway.Card, error) {
			return []gateway.Card{{Name: searchStr, InStock: true, Price: 1, Source: agora.StoreName}}, nil
		},
	}

	_, _ = fetchCardsConcurrently(context.Background(), "Abrade", shops)
	if atomic.LoadInt32(&maxActiveBinderpos) > binderposMaxConcurrent {
		t.Fatalf("expected at most %d concurrent binderpos searches, got %d", binderposMaxConcurrent, maxActiveBinderpos)
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

func TestFormatDiscordErrorSummary(t *testing.T) {
	got := formatDiscordErrorSummary("Uro, Titan of Nature's Wrath", []string{
		"Error encountered searching [Tefuda] for [Uro, Titan of Nature's Wrath]: attempt 3 (scrap-direct): Service Unavailable (proxy_mode=direct proxy=none)",
		"Error encountered searching [Arcane Sanctum] for [Uro, Titan of Nature's Wrath]: attempt 2 (scrap-direct): Service Unavailable (proxy_mode=direct proxy=none)",
		"Recovered from panic in shop [ShopPanic]: panic value",
	})

	if !strings.Contains(got, "Encountered 3 error(s) while searching [Uro, Titan of Nature's Wrath]:") {
		t.Fatalf("expected summary header, got: %s", got)
	}
	if !strings.Contains(got, "- [Arcane Sanctum] attempt 2 (scrap-direct): Service Unavailable (proxy_mode=direct proxy=none)") {
		t.Fatalf("expected Arcane Sanctum concise line, got: %s", got)
	}
	if !strings.Contains(got, "- [Tefuda] attempt 3 (scrap-direct): Service Unavailable (proxy_mode=direct proxy=none)") {
		t.Fatalf("expected Tefuda concise line, got: %s", got)
	}
	if !strings.Contains(got, "- Recovered from panic in shop [ShopPanic]: panic value") {
		t.Fatalf("expected fallback line for non-search error, got: %s", got)
	}
}
