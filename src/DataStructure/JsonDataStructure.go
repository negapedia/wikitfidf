package DataStructure

type StemmedPageJson struct {
	Title string `json:"Title"`
	Ns string `json:"Ns"`
	PageID string `json:"PageID"`
	Revision []stemmedRevision_Json `json:"Revision"`
}

type stemmedRevision_Json struct {
	Text []string `json:"Text"`
}
