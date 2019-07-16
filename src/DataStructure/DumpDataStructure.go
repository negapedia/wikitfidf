package DataStructure

import "time"

type Page struct {
	//Title string `xml:"title"`
	//Ns string `xml:"ns"`
	PageID uint32
	Revision []Revision
}

type Revision struct {
	Timestamp time.Time
	Text string
	//Sha1 string `xml:"sha1"`
	//Reverted bool	// goes to False by default while parsing
}
