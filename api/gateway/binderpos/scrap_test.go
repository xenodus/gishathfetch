package binderpos

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScrapInvalidVariant(t *testing.T) {
	type args struct {
		scrapVariant int
		storeName    string
		baseUrl      string
		searchUrl    string
		searchStr    string
		expErr       error
	}
	tests := map[string]args{
		"invalid variant": {
			scrapVariant: 999,
			expErr:       errors.New("invalid scrap variant: 999"),
		},
	}
	for testName, testArg := range tests {
		t.Run(testName, func(t *testing.T) {
			i := New()
			result, err := i.Scrap(
				context.Background(),
				testArg.scrapVariant,
				testArg.storeName,
				testArg.baseUrl,
				testArg.searchUrl,
				testArg.searchStr,
			)

			require.Error(t, err)
			require.Equal(t, testArg.expErr.Error(), err.Error())
			require.Empty(t, result)
		})
	}
}
