package wikidumpconflictanalyzer

import (
	"fmt"
	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/badwords"
	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/dumpreducer"
	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/tfidf"
	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/topicwords"
	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/wordmapper"
	"github.com/negapedia/wikibrief"
	"github.com/pkg/errors"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
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
	VerbouseMode 			  bool
}

func CheckAvailableLanguage(lang string) (bool, error) {
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
	startDate string, endDate string, specialPageList string, nRevert, topNWords int, compress bool, verbouseMode bool) {

	_, err := CheckAvailableLanguage(lang)
	if err != nil {
		log.Fatal(err)
	}
	wd.Lang = lang

	wd.StartDate, _ = time.Parse(startDate, "2019-01-01T15:00")
	wd.EndDate, _ = time.Parse(endDate, "2019-01-01T15:00")

	wd.date = time.Now().Month().String() + strconv.Itoa(time.Now().Year())
	if !wd.StartDate.IsZero() || !wd.EndDate.IsZero() {
		wd.date += wd.StartDate.String() + "_" + wd.EndDate.String()
	}

	wd.ResultDir = func(resultDir string) string { // assign default result dir if not setted, and add last directory separator if not exists
		if resultDir == "" {
			resultDir = "/Results/"
		} else if resultDir[len(resultDir)-1:] != "/" {
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
		}
		return strings.Split(specialPageList, "-")
	}(specialPageList)

	if _, err := os.Stat(wd.ResultDir + "Stem"); os.IsNotExist(err) {
		err = os.MkdirAll(wd.ResultDir+"Stem", 0755)
		if err != nil {
			log.Fatal("Error happened while trying to create", wd.ResultDir, "and", wd.ResultDir+"Stem")
		}
	}
	wd.CompressAndRemoveFinalOut = compress
	wd.VerbouseMode = verbouseMode
}

// Preprocess given a wikibrief.EvolvingPage channel reduce the amount of information in pages and save them
func (wd *WikiDumpConflitcAnalyzer) Preprocess(channel <-chan wikibrief.EvolvingPage) {
	if wd.VerbouseMode{
		fmt.Println("Parse and reduction start")
	}
	start := time.Now()
	dumpreducer.DumpReducer(channel, wd.ResultDir, time.Time{}, time.Time{}, nil, wd.Nrevert) //("../103KB_test.7z", wd.ResultDir, wd.startDate, wd.endDate, wd.SpecialPageList)// //startDate and endDate must be in the same format of dump timestamp!
	if wd.VerbouseMode {
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Parse and reduction end")
	}
}

// Process is the main procedure where the data process happen. In this method page will be cleaned by wikitext,
// will be performed tokenization, stopwords cleaning and stemming, files aggregation and then files de-stemming
func (wd *WikiDumpConflitcAnalyzer) Process() {
	if wd.VerbouseMode{
		fmt.Println("WikiMarkup cleaning start")
	}
	start := time.Now()
	wikiMarkupClean := exec.Command("java", "-jar", "./internals/textnormalizer/WikipediaMarkupCleaner.jar", wd.ResultDir)
	_ = wikiMarkupClean.Run()
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("WikiMarkup cleaning end")
	}

	if wd.VerbouseMode{
		fmt.Println("Stopwords cleaning and stemming start")
	}
	start = time.Now()
	stopwordsCleanerStemming := exec.Command("python3", "./internals/textnormalizer/runStopwClean.py", wd.ResultDir, wd.Lang)
	_ = stopwordsCleanerStemming.Run()
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Stopwords cleaning and stemming end")
	}

	if wd.VerbouseMode{
		fmt.Println("Word mapping by page start")
	}
	start = time.Now()
	wordmapper.WordMapperByPage(wd.ResultDir)
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Word mapping by page end")
	}

	if wd.VerbouseMode{
		fmt.Println("Processing GlobalWordMap file start")
	}
	start = time.Now()
	wordmapper.GlobalWordMapper(wd.ResultDir)
	if wd.VerbouseMode{
		fmt.Println("Processing GlobalWordMap file start")
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	}

	if wd.VerbouseMode{
		fmt.Println("Processing GlobalStem file start")
	}
	start = time.Now()
	wordmapper.StemRevAggregator(wd.ResultDir)
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Processing GlobalStem file end")
	}

	if wd.VerbouseMode{
		fmt.Println("Processing GlobalPage file start")
	}
	start = time.Now()
	wordmapper.PageMapAggregator(wd.ResultDir)
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Processing GlobalPage file end")
	}

	if wd.VerbouseMode{
		fmt.Println("Processing TFIDF file start")
	}
	start = time.Now()
	tfidf.ComputeTFIDF(wd.ResultDir)
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Processing TFIDF file end")
	}

	if wd.VerbouseMode{
		fmt.Println("Performing Destemming start")
	}
	start = time.Now()
	deStemming := exec.Command("python3", "./internals/destemmer/runDeStemming.py", wd.ResultDir)
	_ = deStemming.Run()
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Performing Destemming file end")
	}

	if wd.VerbouseMode{
		fmt.Println("Processing top N words per page start")
	}
	start = time.Now()
	topNWordsPageExtractor := exec.Command("python3", "./internals/topwordspageextractor/runTopNWordsPageExtractor.py", wd.ResultDir, strconv.Itoa(wd.NTopWords))
	_ = topNWordsPageExtractor.Run()
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Processing top N words per page end")
	}

	if wd.VerbouseMode{
		fmt.Println("Processing topic words start")
	}
	start = time.Now()
	topicwords.TopicWords(wd.ResultDir)
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Processing topic words end")
	}

	if wd.VerbouseMode{
		fmt.Println("Processing Badwords report start")
	}
	start = time.Now()
	badwords.BadWords(wd.Lang, wd.ResultDir)
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Processing Badwords report end")
	}
}

func (wd *WikiDumpConflitcAnalyzer) CompressResultDir(whereToSave string) {
	if wd.CompressAndRemoveFinalOut {
		if wd.VerbouseMode{
			fmt.Println("Compressing ResultDir in 7z start")
		}
		fileName := wd.Lang + "_" + wd.date
		if wd.Nrevert != 0 {
			fileName += "_last" + strconv.Itoa(wd.Nrevert)
		}

		start := time.Now()
		topNWordsPageExtractor := exec.Command("7z", "a", "-r", whereToSave+fileName, wd.ResultDir+"*")
		_ = topNWordsPageExtractor.Run()

		_ = os.RemoveAll(wd.ResultDir)

		if wd.VerbouseMode{
			fmt.Println("Duration: (min) ", time.Now().Sub(start).Minutes())
			fmt.Println("Compressing ResultDir in 7z end")
		}
	}
}
