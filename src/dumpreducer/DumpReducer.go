package dumpreducer

import (
	"sync"

	"../structures"
	"../utils"
	"github.com/negapedia/wikibrief"
)

func keepLastNRevert(page *structures.Page, nRev int) {
	if len(page.Revision) > nRev { // if page history is longer than limits, otherwise skip
		startRemovedIndex := -1
		for i := len(page.Revision) - 1; i >= 0; i-- { //the last is the more recent
			nRev--
			if nRev == 0 {
				startRemovedIndex = i
			} else if nRev < 0 {
				page.Revision[i] = structures.Revision{} // clean revision
			}
		}

		if startRemovedIndex != -1 {
			page.Revision = page.Revision[startRemovedIndex:]
		}
	}
}

func writePage(page *wikibrief.EvolvingPage, revArray *[]structures.Revision, nRevision int, resultDir string) {
	if len(*revArray) > 0 { // if page has revisions
		pageToWrite := structures.Page{PageID: page.PageID, TopicID: page.TopicID, Revision: *revArray}

		if nRevision != 0 { // if reverts limit is set
			keepLastNRevert(&pageToWrite, nRevision)
		}

		utils.WriteCleanPage(resultDir, &pageToWrite)
	}
}

// DumpReducer reduce the page information applying filters to it, like revert time frame, revert number and special page list
func DumpReducer(channel <-chan wikibrief.EvolvingPage, resultDir string, nRevision int) {
	wg := sync.WaitGroup{}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for page := range channel {
				var revArray []structures.Revision
				for rev := range page.Revisions {
					if rev.IsRevert > 0 {
						revArray = append(revArray, structures.Revision{Text: rev.Text, Timestamp: rev.Timestamp})
					}
				}
				writePage(&page, &revArray, nRevision, resultDir)
			}
		}()
	}
	wg.Wait()
}
