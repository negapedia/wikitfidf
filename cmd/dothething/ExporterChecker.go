package main

import (
	"context"

	"github.com/ebonetti/ctxutils"

	"github.com/davecgh/go-spew/spew"

	"github.com/negapedia/wikitfidf"
)

func main() {
	e := wikitfidf.Exporter{Lang: "it", ResultDir: "/Users/marcochilese/Desktop/TEST/"}

	ctx, fail := ctxutils.WithFail(context.Background())
	topicCh := e.Topics(ctx, fail)

	for topic := range topicCh {
		spew.Dump(topic)
	}

}
