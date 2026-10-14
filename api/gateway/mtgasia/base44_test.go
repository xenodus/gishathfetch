package mtgasia

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_cardPathSlug_matchesStorefront(t *testing.T) {
	require.Equal(t, "innistrad-remastered-abrade", cardPathSlug("Abrade", "Innistrad Remastered"))
	require.Equal(t, "abrade", cardPathSlug("Abrade", ""))
	require.Equal(t, "reality-fracture-aerid-konstrari", cardPathSlug("Aerid Konstrari", "Reality Fracture"))
}

func Test_buildCardSinglePageURL(t *testing.T) {
	got, err := buildCardSinglePageURL(
		"https://new.mtg-asia.com",
		"6ac33f1368fe2a219f23d9d0",
		"Abrade",
		"Innistrad Remastered",
	)
	require.NoError(t, err)
	require.Contains(t, got, "https://new.mtg-asia.com/card/innistrad-remastered-abrade/6ac33f1368fe2a219f23d9d0")
	require.Contains(t, got, "utm_source=")
}
