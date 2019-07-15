package DataStructure

import "time"

type Page struct {
	//Title string `xml:"title"`
	//Ns string `xml:"ns"`
	PageID uint32 `xml:"id"`
	Revision []Revision `xml:"revision"`
}

type Revision struct {
	Timestamp time.Time `xml:"timestamp"`
	Text string `xml:"text"`
	//Sha1 string `xml:"sha1"`
	//Reverted bool	// goes to False by default while parsing
}
