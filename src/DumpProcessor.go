package main

import (
	"./DumpReductor"
	"./Utils"
	"./WordMapper"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type WikiDump struct {
	lang string
	date string

	downloadDir string
	resultDir   string
}

func NewWikiDump(lang string, date string, resultDir string) *WikiDump {
	p := new(WikiDump)
	p.lang = lang
	p.date = date
	p.resultDir = resultDir + lang + "_" + date + "/"

	//_ = os.MkdirAll(p.resultDir, os.ModeDir)
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
		fmt.Printf("\rOn %d/%d \n%v", i, nFile, link.Name)
		//Utils.DownloadFile(wd.resultDir+link.Name, link.Link) TODO remove comment

		println("Parse and reduction start")
		DumpReductor.ParseDump("../103KB_en.7z", wd.resultDir, "", "") //(wd.resultDir+link.Name, wd.resultDir,"", "") //startDate and endDate must be in the same format of dump timestamp! ("../113KB_test.7z", wd.resultDir, "", "")
		println("Parse and reduction end")

		println("WikiMarkup cleaning start")
		wikiMarkupRemoval := exec.Command("python3", "./TextNormalizer/WikiMarkupCleaner.py", wd.resultDir)
		_ = wikiMarkupRemoval.Start()
		_ = wikiMarkupRemoval.Wait()
		println("WikiMarkup cleaning end")

		println("Stopwords cleaning and stemming start")
		stopwordsCleanerStemming := exec.Command("python3", "./TextNormalizer/StopwordsCleaner_Stemming.py", wd.resultDir, wd.lang)
		_ = stopwordsCleanerStemming.Start()
		_ = stopwordsCleanerStemming.Wait()
		println("Stopwords cleaning and stemming end")

		println("Word mapping by page start")
		WordMapper.WordMapperByPage(wd.resultDir)
		println("Word mapping by page end")

		break
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	wd := NewWikiDump(os.Args[1], os.Args[2], "../Result/")

	linkToDownload := Utils.DumpLinkGetter(wd.lang, wd.date)

	process(wd, linkToDownload)

	println("Processing GlobalWordMap file start")
	WordMapper.GlobalWordMapper(wd.resultDir)
	println("Processing GlobalWordMap file start")

}
