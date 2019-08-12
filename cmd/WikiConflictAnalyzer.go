package main

import (
	"context"
	"flag"
	"github.com/ebonetti/ctxutils"
	WDCA "github.com/negapedia/Wikipedia-Conflict-Analyzer"
	"github.com/negapedia/wikibrief"
	"log"
)

func main() {
	langFlag := flag.String("l", "", "Dump language")
	dirFlag := flag.String("d", "", "Result dir")
	startDateFlag := flag.String("s", "", "Revision starting date")
	endDateFlag := flag.String("e", "", "Revision ending date")
	specialPageListFlag := flag.String("specialList", "", "Special page list, page not in this list will be ignored. Input PageID like: id1-id2-...")
	nRevert := flag.Int("r", 0, "Number of revert limit")
	nTopWords := flag.Int("t", 0, "Number of top words to process")
	compressFinalOut := flag.Bool("delete", true, "If true compress in 7z and delete the final output folder")
	verboseMode := flag.Bool("verbose", true, "If true verbouse mode on")
	flag.Parse()

	wd := new(WDCA.WikiDumpConflitcAnalyzer)

	wd.NewWikiDump(*langFlag, *dirFlag, *startDateFlag, *endDateFlag, *specialPageListFlag,
		*nRevert, *nTopWords, *compressFinalOut, *verboseMode)

	ctx, fail := ctxutils.WithFail(context.Background())
	pageChannel := wikibrief.New(ctx, fail, wd.ResultDir, wd.Lang)
	wd.Preprocess(pageChannel)

	if err := fail(nil); err != nil {
		log.Fatal("%+v", err)
	}

	err := wd.Process()
	if err != nil{
		log.Fatal(err)
	}
	wd.CompressResultDir("/Result/")
}
