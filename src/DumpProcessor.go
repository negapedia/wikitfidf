package main

import (
	"./Utils"
	"./DumpReductor"
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
		Utils.DownloadFile(wd.resultDir+link.Name, link.Link)

		println("Parse and reduction start")
		DumpReductor.ParseDump(wd.resultDir+link.Name, wd.resultDir,"", "") //startDate and endDate must be in the same format of dump timestamp!
		println("Parse and reduction end")

		println("WikiMarkup cleaning start")
		wikiMarkupRemoval := exec.Command("python3", "./TextNormalizer/WikiMarkupCleaner.py", wd.resultDir+link.Name[:3]+".json")
		_ = wikiMarkupRemoval.Start()
		_ = wikiMarkupRemoval.Wait()
		println("WikiMarkup cleaning end")
	}
}
