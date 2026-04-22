package binderpos

import (
	"testing"
)

func TestStorefrontShopifyDomainForBaseURL(t *testing.T) {
	t.Run("returns mapped domain for known host with www", func(t *testing.T) {
		domain, ok := storefrontShopifyDomainForBaseURL("https://www.mtg-asia.com")
		if !ok {
			t.Fatalf("expected mapping to exist")
		}
		if domain != "mtgasia.myshopify.com" {
			t.Fatalf("unexpected domain: %s", domain)
		}
	})

	t.Run("returns false for unknown host", func(t *testing.T) {
		_, ok := storefrontShopifyDomainForBaseURL("https://example.com")
		if ok {
			t.Fatalf("expected unknown host lookup to fail")
		}
	})
}

func TestMapDecklistLinesToCards(t *testing.T) {
	lines := []storefrontDecklistLine{
		{
			ValidName: "Abrade",
			Products: []storefrontDecklistProduct{
				{
					Title:   "Abrade [Foundations]",
					Handle:  "abrade-foundations",
					SetName: "Foundations",
					Image:   "https://images.example/abrade.png",
					Variants: []storefrontDecklistStock{
						{
							ShopifyID: 111,
							Title:     "Near Mint Foil",
							Price:     0.85,
							Quantity:  2,
						},
					},
				},
			},
		},
	}

	t.Run("maps variant into card payload", func(t *testing.T) {
		cards := mapDecklistLinesToCards(2, "MTG Asia", "https://www.mtg-asia.com", lines)
		if len(cards) != 1 {
			t.Fatalf("expected 1 card, got %d", len(cards))
		}
		card := cards[0]
		if card.Name != "Abrade [Foundations] - Near Mint Foil" {
			t.Fatalf("unexpected card name: %s", card.Name)
		}
		if card.Price != 0.85 {
			t.Fatalf("unexpected card price: %f", card.Price)
		}
		if !card.InStock {
			t.Fatalf("expected card to be in stock")
		}
		if !card.IsFoil {
			t.Fatalf("expected foil card")
		}
	})

	t.Run("uses set name in extra info for variant 3", func(t *testing.T) {
		cards := mapDecklistLinesToCards(3, "Games Haven", "https://www.gameshaventcg.com", lines)
		if len(cards) != 1 {
			t.Fatalf("expected 1 card, got %d", len(cards))
		}
		if len(cards[0].ExtraInfo) != 1 || cards[0].ExtraInfo[0] != "Foundations" {
			t.Fatalf("unexpected extra info: %+v", cards[0].ExtraInfo)
		}
	})

	t.Run("skips variants with invalid URL data", func(t *testing.T) {
		invalid := []storefrontDecklistLine{
			{
				Products: []storefrontDecklistProduct{
					{
						Title:  "Broken Product",
						Handle: "",
						Variants: []storefrontDecklistStock{
							{
								ShopifyID: 10,
								Title:     "Near Mint",
								Price:     1.23,
								Quantity:  1,
							},
						},
					},
				},
			},
		}
		cards := mapDecklistLinesToCards(2, "Store", "https://store.example", invalid)
		if len(cards) != 0 {
			t.Fatalf("expected no cards, got %+v", cards)
		}
	})
}
