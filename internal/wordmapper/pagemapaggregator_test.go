package wordmapper

import (
	"math/rand"
	"testing"

	"github.com/negapedia/wikitfidf/internal/structures"
)

func Test_getTotalWordInPage(t *testing.T) {
	words := make(map[string]uint32)
	words["abc"] = 10
	words["def"] = 50
	words["ghi"] = 1
	words["lmn"] = 15

	page := structures.PageElement{
		PageID:  rand.Uint32(),
		TopicID: 0,
		Word:    words,
	}

	if getTotalWordInPage(&page) != 76 {
		t.Error("fail")
	} else {
		t.Log("Success")
	}
}
