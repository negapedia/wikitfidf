package main

import (
	"./DumpReducer"
	"./TFIDF"
	"./WordMapper"
	"context"
	"github.com/ebonetti/wikidump"
	"github.com/negapedia/wikibrief"
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

func (wd *WikiDump) NewWikiDump(lang string, resultDir string, specialPageList *[]uint32,
	startDate time.Time, endDate time.Time, nRevert int) {
	wd.lang = lang
	if startDate.IsZero() && endDate.IsZero(){
		wd.date = time.Now().Month().String()+strconv.Itoa(time.Now().Year())
	} else if startDate.IsZero() && !endDate.IsZero(){
		wd.date = startDate.String()+"-"+time.Now().Month().String()+strconv.Itoa(time.Now().Year())
	} else {
		wd.date = time.Now().Month().String()+strconv.Itoa(time.Now().Year())+"-"+endDate.String()
	}
	wd.resultDir = resultDir + lang + "_" + wd.date + "/"
	wd.specialPageList = specialPageList
	wd.startDate = startDate
	wd.endDate = endDate
	wd.nRevert = nRevert

	if _, err := os.Stat(wd.resultDir + "Stem"); os.IsNotExist(err) {
		err = os.MkdirAll(wd.resultDir+"Stem", 0755)
		if err != nil {
			panic(err)
		}
	}
}

func (wd *WikiDump) PreProcess(channel chan wikibrief.EvolvingPage) {
		println("\nParse and reduction start")
		DumpReducer.DumpReducer(channel, wd.resultDir, wd.lang, wd.startDate, wd.endDate, wd.specialPageList,  wd.nRevert) //("../103KB_test.7z", wd.resultDir, wd.startDate, wd.endDate, wd.specialPageList)// //startDate and endDate must be in the same format of dump timestamp!
		println("Parse and reduction end")
}

func (wd *WikiDump) Process() {
	//defer profile.Start().Stop()	//"github.com/pkg/profile"

	/*langFlag := flag.String("l", "", "Dump language")
	//dateFlag := flag.String("d", "", "Dump date")
	startDateFlag, _ := time.Parse("2019-07-01T16:00:00", *flag.String("s", "", "Revision starting date"))
	endDateFlag, _:= time.Parse("2019-07-01T16:00:00", *flag.String("e", "", "Revision ending date"))
	//specialPageListFlag := flag.String("specialList", "", "Special page list, page not in this list will be ignored. Input PageID like: id1-id2-...")

	nRevert := flag.Int("N", 0, "Number of revert to consider")
	flag.Parse()

	wd := new(WikiDump)
	wd.NewWikiDump()*/


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

func main(){
	wd := new(WikiDump)
	wd.NewWikiDump("vec", "/Users/marcochilese/Desktop/Tesi/NegapediaConflicutalWords/Result/",nil, time.Time{}, time.Time{}, 10)

	dump, err := wikidump.Latest(wd.resultDir, wd.lang, "metahistory7zdump")
	if err != nil {
		panic(err)
	}

	it := dump.Open("metahistory7zdump")
	reader, err := it(context.Background())
	if err != nil {
		panic(err)
	}

	channel := make(chan wikibrief.EvolvingPage)
	wd.PreProcess(channel)

	err = wikibrief.Transform(context.Background(), reader, func(uint32) bool { return true }, channel)
	if err != nil {
		panic(err)
	}

	wd.Process()
}
