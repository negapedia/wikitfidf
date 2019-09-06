/*
 * Developed by Marco Chilese.
 * Last modified 7/18/19 5:57 PM
 *
 */

package structures

// AggregatedPage represent a wikipedia page after the cleaning process and the word mapping process
type AggregatedPage struct {
	PageID  uint32
	TopicID uint32
	Tot     uint32
	Words   map[string]uint32
}

// TfidfAggregatedPage represent a wikipedia page word data after the cleaning process of TFIDF computation
type TfidfAggregatedPage struct {
	TopicID uint32
	Tot     uint32
	Words   *map[string]map[string]float64
}

// TfidfTopNWordPage represent a wikipedia page words data with only tfidf value
type TfidfTopNWordPage struct {
	TopicID uint32
	Tot     uint32
	Words   *map[string]float64
}

// PageTopNWords represent a Wikipedia page words data with only the top N words
type PageTopNWords struct {
	TopicID    uint32
	TotWords   uint32
	Word2TFIDF map[string]float64
	Word2Occur map[string]uint32
}
