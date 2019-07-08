package DataStructure

type AggregatedPage struct {
	Title string
	Tot float64
	Words map[string]float64
}


type TfidfAggregatedPage struct {
	Title string
	Tot float64
	Words *map[string]map[string]float64
}