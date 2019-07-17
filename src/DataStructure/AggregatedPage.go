package DataStructure

type AggregatedPage struct {
	PageID uint32
	Tot float64
	Words map[string]float64
}


type TfidfAggregatedPage struct {
	Tot float64
	Words *map[string]map[string]float64
}