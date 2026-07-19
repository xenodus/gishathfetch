package binderpos

import (
	"fmt"
	"strings"

	"mtg-price-checker-sg/gateway"
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
func runFallbackAttempts(attempts ...fallbackAttempt) ([]gateway.Card, error) {
	var (
		cards           []gateway.Card
		err             error
		executedAttempt int
	)

	for _, attempt := range attempts {
		executedAttempt++
		cards, err = attempt.fn()
		err = annotateAttemptError(executedAttempt, attempt.strategy, err)

		if err == nil {
			return cards, nil
		}

		if (attempt.family == strategyFamilyScrap || attempt.family == strategyFamilyGraphQL) && gateway.IsHTTPServerError(err) {
			return cards, err
		}
	}

	return cards, err
}

func annotateAttemptError(attempt int, strategy string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("attempt %d (%s): %w", attempt, strategy, err)
}
