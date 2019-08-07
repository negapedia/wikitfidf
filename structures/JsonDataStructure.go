/*
 * Developed by Marco Chilese.
 * Last modified 7/18/19 6:07 PM
 *
 */

package structures

// StemmedPageJson represent a page written in json after the tokenization, stopwords cleaning and stemming process
type StemmedPageJson struct {
	PageID   uint32                `json:"PageID"`
	TopicID  uint32                `json:"TopicID"`
	Revision []stemmedRevisionJSON `json:"Revision"`
}

type stemmedRevisionJSON struct {
	Text []string `json:"Text"`
}
