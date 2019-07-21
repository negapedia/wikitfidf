package dumpreducer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"../structures"
	"github.com/negapedia/wikibrief"
)

func keepLastNRevert(page *structures.Page, nRev int) {
	if len(page.Revision) > nRev {
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

// DumpReducer reduce the page information applying filters to it, like revert time frame, revert number and special page list
func DumpReducer(channel <-chan wikibrief.EvolvingPage, resultDir string, startDate time.Time, endDate time.Time, specialPageList *[]uint32, nRevision int) {
	wg := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for page := range channel {
				var revArray []structures.Revision
				for rev := range page.Revisions {
					if rev.IsRevert > 0 {
						if !startDate.IsZero() || !endDate.IsZero() { // if data filter is setted
							timestamp := rev.Timestamp
							if !startDate.IsZero() && !endDate.IsZero() {
								if timestamp.Sub(startDate) >= 0 && timestamp.Sub(endDate) <= 0 {
									revArray = append(revArray, structures.Revision{Text: rev.Text, Timestamp: rev.Timestamp})
								}
							} else if startDate.IsZero() && !endDate.IsZero() {
								if timestamp.Sub(endDate) <= 0 {
									revArray = append(revArray, structures.Revision{Text: rev.Text, Timestamp: rev.Timestamp})
								}
							} else if !startDate.IsZero() && endDate.IsZero() {
								if timestamp.Sub(startDate) >= 0 {
									revArray = append(revArray, structures.Revision{Text: rev.Text, Timestamp: rev.Timestamp})
								}
							}
						} else if specialPageList != nil { // if page list is setted
							inList := func() bool {
								for _, pageID := range *specialPageList {
									if pageID == page.PageID {
										return true
									}
								}
								return false
							}
							if inList() {
								revArray = append(revArray, structures.Revision{Text: rev.Text, Timestamp: rev.Timestamp})
							}

						} else { // otherwise
							revArray = append(revArray, structures.Revision{Text: rev.Text, Timestamp: rev.Timestamp})
						}
					}
				}
				if len(revArray) > 0 {
					fmt.Println(page.PageID)
					pageToWrite := structures.Page{PageID: page.PageID, Revision: revArray}

					if nRevision != 0 { // if reverts limit is set
						keepLastNRevert(&pageToWrite, nRevision)
					}

					outFile, err := os.Create(resultDir + fmt.Sprint(page.PageID) + ".json")
					if err == nil {
						writer := bufio.NewWriter(outFile)
						defer outFile.Close()

						var dictPage, err = json.Marshal(pageToWrite)
						if err == nil {
							_, _ = writer.Write(dictPage)
							_ = writer.Flush()
						} else {
							panic(err)
						}
					}
				}
			}
		}()
	}
	wg.Wait()
}
