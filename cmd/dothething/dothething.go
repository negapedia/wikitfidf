package main

import (
	"context"
	"flag"
	"log"

	"github.com/ebonetti/ctxutils"
	"github.com/negapedia/wikibrief"
	"github.com/negapedia/wikitfidf"
)

func main() {
	langFlag := flag.String("lang", "", "Dump language")
	dirFlag := flag.String("d", "", "Result dir")
	startDateFlag := flag.String("s", "", "Revision starting date")
	endDateFlag := flag.String("e", "", "Revision ending date")
	specialPageListFlag := flag.String("specialList", "", "Special page list, page not in this list will be ignored. Input PageID like: id1-id2-...")
	nRevert := flag.Int("rev", 0, "Number of revert limit")
	nTopWordsPages := flag.Int("topPages", 0, "Number of top words per page to process")
	nTopWordsGlobal := flag.Int("topWords", 0, "Number of top global words to process")
	nTopWordsTopic := flag.Int("topTopic", 0, "Number of top topic words to process")
	testMode := flag.Bool("test", false, "If true verbose mode on and will be processed only a single dump")
	flag.Parse()

	if *langFlag == "" {
		panic("Please specify a language from the commandline")
	}

	ctx, fail := ctxutils.WithFail(context.Background())
	pageChannel := wikibrief.New(ctx, fail, *dirFlag, *langFlag, *testMode)

	pageChannel = Filter(ctx, fail, pageChannel, *startDateFlag, *endDateFlag, *specialPageListFlag)

	limits := wikitfidf.Limits{WordsPages: *nTopWordsPages, GlobalWords: *nTopWordsGlobal, TopicWords: *nTopWordsTopic, Reverts: *nRevert}
	if limits == (wikitfidf.Limits{}) {
		limits = wikitfidf.ReasonableLimits()
	}

	wd, err := wikitfidf.New(context.Background(), *langFlag, pageChannel, *dirFlag, limits, *testMode)
	if err != nil {
		fail(err)
	}

	if _, err := wikitfidf.From(*langFlag, wd.ResultDir); err != nil {
		fail(err)
	}

	if err := fail(nil); err != nil {
		log.Fatalf("%+v", err)
	}
}
