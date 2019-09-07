package structures

// BadWordsReport represent the data structure of the badwords report
type BadWordsReport struct {
	TopicID uint32
	Abs     uint32
	Rel     float64
	BadW    map[string]uint32
}
