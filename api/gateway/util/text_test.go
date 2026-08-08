package util

import "testing"

func TestStripDiacritics(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"Kíli the Resourceful":    "Kili the Resourceful",
		"Kili the Resourceful":     "Kili the Resourceful",
		"Juzám Djinn":             "Juzam Djinn",
		"Lim-Dûl the Necromancer": "Lim-Dul the Necromancer",
		"Palantír of Orthanc":     "Palantir of Orthanc",
		"Lightning Bolt":          "Lightning Bolt",
	}

	for input, want := range tests {
		if got := StripDiacritics(input); got != want {
			t.Errorf("StripDiacritics(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestFoldForMatch(t *testing.T) {
	t.Parallel()

	if FoldForMatch("Kíli the Resourceful") != FoldForMatch("Kili the Resourceful") {
		t.Fatal("expected accented and ASCII variants to fold equally")
	}
	if FoldForMatch("Juzám Djinn") != "juzam djinn" {
		t.Fatalf("FoldForMatch() = %q, want %q", FoldForMatch("Juzám Djinn"), "juzam djinn")
	}
}
