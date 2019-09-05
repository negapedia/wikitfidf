package wikitfidf

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ebonetti/ctxutils"

	"github.com/negapedia/wikitfidf/internal/assets"
	"github.com/negapedia/wikitfidf/internal/badwords"

	"github.com/negapedia/wikitfidf/internal/topicwords"

	"github.com/negapedia/wikibrief"
	"github.com/negapedia/wikitfidf/internal/dumpreducer"
	"github.com/negapedia/wikitfidf/internal/tfidf"
	"github.com/negapedia/wikitfidf/internal/wordmapper"
	"github.com/pkg/errors"
)

// New ingests, processes and stores the desidered Wikipedia dump from the channel.
func New(ctx context.Context, lang string, in <-chan wikibrief.EvolvingPage, resultDir string, limits Limits, verboseMode bool) (exporter Exporter, err error) {
	ctx, fail := ctxutils.WithFail(ctx)

	wb := newBuilder(fail, lang, resultDir, limits, verboseMode).Preprocess(ctx, fail, in).Process(ctx, fail)

	if err = fail(nil); err == nil {
		exporter = Exporter{wb.ResultDir, wb.Lang}
	}

	return
}

//Limits represents limits at which data is cut off
type Limits struct {
	WordsPages  int
	GlobalWords int
	TopicWords  int

	Reverts int
}

//ReasonableLimits returns reasonable limits
func ReasonableLimits() Limits {
	return Limits{
		WordsPages:  50,
		GlobalWords: 100,
		TopicWords:  100,
		Reverts:     10,
	}
}

func newBuilder(fail func(error) error, lang string, resultDir string, limits Limits, verboseMode bool) (w builder) {
	err := CheckAvailableLanguage(lang)
	if err != nil {
		fail(err)
		return
	}

	if limits.WordsPages <= 0 || limits.GlobalWords <= 0 || limits.TopicWords <= 0 || limits.Reverts <= 0 {
		fail(errors.New("Invalid limits"))
		return
	}

	if resultDir, err = filepath.Abs(filepath.Join(resultDir, "TFIDF")); err != nil {
		fail(errors.WithStack(err))
		return
	}

	if err = os.MkdirAll(filepath.Join(resultDir, "Stem"), os.ModePerm); err != nil && !os.IsExist(err) {
		fail(errors.WithStack(err))
		return
	}

	logger := ioutil.Discard
	if verboseMode {
		logger = os.Stdout
	}

	return builder{Lang: lang, ResultDir: resultDir, Limits: limits, Logger: logger}
}

type builder struct {
	Lang      string
	ResultDir string

	Limits Limits

	Logger io.Writer
}

// Preprocess given a wikibrief.EvolvingPage channel reduce the amount of information in pages and save them
func (wt builder) Preprocess(ctx context.Context, fail func(error) error, channel <-chan wikibrief.EvolvingPage) builder {
	if ctx.Err() != nil {
		return wt
	}

	fmt.Fprintln(wt.Logger, "Parse and reduction")
	start := time.Now()
	dumpreducer.DumpReducer(ctx, fail, channel, wt.ResultDir, wt.Limits.Reverts)
	fmt.Fprintln(wt.Logger, "Done in", time.Now().Sub(start))

	return wt
}

// Process is the main procedure where the data process happen. In this method page will be cleaned by wikitext,
// will be performed tokenization, stopwords cleaning and stemming, files aggregation and then files de-stemming
func (wt builder) Process(ctx context.Context, fail func(error) error) (wtOut builder) {
	if ctx.Err() != nil {
		return wt
	}

	fmt.Fprintln(wt.Logger, "WikiMarkup and Stopwords cleaning;")
	start := time.Now()
	err := assets.Run(ctx, "textnormalizer", ".", map[string]string{"RESULTDIR": wt.ResultDir, "LANG": wt.Lang})
	if err != nil {
		fail(errors.WithStack(err))
		return wt
	}
	fmt.Fprintln(wt.Logger, "Done in", time.Now().Sub(start))

	fmt.Fprintln(wt.Logger, "Word mapping by page")
	start = time.Now()
	err = wordmapper.ByPage(wt.ResultDir)
	if err != nil {
		fail(err)
		return wt
	}
	fmt.Fprintln(wt.Logger, "Done in", time.Now().Sub(start))

	if ctx.Err() != nil {
		return wt
	}

	fmt.Fprintln(wt.Logger, "Processing GlobalWordMap file")
	start = time.Now()
	err = wordmapper.GlobalWordMapper(wt.ResultDir)
	if err != nil {
		fail(err)
		return wt
	}
	fmt.Fprintln(wt.Logger, "Done in", time.Now().Sub(start))

	fmt.Fprintln(wt.Logger, "Processing GlobalStem file")
	start = time.Now()
	err = wordmapper.StemRevAggregator(wt.ResultDir)
	if err != nil {
		fail(err)
		return wt
	}
	fmt.Fprintln(wt.Logger, "Done in", time.Now().Sub(start))

	if ctx.Err() != nil {
		return wt
	}

	fmt.Fprintln(wt.Logger, "Processing GlobalPage file")
	start = time.Now()
	err = wordmapper.PageMapAggregator(wt.ResultDir)
	if err != nil {
		fail(err)
		return wt
	}
	fmt.Fprintln(wt.Logger, "Done in", time.Now().Sub(start))

	fmt.Fprintln(wt.Logger, "Processing TFIDF file")
	start = time.Now()
	err = tfidf.ComputeTFIDF(wt.ResultDir)
	if err != nil {
		fail(err)
		return wt
	}
	fmt.Fprintln(wt.Logger, "Done in", time.Now().Sub(start))

	fmt.Fprintln(wt.Logger, "Performing Destemming")
	start = time.Now()
	err = assets.Run(ctx, "destemmer", ".", map[string]string{"RESULTDIR": wt.ResultDir})
	if err != nil {
		fail(errors.WithStack(err))
		return wt
	}
	fmt.Fprintln(wt.Logger, "Done in", time.Now().Sub(start))

	fmt.Fprintln(wt.Logger, "Processing topic words")
	start = time.Now()
	err = topicwords.TopicWords(wt.ResultDir)
	if err != nil {
		fail(err)
		return wt
	}
	fmt.Fprintln(wt.Logger, "Done in", time.Now().Sub(start))

	fmt.Fprintln(wt.Logger, "Processing Badwords report")
	start = time.Now()
	err = badwords.BadWords(wt.Lang, wt.ResultDir)
	if err != nil {
		fail(err)
		return wt
	}
	fmt.Fprintln(wt.Logger, "Done in", time.Now().Sub(start))

	fmt.Fprintln(wt.Logger, "Processing top N words")
	start = time.Now()
	err = assets.Run(ctx, "topwordspageextractor", ".", map[string]string{
		"RESULTDIR":        wt.ResultDir,
		"WORDSPAGELIMIT":   strconv.Itoa(wt.Limits.WordsPages),
		"GLOBALWORDSLIMIT": strconv.Itoa(wt.Limits.GlobalWords),
		"TOPICWORDSLIMIT":  strconv.Itoa(wt.Limits.TopicWords),
	})
	if err != nil {
		fail(errors.WithStack(err))
		return wt
	}
	fmt.Fprintln(wt.Logger, "Done in", time.Now().Sub(start))

	return wt
}

// CheckAvailableLanguage check if a language is handled
func CheckAvailableLanguage(lang string) error {
	languages := map[string]string{
		"en":     "english",
		"ar":     "arabic",
		"da":     "danish",
		"nl":     "dutch",
		"fi":     "finnish",
		"fr":     "french",
		"de":     "german",
		"el":     "greek",
		"hu":     "hungarian",
		"id":     "indonesian",
		"it":     "italian",
		"kk":     "kazakh",
		"ne":     "nepali",
		"no":     "norwegian",
		"pt":     "portuguese",
		"ro":     "romanian",
		"ru":     "russian",
		"es":     "spanish",
		"sv":     "swedish",
		"tr":     "turkish",
		"hy":     "armenian",
		"az":     "azerbaijani",
		"eu":     "basque",
		"bn":     "bengali",
		"bg":     "bulgarian",
		"ca":     "catalan",
		"zh":     "chinese",
		"sh":     "croatian",
		"cs":     "czech",
		"gl":     "galician",
		"he":     "hebrew",
		"hi":     "hindi",
		"ga":     "irish",
		"ja":     "japanese",
		"ko":     "korean",
		"lv":     "latvian",
		"lt":     "lithuanian",
		"mr":     "marathi",
		"fa":     "persian",
		"pl":     "polish",
		"sk":     "slovak",
		"th":     "thai",
		"uk":     "ukrainian",
		"ur":     "urdu",
		"simple": "english",
		"vec":    "italian", // only test
	}

	if lang == "" {
		return errors.New("Empty langugage")
	}

	if _, isIn := languages[lang]; !isIn {
		return errors.New(lang + " is not an available language")
	}
	return nil
}
