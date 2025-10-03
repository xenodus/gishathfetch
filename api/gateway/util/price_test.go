package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePrice(t *testing.T) {
	type args struct {
		givenPrice string
		expResult  float64
		expErr     bool
	}
	tcs := map[string]args{
		"valid with everything": {
			givenPrice: "From SGD S$1,234.56",
			expResult:  1234.56,
		},
		"valid with From": {
			givenPrice: "From 1234.56",
			expResult:  1234.56,
		},
		"valid with S$": {
			givenPrice: "S$1234.56",
			expResult:  1234.56,
		},
		"valid with $": {
			givenPrice: "$1234.56",
			expResult:  1234.56,
		},
		"valid with ,": {
			givenPrice: "1,234.56",
			expResult:  1234.56,
		},
		"valid with SGD": {
			givenPrice: "SGD 1234.56",
			expResult:  1234.56,
		},
		"invalid float64": {
			givenPrice: "From SGD S$ABC",
			expErr:     true,
		},
	}
	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			result, err := ParsePrice(tc.givenPrice)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expResult, result)
			}
		})
	}
}
