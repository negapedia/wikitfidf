/*
 * Developed by Marco Chilese.
 * Last modified 7/18/19 6:00 PM
 *
 */

package structures

// PageContainer represent a list of PageElement, which are page containing complete data about word frequency
type PageContainer struct {
	PageList []PageElement
}

// PageElement represent a page containing complete data about word frequency
type PageElement struct {
	PageId  uint32
	TopicID uint32
	Word    map[string]uint32
}
