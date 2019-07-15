package DumpParser
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
)

func ReduceDump(resultDir string, lang string, startDate string, endDate string, specialPageList *[]string) {
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
						revArray = append(revArray, DataStructure.Revision{Text: rev.Text, Timestamp: rev.Timestamp})
					}
				}
				if len(revArray) > 0 {
					fmt.Println(page.PageID)
					pageToWrite := DataStructure.Page{PageID:p.PageID, Revision:revArray}

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

/*
func main(){
	ReduceDump("./Result/Test/", "", "", nil)
}*/
