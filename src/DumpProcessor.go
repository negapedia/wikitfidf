package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/ebonetti/ctxutils"

	"./badwords"
	"./dumpreducer"
	"./tfidf"
	"./wordmapper"
	"github.com/negapedia/wikibrief"
	"github.com/pkg/errors"
)

// WikiDumpConflitcAnalyzer represent the main specific of desiderd Wikipedia dumps
// and some options for the elaboration process
type WikiDumpConflitcAnalyzer struct {
	lang string
	date string

	downloadDir     string
	resultDir       string
	specialPageList *[]uint32
	startDate       time.Time
	endDate         time.Time
	nRevert         int
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
func (wd *WikiDumpConflitcAnalyzer) NewWikiDump(lang string, resultDir string, specialPageList *[]uint32,
	startDate time.Time, endDate time.Time, nRevert int) {
	_, err := checkAvailableLanguage(lang)
	if err != nil {
		panic(err)
	}
	wd.lang = lang

	if startDate.IsZero() && endDate.IsZero() {
		wd.date = time.Now().Month().String() + strconv.Itoa(time.Now().Year())
	} else if startDate.IsZero() && !endDate.IsZero() {
		wd.date = startDate.String() + "-" + time.Now().Month().String() + strconv.Itoa(time.Now().Year())
	} else {
		wd.date = time.Now().Month().String() + strconv.Itoa(time.Now().Year()) + "-" + endDate.String()
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

// Preprocess given a wikibrief.EvolvingPage channel reduce the amount of information in pages and save them
func (wd *WikiDumpConflitcAnalyzer) Preprocess(channel <-chan wikibrief.EvolvingPage) {
	println("\nParse and reduction start")
	dumpreducer.DumpReducer(channel, wd.resultDir, wd.startDate, wd.endDate, wd.specialPageList, wd.nRevert) //("../103KB_test.7z", wd.resultDir, wd.startDate, wd.endDate, wd.specialPageList)// //startDate and endDate must be in the same format of dump timestamp!
	println("Parse and reduction end")
}

// Process is the main procedure where the data process happen. In this method page will be cleaned by wikitext,
// will be performed tokenization, stopwords cleaning and stemming, files aggregation and then files de-stemming
func (wd *WikiDumpConflitcAnalyzer) Process() {
	println("WikiMarkup cleaning start")
	wikiMarkupClean := exec.Command("java", "-jar", "./textnormalizer/WikipediaMarkupCleaner.jar", wd.resultDir)
	_ = wikiMarkupClean.Run()

	println("WikiMarkup cleaning end")

	println("Stopwords cleaning and stemming start")
	stopwordsCleanerStemming := exec.Command("python3", "./textnormalizer/runStopwClean.py", wd.resultDir, wd.lang)
	_ = stopwordsCleanerStemming.Run()
	println("Stopwords cleaning and stemming end")

	println("Word mapping by page start")
	wordmapper.WordMapperByPage(wd.resultDir)
	println("Word mapping by page end")

	println("Processing GlobalWordMap file start")
	wordmapper.GlobalWordMapper(wd.resultDir)
	println("Processing GlobalWordMap file start")

	println("Processing GlobalStem file start")
	wordmapper.StemRevAggregator(wd.resultDir)
	println("Processing GlobalStem file end")

	println("Processing GlobalPage file start")
	wordmapper.PageMapAggregator(wd.resultDir)
	println("Processing GlobalPage file end")

	if wd.specialPageList == nil {
		println("Processing TFIDF file start")
		tfidf.ComputeTFIDF(wd.resultDir)
		println("Processing TFIDF file end")
	}

	println("Performing Destemming start")
	deStemming := exec.Command("python3", "./destemmer/runDeStemming.py", wd.resultDir)
	_ = deStemming.Run()
	println("Performing Destemming file end")

	println("Processing Badwords report start")
	badwords.BadWords(wd.lang, wd.resultDir)
	println("Processing Badwords report end")
}

func main() {
	wd := new(WikiDumpConflitcAnalyzer)
	wd.NewWikiDump("vec", "/Users/marcochilese/Desktop/Tesi/NegapediaConflicutalWords/Result/", nil, time.Time{}, time.Time{}, 10)

	ctx, fail := ctxutils.WithFail(context.Background())
	pageChannel := wikibrief.New(ctx, fail, wd.resultDir, wd.lang)
	wd.Preprocess(pageChannel)

	if err := fail(nil); err != nil {
		log.Fatal("%+v", err)
	}

	wd.Process()
}
