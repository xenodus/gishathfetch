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
		fmt.Sprintf("%s - S$%.2f @ %s", card.Name, card.Price, store),
	}
	if details != "" {
		lines = append(lines, details)
	}
	if card.URL != "" {
		lines = append(lines, card.URL, "")
	}

	if summary.ResultCount > 1 {
		lines = append(lines, fmt.Sprintf("%d results - view all on Gishath Fetch:", summary.ResultCount))
	} else {
		lines = append(lines, "View on Gishath Fetch:")
	}
	lines = append(lines, summary.WebsiteURL)

	return strings.Join(lines, "\n")
}

func formatCKReply(query string, summary *CKSummary) string {
	if summary == nil || summary.Listing == nil {
		return fmt.Sprintf("No Card Kingdom listing for %q.", query)
	}

	listing := summary.Listing
	name := listing.CardName
	if name == "" {
		name = query
	}

	lines := []string{
		fmt.Sprintf("%s - US$%.2f @ Card Kingdom", name, listing.PriceUsd),
	}

	details := strings.TrimSpace(strings.Join(filterNonEmpty([]string{
		listing.Edition,
		ckFinishLabel(listing.IsFoil),
	}), " · "))
	if details != "" {
		lines = append(lines, details)
	}

	if listing.URL != "" {
		lines = append(lines, listing.URL)
	}

	return strings.Join(lines, "\n")
}

func ckFinishLabel(isFoil bool) string {
	if isFoil {
		return "foil"
	}
	return ""
}

const ckPromptBody = "Enter a card name for Card Kingdom price"

func formatCKPrompt(user *User) (text, parseMode string) {
	if user == nil {
		return ckPromptBody, ""
	}
	if username := strings.TrimPrefix(strings.TrimSpace(user.Username), "@"); username != "" {
		return fmt.Sprintf("@%s, %s", username, ckPromptBody), ""
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
			ckPromptBody,
		), "HTML"
	}
	return ckPromptBody, ""
}

func isCKPrompt(text string) bool {
	return strings.Contains(strings.TrimSpace(text), ckPromptBody)
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
