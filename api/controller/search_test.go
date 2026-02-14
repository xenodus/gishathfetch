package controller

import (
	"fmt"
	"mtg-price-checker-sg/gateway"
	"testing"
)

// MockLGS is a mock implementation of gateway.LGS
type MockLGS struct {
	SearchFunc func(searchStr string) ([]gateway.Card, error)
}

func (m *MockLGS) Search(searchStr string) ([]gateway.Card, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(searchStr)
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
		extraInfo         []string
		expectedName      string
		expectedExtraInfo []string
	}{
		"Name [Tag]": {
			name:              "Name [Tag]",
			quality:           "",
			extraInfo:         []string{},
			expectedName:      "Name",
			expectedExtraInfo: []string{"[Tag]"},
		},
		"Name[Tag]": {
			name:              "Name[Tag]",
			quality:           "",
			extraInfo:         []string{},
			expectedName:      "Name",
			expectedExtraInfo: []string{"[Tag]"},
		},
		"Name (Tag)": {
			name:              "Name (Tag)",
			quality:           "",
			extraInfo:         []string{},
			expectedName:      "Name",
			expectedExtraInfo: []string{"(Tag)"},
		},
		"Name [Tag1] (Tag2)": {
			name:              "Name [Tag1] (Tag2)",
			quality:           "",
			extraInfo:         []string{},
			expectedName:      "Name",
			expectedExtraInfo: []string{"[Tag1] (Tag2)"},
		},
		"Name (Tag1) [Tag2]": {
			name:              "Name (Tag1) [Tag2]",
			quality:           "",
			extraInfo:         []string{},
			expectedName:      "Name",
			expectedExtraInfo: []string{"[Tag2]", "(Tag1)"},
		},
		"Name - Quality": {
			name:              "Name - Quality",
			quality:           "Quality",
			extraInfo:         []string{},
			expectedName:      "Name",
			expectedExtraInfo: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotName, gotExtra := cleanName(tt.name, tt.quality, tt.extraInfo)
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
		"Error Handling - LGS Error should not crash": {
			input: SearchInput{SearchString: "Card A"},
			lgsResponses: map[string][]gateway.Card{
				"Shop1": nil, // Simulates error or empty
				"Shop2": {
					{Name: "Card A", Price: 10.0, InStock: true, Source: "Shop2"},
				},
			},
			expectedCount: 1,
			verifyFunc: func(t *testing.T, cards []Card) {
				if cards[0].Source != "Shop2" {
					t.Errorf("Expected result from Shop2, got %s", cards[0].Source)
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
					SearchFunc: func(searchStr string) ([]gateway.Card, error) {
						if cardsToReturn == nil {
							return nil, fmt.Errorf("simulated error")
						}
						return cardsToReturn, nil
					},
				}
			}

			results, err := searchShops(tt.input, mockMap)
			if err != nil {
				t.Fatalf("searchShops returned unexpected error: %v", err)
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
