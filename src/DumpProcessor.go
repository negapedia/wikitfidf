package main

import (
	"./DumpReducer"
	"./TFIDF"
	"./WordMapper"
	"flag"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type WikiDump struct {
	lang string
	date string

	downloadDir string
	resultDir   string
	specialPageList *[]uint32
	startDate time.Time
	endDate time.Time
	nRevert int
}

func NewWikiDump(lang string, resultDir string, specialPageList *[]uint32,
	startDate time.Time, endDate time.Time, nRevert int) *WikiDump {
	p := new(WikiDump)
	p.lang = lang
	if startDate.IsZero() && endDate.IsZero(){
		p.date = time.Now().Month().String()+strconv.Itoa(time.Now().Year())
	} else if startDate.IsZero() && !endDate.IsZero(){
		p.date = startDate.String()+"-"+time.Now().Month().String()+strconv.Itoa(time.Now().Year())
	} else {
		p.date = time.Now().Month().String()+strconv.Itoa(time.Now().Year())+"-"+endDate.String()
	}
	p.resultDir = resultDir + lang + "_" + p.date + "/"
	p.specialPageList = specialPageList
	p.startDate = startDate
	p.endDate = endDate
	p.nRevert = nRevert

	if _, err := os.Stat(p.resultDir + "Stem"); os.IsNotExist(err) {
		err = os.MkdirAll(p.resultDir+"Stem", 0755)
		if err != nil {
			panic(err)
		}
	}

	return p
}

func process(wd *WikiDump) {
		println("\nParse and reduction start")
		DumpReducer.ReduceDump(wd.resultDir, wd.lang, wd.startDate, wd.endDate, wd.specialPageList,  wd.nRevert) //("../103KB_test.7z", wd.resultDir, wd.startDate, wd.endDate, wd.specialPageList)// //startDate and endDate must be in the same format of dump timestamp!
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
}

func main() {
	//defer profile.Start().Stop()	//"github.com/pkg/profile"

	langFlag := flag.String("l", "", "Dump language")
	//dateFlag := flag.String("d", "", "Dump date")
	startDateFlag, _ := time.Parse("2019-07-01T16:00:00", *flag.String("s", "", "Revision starting date"))
	endDateFlag, _:= time.Parse("2019-07-01T16:00:00", *flag.String("e", "", "Revision ending date"))
	//specialPageListFlag := flag.String("specialList", "", "Special page list, page not in this list will be ignored. Input PageID like: id1-id2-...")

	nRevert := flag.Int("N", 0, "Number of revert to consider")
	flag.Parse()

	wd := NewWikiDump(*langFlag, "../Result/", nil, startDateFlag, endDateFlag, *nRevert)

	process(wd)

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
