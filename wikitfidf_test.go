package wikitfidf

import (
	"testing"
)

func TestCheckAvailableLanguage(t *testing.T) {
	lang1 := CheckAvailableLanguage("en")
	lang2 := CheckAvailableLanguage("abc")
	lang3 := CheckAvailableLanguage("eml")

	if lang1 == nil && lang2 != nil && lang3 != nil {
		t.Log("Success.")
	} else {
		t.Errorf("Failed.")
	}
}
