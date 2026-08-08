package scryfall

import "testing"

func TestFoldCardNameForMatch(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"Kíli the Resourceful":    "kili the resourceful",
		"Kili the Resourceful":      "kili the resourceful",
		"Juzám Djinn":             "juzam djinn",
		"Juzam Djinn":               "juzam djinn",
		"Lim-Dûl the Necromancer": "lim-dul the necromancer",
		"Lim-Dul the Necromancer": "lim-dul the necromancer",
		"Fíli and Kíli, Joyous":   "fili and kili, joyous",
		"Palantír of Orthanc":     "palantir of orthanc",
		"Ifh-Bíff Efreet":         "ifh-biff efreet",
		"Círdan the Shipwright":   "cirdan the shipwright",
		"Gríma Wormtongue":        "grima wormtongue",
		"Séance":                  "seance",
		"Seance":                  "seance",
	}

	for input, want := range tests {
		if got := FoldCardNameForMatch(input); got != want {
			t.Errorf("FoldCardNameForMatch(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCardNamesMatchForVerify(t *testing.T) {
	t.Parallel()

	pairs := [][2]string{
		{"Kíli the Resourceful", "Kili the Resourceful"},
		{"Juzám Djinn", "Juzam Djinn"},
		{"Lim-Dûl the Necromancer", "Lim-Dul the Necromancer"},
		{"Lightning Bolt", "lightning bolt"},
	}

	for _, pair := range pairs {
		if !cardNamesMatchForVerify(pair[0], pair[1]) {
			t.Errorf("cardNamesMatchForVerify(%q, %q) = false, want true", pair[0], pair[1])
		}
	}

	if cardNamesMatchForVerify("Lightning Bolt", "Counterspell") {
		t.Fatal("expected unrelated card names not to match")
	}
}
