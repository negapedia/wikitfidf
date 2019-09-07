package structures

// BadWordsReport represent the data structure of the badwords report
type BadWordsReport struct {
	TopicID uint32
	Abs     uint32
	Rel     float64
	BadW    map[string]uint32
}

// TopicBadWords represent the data structure of the badwords report for topic
type TopicBadWords struct {
	TotBadw uint32
	BadW    map[string]uint32
}
