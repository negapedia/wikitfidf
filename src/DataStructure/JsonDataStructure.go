package DataStructure

type StemmedPageJson struct {
	PageID uint32 `json:"PageID"`
	Revision []stemmedRevision_Json `json:"Revision"`
}

type stemmedRevision_Json struct {
	Text []string `json:"Text"`
}
