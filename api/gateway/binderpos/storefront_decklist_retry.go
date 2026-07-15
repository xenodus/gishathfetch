package binderpos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"mtg-price-checker-sg/gateway"
)

// doDecklistRequestWithRetry sends the decklist request once. On success it
// returns the response with its body still open for the caller to decode.
func doDecklistRequestWithRetry(ctx context.Context, client *http.Client, newRequest func() (*http.Request, error)) (*http.Response, error) {
	req, err := newRequest()
	if err != nil {
		return nil, err
	}
	if err := gateway.WaitForDomainRequestSlot(ctx, req.URL); err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusOK {
		return res, nil
	}

	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	return nil, fmt.Errorf("binderpos decklist request failed status=%d body=%s", res.StatusCode, strings.TrimSpace(string(body)))
}
