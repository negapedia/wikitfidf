package main

import (
	"./DumpReductor"
	"./TFIDF"
	"./WordMapper"
	"./Utils"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type WikiDump struct {
	lang string
	date string

	downloadDir string
	resultDir   string
	specialPageList *[]string
	startDate string
	endDate string
}

func NewWikiDump(lang string, date string, resultDir string, specialPageList *[]string,
	startDate string, endDate string) *WikiDump {
	p := new(WikiDump)
	p.lang = lang
	p.date = date
	p.resultDir = resultDir + lang + "_" + date + "/"
	p.specialPageList = specialPageList
	p.startDate = startDate
	p.endDate = endDate

	if _, err := os.Stat(p.resultDir + "Stem"); os.IsNotExist(err) {
		err = os.MkdirAll(p.resultDir+"Stem", 0755)
		if err != nil {
			panic(err)
		}
	}

	return p
}

func process(wd *WikiDump, linkToDownload []*Utils.DumpLink) {
	nFile := len(linkToDownload)

	for i, link := range linkToDownload {
		fmt.Printf("\rOn %d/%d \n%v", i+1, nFile, link.Name)
		Utils.DownloadFile(wd.resultDir+link.Name, link.Link) //TODO remove comment

		println("\nParse and reduction start")
		DumpReductor.ParseDump(wd.resultDir+link.Name, wd.resultDir, wd.startDate, wd.endDate, wd.specialPageList) //("../103KB_test.7z", wd.resultDir, wd.startDate, wd.endDate, wd.specialPageList)// //startDate and endDate must be in the same format of dump timestamp!
		println("Parse and reduction end")

		println("WikiMarkup cleaning start")
		wikiMarkupClean := exec.Command("java","-jar", "./TextNormalizer/WikipediaMarkupCleaner.jar", wd.resultDir)
		_ = wikiMarkupClean.Run()
		println("WikiMarkup cleaning end")

		println("Stopwords cleaning and stemming start")
		stopwordsCleanerStemming := exec.Command("python3","./TextNormalizer/runStopwClean.py", wd.resultDir, wd.lang)
		_ = stopwordsCleanerStemming.Run()
		println("Stopwords cleaning and stemming end")

		println("Word mapping by page start")
		WordMapper.WordMapperByPage(wd.resultDir)
		println("Word mapping by page end")

		break
	}
}

func main() {
	//defer profile.Start().Stop()	//"github.com/pkg/profile"

	langFlag := flag.String("l", "", "Dump language")
	dateFlag := flag.String("d", "", "Dump date")
	startDateFlag := flag.String("s", "", "Revision starting date")
	endDateFlag := flag.String("e", "", "Revision ending date")
	specialPageListFlag := flag.String("specialList", "", "Special page list, page not in this list will be ignored. Input PageID like: id1-id2-...")
	flag.Parse()

	println(*langFlag)
	println(*dateFlag)
	println(*specialPageListFlag)
	specialPageList := func(specialPageListFlag string) []string {
		if specialPageListFlag == ""{
			return nil
		} else {
			return strings.Split(specialPageListFlag, "-")
		}
	}(*specialPageListFlag)

	wd := NewWikiDump(*langFlag, *dateFlag, "../Result/", &specialPageList, *startDateFlag, *endDateFlag)

	linkToDownload := Utils.DumpLinkGetter(wd.lang, wd.date)

	process(wd, linkToDownload)

	println("Processing GlobalWordMap file start")
	WordMapper.GlobalWordMapper(wd.resultDir)
	println("Processing GlobalWordMap file start")

	println("Processing GlobalStem file start")
	WordMapper.StemRevAggregator(wd.resultDir)
	println("Processing GlobalStem file end")

	println("Processing GlobalPage file start")
	WordMapper.PageMapAggregator(wd.resultDir)
	println("Processing GlobalPage file end")

	println("Processing TFIDF file start")
	TFIDF.ComputeTFIDF(wd.resultDir)
	println("Processing TFIDF file end")

	println("Performing Destemming start")
	deStemming := exec.Command("python3","./DeStemmer/runDeStemming.py", wd.resultDir)
	_ = deStemming.Run()
	println("Performing Destemming file end")
}
