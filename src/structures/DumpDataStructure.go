/*
 * Developed by Marco Chilese.
 * Last modified 7/18/19 5:57 PM
 *
 */

package structures

import "time"

// Page represent the information in a wikipedia page
type Page struct {
	PageID   uint32
	Revision []Revision
}

// Revision represent the information of revert in a wikipedia page
type Revision struct {
	Timestamp time.Time
	Text      string
}
