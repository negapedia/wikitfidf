package RevertCleaner

import (
	"../DataStructure"
	"sort"
	"../Utils"
)

// This function build an array of revisions' sha1 field. We will need this function for individuating reverts
func sha1ArrayBuilder(page *DataStructure.Page) []string {

	var sha1Array []string

	for _, rev := range page.Revision{
		sha1Array = append(sha1Array, rev.Sha1)
	}

	return sha1Array
}

// The function given an array returns the position of repeted sha1 sign
func getAllOccurencePosition(array []string) []int {
	var occurence []int
	for i, toFind := range array {
		found := false
		for j := i+1; j < len(array); j++{
			if array[j] == toFind {
				occurence = append(occurence, j)
				found = true
			}
			if found {
				occurence = append(occurence, i)
				found = false
			}
		}
 	}
	sort.Ints(occurence)
	return occurence
}

// The function given an array returns the gap between the array's number
// Like: [1 2 3 5] --> [4]
func getRevertsPosition(array []int) []int{
	var gaps []int
	for i, element := range array{
		if i+1 < len(array) && array[i+1] > element+1{
			start := element+1
			end := array[i+1]
			if start - end == 0{
				gaps = append(gaps, start)
			} else {
				for k := start; k < end; k++{
					gaps = append(gaps, k)
				}
			}
		}
	}
	return gaps
}


func nonRevertRemover(page *DataStructure.Page){
	var newRev []DataStructure.Revision

	for _, rev := range page.Revision {
		if rev.Reverted == true{
			newRev = append(newRev, rev)
		}
	}

	page.Revision = newRev
}


func RevertBuilder(page *DataStructure.Page) {
	repetedSha1 := getAllOccurencePosition(sha1ArrayBuilder(page))

	if len(repetedSha1) > 0 {

		revertedRevision := getRevertsPosition(repetedSha1)

		// page := getPage(dir, fileName)

		for _, element := range revertedRevision {
			page.Revision[element].Reverted = true
		}

		nonRevertRemover(page)

		Utils.WritePage("../out/", page)
	}
}
