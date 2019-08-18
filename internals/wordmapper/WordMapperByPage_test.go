package wordmapper

import (
	"encoding/json"
	"testing"

	"github.com/negapedia/wikiconflict/internals/structures"

	"github.com/davecgh/go-spew/spew"
)

func Test_getMappedPage(t *testing.T) {
	pagestr := []byte(`{"PageID": 12345, "Revision": [{"Text": ["go", "go", "gopher", "lang", "code", "gopher"], "Timestamp": "2008-11-22T21:46:24Z"}], "TopicID": 2147483638}`)
	var page structures.StemmedPageJSON

	_ = json.Unmarshal(pagestr, &page)
	spew.Dump(page.Revision)

	result := getMappedPage(&page)

	if result.Word["go"] == 2 && result.Word["lang"] == 1 &&
		result.Word["gopher"] == 2 && result.Word["code"] == 1 {
		t.Log("Success")
	} else {
		t.Error("fail")
	}
}
