package Utils

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
)

type DumpLink struct {
	Link string
	Name string
}

func linkGetter(url string) []*DumpLink {

	doc, err := goquery.NewDocument(url)

	if err != nil {
		panic(err)
	}

	var links []*DumpLink

	doc.Find(".file").Each(func(i int, s *goquery.Selection) {
		doc.Find(".file a").Each(func(i int, s *goquery.Selection) {
			Link, _ := s.Attr("href")

			if strings.Contains(Link, "pages-meta-history") && strings.Contains(Link, ".7z") {
				l := DumpLink{"https://dumps.wikimedia.org"+Link, Link[17:]}
				links = append(links, &l)
			}
		})
	})

	return links
}


func DumpLinkGetter(lang string, date string) []*DumpLink {
	return linkGetter("https://dumps.wikimedia.org/"+lang+"wiki/"+date)
}
