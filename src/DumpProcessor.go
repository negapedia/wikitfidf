package main

import (
	"./Utils"
	"./DumpReductor"
	"os"
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
	wd := NewWikiDump(os.Args[1], os.Args[2], "../Tmp/", "../Result/")

	linkToDownload := Utils.DumpLinkGetter(wd.lang, wd.date)

	for _, link := range linkToDownload{
		println(link.Link)
		Utils.DownloadFile(wd.resultDir+link.Name, link.Link)
		DumpReductor.ParseDump(wd.resultDir+link.Name, "", "") //startDate and endDate must be in the same format of dump timestamp!
	}


}
