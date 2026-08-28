package telegrambot

import (
	"fmt"
	"strings"
)

func formatSearchReply(query string, summary *SearchSummary) string {
	if summary.ResultCount == 0 || summary.Cheapest == nil {
		return fmt.Sprintf("No in-stock matches for %q.", query)
	}

	card := summary.Cheapest
	finish := ""
	if card.IsFoil {
		finish = "foil"
	}

	details := strings.TrimSpace(strings.Join(filterNonEmpty([]string{card.Quality, finish, card.ExtraInfo}), " · "))

	store := card.Source
	if store == "" {
		store = "Unknown store"
	}

	lines := []string{
		fmt.Sprintf("%s — S$%.2f @ %s", card.Name, card.Price, store),
	}
	if details != "" {
		lines = append(lines, details)
	}
	if card.URL != "" {
		lines = append(lines, card.URL)
	}

	if summary.ResultCount > 1 {
		lines = append(lines, fmt.Sprintf("%d results — view all on Gishath Fetch:", summary.ResultCount))
	} else {
		lines = append(lines, "View on Gishath Fetch:")
	}
	lines = append(lines, summary.WebsiteURL)

	return strings.Join(lines, "\n")
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

const pricePromptBody = "Enter a card name to search"

func formatPricePrompt(user *User) (text, parseMode string) {
	if user == nil {
		return pricePromptBody, ""
	}
	if username := strings.TrimPrefix(strings.TrimSpace(user.Username), "@"); username != "" {
		return fmt.Sprintf("@%s, %s", username, pricePromptBody), ""
	}
	name := strings.TrimSpace(user.FirstName)
	if name == "" {
		name = "there"
	}
	if user.ID != 0 {
		return fmt.Sprintf(
			`<a href="tg://user?id=%d">%s</a>, %s`,
			user.ID,
			htmlEscape(name),
			pricePromptBody,
		), "HTML"
	}
	return pricePromptBody, ""
}

func isPricePrompt(text string) bool {
	return strings.Contains(strings.TrimSpace(text), pricePromptBody)
}

func htmlEscape(value string) string {
	value = strings.ReplaceAll(value, "&", "&amp;")
	value = strings.ReplaceAll(value, "<", "&lt;")
	value = strings.ReplaceAll(value, ">", "&gt;")
	return value
}
