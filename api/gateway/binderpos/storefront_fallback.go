package binderpos

import (
	"fmt"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/alert"
)

type strategyFamily int

const (
	strategyFamilyUnknown strategyFamily = iota
	strategyFamilyDecklist
	strategyFamilyScrap
	strategyFamilyGraphQL
)

type fallbackAttempt struct {
	strategy string
	family   strategyFamily
	fn       func() ([]gateway.Card, error)
}

// notifyStrategyFallback logs and Slack-alerts a mid-chain search strategy fallback.
// Tests may replace this to observe notifications without hitting webhooks.
var notifyStrategyFallback = alert.NotifySearchStrategyFallback

func strategyFamilyFromName(name string) strategyFamily {
	switch {
	case strings.HasPrefix(name, "decklist-"):
		return strategyFamilyDecklist
	case strings.HasPrefix(name, "scrap-"):
		return strategyFamilyScrap
	case strings.HasPrefix(name, "graphql-"):
		return strategyFamilyGraphQL
	default:
		return strategyFamilyUnknown
	}
}

// runFallbackAttempts runs the supplied attempts in order, returning the first
// result that returns cards.
//
// When any scrap, GraphQL, or decklist attempt returns no cards without error,
// that empty result is final and no further strategies run.
// HTTP 5xx errors on scrape and GraphQL attempts are final so a failing
// storefront is not followed by the shared portal or HTML scrap. Other GraphQL
// errors fall through to HTML scrap.
// Each attempt's error is annotated with its position and strategy name so the
// final error reflects the last attempt tried.
// When an attempt fails and another strategy will run, a Slack alert is sent
// (in addition to the console log) describing the fallback.
func runFallbackAttempts(storeName, searchStr string, attempts ...fallbackAttempt) ([]gateway.Card, error) {
	var (
		cards           []gateway.Card
		err             error
		executedAttempt int
	)

	for i, attempt := range attempts {
		executedAttempt++
		cards, err = attempt.fn()
		err = annotateAttemptError(executedAttempt, attempt.strategy, err)

		if err == nil {
			return cards, nil
		}

		if (attempt.family == strategyFamilyScrap || attempt.family == strategyFamilyGraphQL) && gateway.IsHTTPServerError(err) {
			return cards, err
		}

		if i+1 < len(attempts) {
			notifyStrategyFallback(formatStrategyFallback(storeName, searchStr, err, attempts[i+1].strategy))
		}
	}

	return cards, err
}

func formatStrategyFallback(storeName, searchStr string, failedErr error, nextStrategy string) string {
	return fmt.Sprintf(
		"Search strategy fallback [%s] for [%s]: %v; falling back to %s",
		storeName,
		searchStr,
		failedErr,
		nextStrategy,
	)
}

func annotateAttemptError(attempt int, strategy string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("attempt %d (%s): %w", attempt, strategy, err)
}
