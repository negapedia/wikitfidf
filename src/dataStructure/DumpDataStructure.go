package dataStructure

import "time"

type Page struct {
	PageID uint32
	Revision []Revision
}

type Revision struct {
	Timestamp time.Time
	Text      string
}
