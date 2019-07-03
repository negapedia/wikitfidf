package main

import (
	"./Utils"
	"./DumpReductor"
	"./WordMapper"
	"os"
	"os/exec"
	"runtime"
)

type WikiDump struct {
	lang string
	date string

	downloadDir string
	resultDir string
}

func NewWikiDump(lang string, date string, downloadDir string, resultDir string) *WikiDump {
	p:= new(WikiDump)
	p.lang = lang
	p.date = date
	p.downloadDir = downloadDir
	p.resultDir = resultDir+lang+"_"+date+"/"
	return p
}

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())

	wd := NewWikiDump(os.Args[1], os.Args[2], "../Tmp/", "../Result/")

	linkToDownload := Utils.DumpLinkGetter(wd.lang, wd.date)


	for _, link := range linkToDownload{
		println(link.Link)
		//Utils.DownloadFile(wd.resultDir+link.Name, link.Link) TODO remove comment

		println("Parse and reduction start")
		DumpReductor.ParseDump("../6MB_test.7z", wd.resultDir, "", "") //(wd.resultDir+link.Name, wd.resultDir,"", "") //startDate and endDate must be in the same format of dump timestamp! ("../113KB_test.7z", wd.resultDir, "", "")
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
