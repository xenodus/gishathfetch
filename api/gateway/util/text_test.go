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

func TestStoreSearchQueries(t *testing.T) {
	t.Parallel()

	if got := StoreSearchQueries("Lightning Bolt"); len(got) != 1 || got[0] != "Lightning Bolt" {
		t.Fatalf("StoreSearchQueries() = %v, want single plain query", got)
	}

	got := StoreSearchQueries("Kíli")
	if len(got) != 2 || got[0] != "Kíli" || got[1] != "Kili" {
		t.Fatalf("StoreSearchQueries() = %v, want [Kíli Kili]", got)
	}
}
