package telegrambot

import (
	"fmt"
	"strings"
)

func formatSearchReply(query string, summary *SearchSummary) string {
	if summary.ResultCount == 0 || summary.Cheapest == nil {
		return fmt.Sprintf(
			"No in-stock matches for %q.\n\nView on Gishath Fetch:\n%s",
			query,
			summary.WebsiteURL,
		)
	}

	card := summary.Cheapest
	finish := "non-foil"
	if card.IsFoil {
		finish = "foil"
	}

	details := strings.TrimSpace(strings.Join(filterNonEmpty([]string{card.Quality, finish, card.ExtraInfo}), " · "))
	if details == "" {
		details = finish
	}

	store := card.Source
	if store == "" {
		store = "Unknown store"
	}

	lines := []string{
		fmt.Sprintf("%s — S$%.2f @ %s", card.Name, card.Price, store),
		details,
	}
	if card.URL != "" {
		lines = append(lines, fmt.Sprintf("Buy: %s", card.URL))
	}

	if summary.ResultCount > 1 {
		lines = append(lines, fmt.Sprintf("%d results — view all on Gishath Fetch:", summary.ResultCount))
	} else {
		lines = append(lines, "View on Gishath Fetch:")
	}
	lines = append(lines, summary.WebsiteURL)

	return strings.Join(lines, "\n")
}

func formatHelpMessage() string {
	return "Search Singapore MTG singles prices.\n\n" +
		"Commands:\n" +
		"/price <card name> — cheapest in-stock match\n" +
		"/help — show this message\n\n" +
		"Example: /price Lightning Bolt"
}

func filterNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out
}
