package DumpReducer
//package main

import (
	"../DataStructure"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ebonetti/wikidump"
	"github.com/negapedia/wikibrief"
	"os"
	"time"
)


func keepLastNRevert(page *DataStructure.Page, nRev int){
	if len(page.Revision)  > nRev {
		startRemovedIndex := -1
		for i := len(page.Revision)-1; i>=0; i-- { //the last is the more recent
			nRev--
			if nRev == 0{
				startRemovedIndex = i
			} else if nRev < 0 {
				page.Revision[i] = DataStructure.Revision{} // clean revision
			}
		}

		if startRemovedIndex != -1 {
			page.Revision = page.Revision[startRemovedIndex:]
		}
	}
}


func ReduceDump(resultDir string, lang string, startDate time.Time, endDate time.Time, specialPageList *[]uint32, nRevision int) {
	dump, err := wikidump.Latest(resultDir, lang, "metahistory7zdump")
	if err != nil {
		panic(err)
	}

	it := dump.Open("metahistory7zdump")
	reader,err := it(context.Background())
	if err!=nil {
		panic(err)
	}

	channel := make(chan wikibrief.EvolvingPage)

	go func() {
		for page := range channel{
			go func(p wikibrief.EvolvingPage) {
				var revArray []DataStructure.Revision
				for rev := range p.Revisions {
					if rev.IsRevert > 0 {
						if !startDate.IsZero() || !endDate.IsZero() {	// if data filter is setted
							timestamp := rev.Timestamp
							if !startDate.IsZero() && !endDate.IsZero() {
								if timestamp.Sub(startDate) >= 0 &&  timestamp.Sub(endDate) <= 0 {
									revArray = append(revArray, DataStructure.Revision{Text: rev.Text, Timestamp: rev.Timestamp})
								}
							} else if startDate.IsZero() && !endDate.IsZero() {
								if timestamp.Sub(endDate) <= 0 {
									revArray = append(revArray, DataStructure.Revision{Text: rev.Text, Timestamp: rev.Timestamp})
								}
							} else if !startDate.IsZero() && endDate.IsZero() {
								if timestamp.Sub(startDate) >= 0 {
									revArray = append(revArray, DataStructure.Revision{Text: rev.Text, Timestamp: rev.Timestamp})
								}
							}
						} else if specialPageList != nil {	// if page list is setted
							inList := func() bool {
								for _, pageID := range *specialPageList{
									if pageID == page.PageID{
										return true
									}
								}
								return false
							}
							if inList(){
								revArray = append(revArray, DataStructure.Revision{Text: rev.Text, Timestamp: rev.Timestamp})
							}

						} else {	// otherwise
							revArray = append(revArray, DataStructure.Revision{Text: rev.Text, Timestamp: rev.Timestamp})
						}
					}
				}
				if len(revArray) > 0 {
					fmt.Println(page.PageID)
					pageToWrite := DataStructure.Page{PageID:p.PageID, Revision:revArray}

					if nRevision != 0 {	// if reverts limit is set
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
						}
					}
				}
			}(page)
		}
	}()

	err = wikibrief.Transform(context.Background(), reader, func(uint32) bool { return true }, channel)
	if err!=nil {
		panic(err)
	}
}
