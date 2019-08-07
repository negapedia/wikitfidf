package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
	Lang      string
	ResultDir string
	date      string

	Nrevert   int
	NTopWords int

	StartDate                 time.Time
	EndDate                   time.Time
	SpecialPageList           []string
	CompressAndRemoveFinalOut bool
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
		"vec":    "italian", // only test
	}

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
func (wd *WikiDumpConflitcAnalyzer) NewWikiDump(lang string, resultDir string,
	startDate string, endDate string, specialPageList string, nRevert, topNWords int, compress bool) {

	_, err := checkAvailableLanguage(lang)
	if err != nil {
		panic(err)
	}
	wd.Lang = lang

	wd.StartDate, _ = time.Parse(startDate, "2019-01-01T15:00")
	wd.EndDate, _ = time.Parse(endDate, "2019-01-01T15:00")

	wd.date = time.Now().Month().String() + strconv.Itoa(time.Now().Year())
	if !wd.StartDate.IsZero() || !wd.EndDate.IsZero() {
		wd.date += wd.StartDate.String() + "_" + wd.EndDate.String()
	}

	wd.ResultDir = func(resultDir string) string { // add last directory separator if not exists
		if resultDir[len(resultDir)-1:] != "/" {
			resultDir += "/"
		}
		return resultDir
	}(resultDir) + lang + "_" + wd.date

	wd.Nrevert = nRevert
	if nRevert != 0 {
		wd.ResultDir += "_last" + strconv.Itoa(nRevert)
	}
	wd.ResultDir += "/"

	wd.NTopWords = topNWords

	wd.SpecialPageList = func(specialPageList string) []string {
		if specialPageList == "" {
			return nil
		} else {
			return strings.Split(specialPageList, "-")
		}
	}(specialPageList)

	if _, err := os.Stat(wd.ResultDir + "Stem"); os.IsNotExist(err) {
		err = os.MkdirAll(wd.ResultDir+"Stem", 0755)
		if err != nil {
			panic(err)
		}
	}
	wd.CompressAndRemoveFinalOut = compress
}

// Preprocess given a wikibrief.EvolvingPage channel reduce the amount of information in pages and save them
func (wd *WikiDumpConflitcAnalyzer) Preprocess(channel <-chan wikibrief.EvolvingPage) {
	println("Parse and reduction start")
	start := time.Now()
	dumpreducer.DumpReducer(channel, wd.ResultDir, time.Time{}, time.Time{}, nil, wd.Nrevert) //("../103KB_test.7z", wd.ResultDir, wd.startDate, wd.endDate, wd.SpecialPageList)// //startDate and endDate must be in the same format of dump timestamp!
	fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	println("Parse and reduction end")
}

// Process is the main procedure where the data process happen. In this method page will be cleaned by wikitext,
// will be performed tokenization, stopwords cleaning and stemming, files aggregation and then files de-stemming
func (wd *WikiDumpConflitcAnalyzer) Process() {
	println("WikiMarkup cleaning start")
	start := time.Now()
	wikiMarkupClean := exec.Command("java", "-jar", "./textnormalizer/WikipediaMarkupCleaner.jar", wd.ResultDir)
	_ = wikiMarkupClean.Run()
	fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	println("WikiMarkup cleaning end")

	println("Stopwords cleaning and stemming start")
	start = time.Now()
	stopwordsCleanerStemming := exec.Command("python3", "./textnormalizer/runStopwClean.py", wd.ResultDir, wd.Lang)
	_ = stopwordsCleanerStemming.Run()
	fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	println("Stopwords cleaning and stemming end")

	println("Word mapping by page start")
	start = time.Now()
	wordmapper.WordMapperByPage(wd.ResultDir)
	fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	println("Word mapping by page end")

	println("Processing GlobalWordMap file start")
	start = time.Now()
	wordmapper.GlobalWordMapper(wd.ResultDir)
	fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	println("Processing GlobalWordMap file start")

	println("Processing GlobalStem file start")
	start = time.Now()
	wordmapper.StemRevAggregator(wd.ResultDir)
	fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	println("Processing GlobalStem file end")

	println("Processing GlobalPage file start")
	start = time.Now()
	wordmapper.PageMapAggregator(wd.ResultDir)
	fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	println("Processing GlobalPage file end")

	println("Processing TFIDF file start")
	start = time.Now()
	tfidf.ComputeTFIDF(wd.ResultDir)
	fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	println("Processing TFIDF file end")

	println("Performing Destemming start")
	start = time.Now()
	deStemming := exec.Command("python3", "./destemmer/runDeStemming.py", wd.ResultDir)
	_ = deStemming.Run()
	fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	println("Performing Destemming file end")

	println("Processing top N words per page start")
	start = time.Now()
	topNWordsPageExtractor := exec.Command("python3", "./topwordspageextractor/runTopNWordsPageExtractor.py", wd.ResultDir, strconv.Itoa(wd.NTopWords))
	_ = topNWordsPageExtractor.Run()
	fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	println("Processing top N words per page end")

	println("Processing topic words start")
	start = time.Now()
	topicwords.TopicWords(wd.ResultDir)
	fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	println("Processing topic words end")

	println("Processing Badwords report start")
	start = time.Now()
	badwords.BadWords(wd.Lang, wd.ResultDir)
	fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	println("Processing Badwords report end")
}

func (wd *WikiDumpConflitcAnalyzer) CompressResultDir(whereToSave string) {
	if wd.CompressAndRemoveFinalOut {
		println("Compressing ResultDir in 7z start")
		fileName := wd.Lang + "_" + wd.date
		if wd.Nrevert != 0 {
			fileName += "_last" + strconv.Itoa(wd.Nrevert)
		}

		start := time.Now()
		topNWordsPageExtractor := exec.Command("7z", "a", "-r", whereToSave+fileName, wd.ResultDir+"*")
		_ = topNWordsPageExtractor.Run()

		_ = os.RemoveAll(wd.ResultDir)

		fmt.Println("Duration: (min) ", time.Now().Sub(start).Minutes())
		println("Compressing ResultDir in 7z end")
	}
}

func main() {
	langFlag := flag.String("l", "", "Dump language")
	dirFlag := flag.String("d", "", "Result dir")
	startDateFlag := flag.String("s", "", "Revision starting date")
	endDateFlag := flag.String("e", "", "Revision ending date")
	specialPageListFlag := flag.String("specialList", "", "Special page list, page not in this list will be ignored. Input PageID like: id1-id2-...")
	nRevert := flag.Int("r", 0, "Number of revert limit")
	nTopWords := flag.Int("t", 0, "Number of top words to process")
	compressFinalOut := flag.Bool("compress", true, "If true compress in 7z and delete the final output folder")
	flag.Parse()

	wd := new(WikiDumpConflitcAnalyzer)

	wd.NewWikiDump(*langFlag, *dirFlag, *startDateFlag, *endDateFlag, *specialPageListFlag, *nRevert, *nTopWords, *compressFinalOut)

	ctx, fail := ctxutils.WithFail(context.Background())
	pageChannel := wikibrief.New(ctx, fail, wd.ResultDir, wd.Lang)
	wd.Preprocess(pageChannel)

	if err := fail(nil); err != nil {
		log.Fatal("%+v", err)
	}

	wd.Process()
	wd.CompressResultDir("/Result/")
}
