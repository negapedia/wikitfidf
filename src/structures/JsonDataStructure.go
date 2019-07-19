/*
 * Developed by Marco Chilese.
 * Last modified 7/18/19 6:07 PM
 *
 */

package structures

// StemmedPageJson represent a page written in json after the tokenization, stopwords cleaning and stemming process
type StemmedPageJson struct {
	PageID   uint32                `json:"PageID"`
	Revision []stemmedRevisionJson `json:"Revision"`
}

type stemmedRevisionJson struct {
	Text []string `json:"Text"`
}
