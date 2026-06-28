package shopifysuggest

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// ProbeSuggestStructure fetches the predictive search endpoint and verifies the
// JSON response shape is unchanged. An empty products array is acceptable.
func ProbeSuggestStructure(ctx context.Context, opts Options) error {
	if opts.MapProduct == nil {
		return fmt.Errorf("shopifysuggest: MapProduct is required")
	}

	apiURL, err := buildSuggestURL(opts)
	if err != nil {
		return err
	}

	var lastErr error
	for _, attempt := range buildSearchAttempts() {
		body, err := doSuggestGETWithRetry(ctx, attempt.client, apiURL)
		if err != nil {
			lastErr = annotateSuggestAttemptError(1, attempt.strategy, err)
			continue
		}

		var res suggestResponse
		if err := json.Unmarshal(body, &res); err != nil {
			lastErr = annotateSuggestAttemptError(1, attempt.strategy, fmt.Errorf("suggest response is not valid JSON: %w", err))
			continue
		}
		if res.Resources.Results.Products == nil {
			lastErr = annotateSuggestAttemptError(1, attempt.strategy, fmt.Errorf("suggest response missing resources.results.products"))
			continue
		}
		return nil
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("suggest structure probe failed with no attempts")
}

// RequireSuggestStructure is a testify wrapper around ProbeSuggestStructure.
func RequireSuggestStructure(t *testing.T, ctx context.Context, opts Options) {
	t.Helper()
	require.NoError(t, ProbeSuggestStructure(ctx, opts))
}
