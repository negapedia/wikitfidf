package datastructure

// PageContainer represent a list of PageElement, which are page containing complete data about word frequency
type PageContainer struct {
	PageList []PageElement
}

// PageElement represent a page containing complete data about word frequency
type PageElement struct {
	PageId uint32
	Word   map[string]float64
}
