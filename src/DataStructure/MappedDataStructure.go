package DataStructure

type PageContainer struct {
	PageList []PageElement
}

type PageElement struct {
	PageId uint32
	Word   map[string]float64
}

