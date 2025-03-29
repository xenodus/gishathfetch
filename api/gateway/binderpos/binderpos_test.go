package binderpos

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetCards_Success(t *testing.T) {
	givenStoreName := "MTG Asia"
	givenStoreBaseURL := "https://www.mtg-asia.com"
	givePayload, err := json.Marshal(Payload{
		StoreURL:    "mtgasia.myshopify.com",
		Game:        ProductTypeMTG.ToString(),
		Title:       "Abrade",
		InstockOnly: true,
	})
	require.NoError(t, err)

	cards, httpStatusCode, err := GetCards(givenStoreName, givenStoreBaseURL, givePayload)

	require.NoError(t, err)
	require.True(t, len(cards) > 0)
	require.Equal(t, http.StatusOK, httpStatusCode)

	for _, card := range cards {
		if card.InStock {
			require.NotEmpty(t, card.Name)
			require.NotEmpty(t, card.Source)
			require.NotEmpty(t, card.Url)
			require.NotEmpty(t, card.Img)
			require.NotEmpty(t, card.Price)
			require.Contains(t, card.Url, givenStoreBaseURL+productPath)
		}
	}
}

func Test_GetCards_HttpFailure(t *testing.T) {
	givenStoreName := "ASDF"
	givenStoreBaseURL := "ASDF"
	givePayload, err := json.Marshal(Payload{
		StoreURL:    "ASDF",
		Game:        ProductTypeMTG.ToString(),
		Title:       "Abrade",
		InstockOnly: true,
	})
	require.NoError(t, err)

	cards, httpStatusCode, err := GetCards(givenStoreName, givenStoreBaseURL, givePayload)

	require.NoError(t, err)
	require.Equal(t, 0, len(cards))
	require.NotEqual(t, http.StatusOK, httpStatusCode)
}
