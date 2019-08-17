package wikiconflict

import (
	"testing"

	WC "github.com/negapedia/wikiconflict"
)

func TestCheckAvailableLanguage(t *testing.T) {
	lang1 := WC.CheckAvailableLanguage("en")
	lang2 := WC.CheckAvailableLanguage("abc")
	lang3 := WC.CheckAvailableLanguage("eml")

	if lang1 == nil && lang2 != nil && lang3 != nil {
		t.Log("Success.")
	} else {
		t.Errorf("Failed.")
	}
}
