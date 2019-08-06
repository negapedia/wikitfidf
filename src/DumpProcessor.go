package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/ebonetti/ctxutils"

	"./badwords"
	"./dumpreducer"
	"./tfidf"
	"./topicwords"
	"./wordmapper"
	"github.com/negapedia/wikibrief"
	"github.com/pkg/errors"
)

// WikiDumpConflitcAnalyzer represent the main specific of desiderd Wikipedia dumps
// and some options for the elaboration process
type WikiDumpConflitcAnalyzer struct {
	lang string
	date string

	downloadDir string
	resultDir   string
	nRevert     int
	topNWords   int
}

func checkAvailableLanguage(lang string) (bool, error) {
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
		"vec":    "italian"}

	var noLang error
	if _, isIn := languages[lang]; !isIn {
		noLang = errors.New(lang + " is not an available language!")
	}
	return true, noLang
}

// NewWikiDump admits to initialize with parameters a WikiDumpConflitcAnalyzer. Parameters are about
// required Wikipedia Dump language, result directory, special page list which admits to process only the page in list,
// start and end date which admits to work only in a specific time frame, number of revert to consider: will be processed
// only the last "n" revert per page
func (wd *WikiDumpConflitcAnalyzer) NewWikiDump(lang string, resultDir string, nRevert, topNWords int) {
	_, err := checkAvailableLanguage(lang)
	if err != nil {
		panic(err)
	}
	wd.lang = lang

	wd.date = time.Now().Month().String() + strconv.Itoa(time.Now().Year())

	wd.resultDir = resultDir + lang + "_" + wd.date
	wd.nRevert = nRevert
	if nRevert != 0 {
		wd.resultDir += "_last" + strconv.Itoa(nRevert)
	}
	wd.resultDir += "/"

	wd.topNWords = topNWords

	if _, err := os.Stat(wd.resultDir + "Stem"); os.IsNotExist(err) {
		err = os.MkdirAll(wd.resultDir+"Stem", 0755)
		if err != nil {
			panic(err)
		}
	}
}

// Preprocess given a wikibrief.EvolvingPage channel reduce the amount of information in pages and save them
func (wd *WikiDumpConflitcAnalyzer) Preprocess(channel <-chan wikibrief.EvolvingPage) {
	println("Parse and reduction start")
	start := time.Now()
	dumpreducer.DumpReducer(channel, wd.resultDir, time.Time{}, time.Time{}, nil, wd.nRevert) //("../103KB_test.7z", wd.resultDir, wd.startDate, wd.endDate, wd.specialPageList)// //startDate and endDate must be in the same format of dump timestamp!
	fmt.Println("Duration: (s) ", time.Now().Sub(start).Seconds())
	println("Parse and reduction end")
}

// Process is the main procedure where the data process happen. In this method page will be cleaned by wikitext,
// will be performed tokenization, stopwords cleaning and stemming, files aggregation and then files de-stemming
func (wd *WikiDumpConflitcAnalyzer) Process() {
	println("WikiMarkup cleaning start")
	start := time.Now()
	wikiMarkupClean := exec.Command("java", "-jar", "./textnormalizer/WikipediaMarkupCleaner.jar", wd.resultDir)
	_ = wikiMarkupClean.Run()
	fmt.Println("Duration: (s) ", time.Now().Sub(start).Seconds())
	println("WikiMarkup cleaning end")

	println("Stopwords cleaning and stemming start")
	start = time.Now()
	stopwordsCleanerStemming := exec.Command("python3", "./textnormalizer/runStopwClean.py", wd.resultDir, wd.lang)
	_ = stopwordsCleanerStemming.Run()
	fmt.Println("Duration: (s) ", time.Now().Sub(start).Seconds())
	println("Stopwords cleaning and stemming end")

	println("Word mapping by page start")
	start = time.Now()
	wordmapper.WordMapperByPage(wd.resultDir)
	fmt.Println("Duration: (s) ", time.Now().Sub(start).Seconds())
	println("Word mapping by page end")

	println("Processing GlobalWordMap file start")
	start = time.Now()
	wordmapper.GlobalWordMapper(wd.resultDir)
	fmt.Println("Duration: (s) ", time.Now().Sub(start).Seconds())
	println("Processing GlobalWordMap file start")

	println("Processing GlobalStem file start")
	start = time.Now()
	wordmapper.StemRevAggregator(wd.resultDir)
	fmt.Println("Duration: (s) ", time.Now().Sub(start).Seconds())
	println("Processing GlobalStem file end")

	println("Processing GlobalPage file start")
	start = time.Now()
	wordmapper.PageMapAggregator(wd.resultDir)
	fmt.Println("Duration: (s) ", time.Now().Sub(start).Seconds())
	println("Processing GlobalPage file end")

	println("Processing TFIDF file start")
	start = time.Now()
	tfidf.ComputeTFIDF(wd.resultDir)
	fmt.Println("Duration: (s) ", time.Now().Sub(start).Seconds())
	println("Processing TFIDF file end")

	println("Performing Destemming start")
	start = time.Now()
	deStemming := exec.Command("python3", "./destemmer/runDeStemming.py", wd.resultDir)
	_ = deStemming.Run()
	fmt.Println("Duration: (s) ", time.Now().Sub(start).Seconds())
	println("Performing Destemming file end")

	println("Processing top N words per page start")
	start = time.Now()
	topNWordsPageExtractor := exec.Command("python3", "./topwordspageextractor/runTopNWordsPageExtractor.py", wd.resultDir, strconv.Itoa(wd.topNWords))
	_ = topNWordsPageExtractor.Run()
	fmt.Println("Duration: (s) ", time.Now().Sub(start).Seconds())
	println("Processing top N words per page end")

	println("Processing topic words start")
	start = time.Now()
	topicwords.TopicWords(wd.resultDir)
	fmt.Println("Duration: (s) ", time.Now().Sub(start).Seconds())
	println("Processing topic words end")

	println("Processing Badwords report start")
	start = time.Now()
	badwords.BadWords(wd.lang, wd.resultDir)
	fmt.Println("Duration: (s) ", time.Now().Sub(start).Seconds())
	println("Processing Badwords report end")
}

func (wd *WikiDumpConflitcAnalyzer) CompressResultDir(whereToSave string, removeResultDir bool) {
	println("Compressing ResultDir in 7z start")
	fileName := wd.lang + "_" + wd.date
	if wd.nRevert != 0 {
		fileName += "_last" + strconv.Itoa(wd.nRevert)
	}

	start := time.Now()
	topNWordsPageExtractor := exec.Command("7z", "a", "-r", whereToSave+fileName, wd.resultDir+"*")
	_ = topNWordsPageExtractor.Run()

	if removeResultDir {
		_ = os.RemoveAll(wd.resultDir)
	}

	fmt.Println("Duration: (s) ", time.Now().Sub(start).Seconds())
	println("Compressing ResultDir in 7z end")
}

func main() {
	wd := new(WikiDumpConflitcAnalyzer)

	nRevert, _ := strconv.Atoi(os.Args[3])
	nTopWords, _ := strconv.Atoi(os.Args[4])

	wd.NewWikiDump(os.Args[1], os.Args[2], nRevert, nTopWords)

	ctx, fail := ctxutils.WithFail(context.Background())
	pageChannel := wikibrief.New(ctx, fail, wd.resultDir, wd.lang)
	wd.Preprocess(pageChannel)

	if err := fail(nil); err != nil {
		log.Fatal("%+v", err)
	}

	wd.Process()
	wd.CompressResultDir("/Result/", true)
}
