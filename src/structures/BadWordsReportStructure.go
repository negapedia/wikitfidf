package structures

type BadWordsReport struct {
	Abs  uint32
	Rel  float64
	BadW map[string]int
}
