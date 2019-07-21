package structures

// BadWordsReport represent the data structure of the badwords report
type BadWordsReport struct {
	Abs  uint32
	Rel  float64
	BadW map[string]int
}
