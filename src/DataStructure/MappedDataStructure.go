package DataStructure

type PageContainer struct {
	PageList []PageElement
}

type PageElement struct {
	PageId string
	Title  string
	Word   map[string]uint64
}
