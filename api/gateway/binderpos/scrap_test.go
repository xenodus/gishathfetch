package binderpos

import (
	"errors"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Scrap(t *testing.T) {
	type args struct {
		scrapVariant int
		storeName    string
		baseUrl      string
		searchUrl    string
		searchStr    string
		expErr       error
	}
	tests := map[string]args{
		"variant 1": {
			scrapVariant: 1,
			storeName:    "Cards Citadel",
			baseUrl:      "https://cardscitadel.com",
			searchUrl:    "/search?q=*%s*",
			searchStr:    "Abrade",
		},
		"variant 2": {
			scrapVariant: 2,
			storeName:    "OneMtg",
			baseUrl:      "https://onemtg.com.sg",
			searchUrl:    "/search?q=%s",
			searchStr:    "Abrade",
		},
		"variant 2 - CA": {
			scrapVariant: 2,
			storeName:    "Card Affinity",
			baseUrl:      "https://card-affinity.com",
			searchUrl:    "/search?q=%s",
			searchStr:    "chocobo%20camp",
		},
		"variant 3": {
			scrapVariant: 3,
			storeName:    "Grey Ogre Games",
			baseUrl:      "https://www.greyogregames.com",
			searchUrl:    "/search?q=%s",
			searchStr:    "Abrade",
		},
		"variant 4": {
			scrapVariant: 4,
			storeName:    "Tefuda",
			baseUrl:      "https://tefudagames.com",
			searchUrl:    "/search?q=%s",
			searchStr:    "smothering tithe",
		},
		"variant 5": {
			scrapVariant: 5,
			storeName:    "Arcane Sanctum",
			baseUrl:      "https://arcanesanctumtcg.com",
			searchUrl:    "/search?q=%s",
			searchStr:    "signet",
		},
		"invalid variant": {
			scrapVariant: 999,
			expErr:       errors.New("invalid scrap variant: 999"),
		},
	}
	for testName, testArg := range tests {
		t.Run(testName, func(t *testing.T) {
			i := New()
			result, err := i.Scrap(
				testArg.scrapVariant,
				testArg.storeName,
				testArg.baseUrl,
				testArg.searchUrl,
				testArg.searchStr,
			)

			if testArg.expErr != nil {
				require.Error(t, err)
				require.Equal(t, testArg.expErr.Error(), err.Error())
				return
			} else {
				require.NoError(t, err)
				require.True(t, len(result) > 0)

				for _, card := range result {
					if card.InStock {
						require.NotEmpty(t, card.Name)
						require.NotEmpty(t, card.Source)
						require.NotEmpty(t, card.Url)
						require.NotEmpty(t, card.Img)
						require.NotEmpty(t, card.Price)
						require.Contains(t, card.Url, testArg.baseUrl+"/products/")

						log.Println(card.Name)
					}
				}
			}
		})
	}
}
