package dumpreducer

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/negapedia/wikibrief"
	"github.com/negapedia/wikitfidf/internals/structures"
)

func getRevChan() chan wikibrief.Revision {
	revch := make(chan wikibrief.Revision, 50)

	go func() {
		defer close(revch)
		for i := 0; i < 50; i++ {
			revch <- wikibrief.Revision{
				ID:        rand.Uint32(),
				UserID:    0,
				IsBot:     false,
				Text:      "testo",
				SHA1:      "abcd",
				IsRevert:  0,
				Timestamp: time.Now(),
			}
		}
	}()

	return revch
}

func getPageChan() chan wikibrief.EvolvingPage {
	ch := make(chan wikibrief.EvolvingPage, 50)

	go func() {
		defer close(ch)
		for i := 0; i < 50; i++ {
			ch <- wikibrief.EvolvingPage{
				PageID:    uint32(1000 + i),
				Title:     "page",
				Abstract:  "-",
				TopicID:   0,
				Revisions: getRevChan(),
			}
		}
	}()

	return ch
}

func Test_applySpecialListFilter(t *testing.T) {
	pages := getPageChan()

	specialPageList := []string{"1001", "1005"}

	for page := range pages {
		var revArray []structures.Revision
		for rev := range page.Revisions {
			applySpecialListFilter(&specialPageList, &page, &rev, &revArray)
		}

		inList := func() bool {
			for _, pageID := range specialPageList {
				if pageID == fmt.Sprint(page.PageID) {
					return true
				}
			}
			return false
		}

		if inList() && len(revArray) != 50 || !inList() && len(revArray) > 0 {
			t.Fail()
		}
	}

	t.Log("Success")
}

func Test_applyTimeFilter(t *testing.T) {
	ch := getRevChan()
	page := wikibrief.EvolvingPage{
		PageID:    1234567,
		Title:     "pagina1",
		Abstract:  "-",
		TopicID:   67890,
		Revisions: ch,
	}

	var revArray []structures.Revision

	for rev := range page.Revisions {
		applyTimeFilter(&rev, &revArray, time.Time{}, time.Now())
		t.Log(rev)
	}

	if len(revArray) != 50 {
		t.Error()
	} else {
		t.Log("Success")
	}

}

func Test_keepLastNRevert(t *testing.T) {
	page := structures.Page{PageID: 123456, TopicID: 456789, Revision: []structures.Revision{
		{Timestamp: time.Now(), Text: "rev1 testo"},
		{Timestamp: time.Now(), Text: "rev2 testo"},
		{Timestamp: time.Now(), Text: "rev3 testo"},
		{Timestamp: time.Now(), Text: "rev5 testo"},
		{Timestamp: time.Now(), Text: "rev6 testo"},
		{Timestamp: time.Now(), Text: "rev7 testo"},
		{Timestamp: time.Now(), Text: "rev8 testo"},
		{Timestamp: time.Now(), Text: "rev9 testo"},
		{Timestamp: time.Now(), Text: "rev10 testo"},
		{Timestamp: time.Now(), Text: "rev11 testo"},
		{Timestamp: time.Now(), Text: "rev12 testo"},
	}}

	keepLastNRevert(&page, 3)

	if len(page.Revision) != 3 && page.Revision[0].Text != "rev10 testo" &&
		page.Revision[1].Text != "rev11 testo" && page.Revision[2].Text != "rev12 testo" {
		t.Errorf("Failed")
	} else {
		t.Log("Success")
	}
}
