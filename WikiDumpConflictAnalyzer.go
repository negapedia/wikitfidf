package WikiConflict

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/badwords"
	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/dumpreducer"
	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/structures"
	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/tfidf"
	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/topicwords"
	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/utils"
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

type nTopWords struct{
	TopNWordsPages int
	TopNGlobalWords int
	TopNTopicWords int
}

// WikiDumpConflitcAnalyzer represent the main specific of desiderd Wikipedia dumps
// and some options for the elaboration process
type WikiDumpConflitcAnalyzer struct {
	Lang      string
	ResultDir string
	date      string

	Nrevert   int
	TopNWords nTopWords

	StartDate                 time.Time
	EndDate                   time.Time
	SpecialPageList           []string
	CompressAndRemoveFinalOut bool
	VerbouseMode 			  bool

	Error	error
}

func New() (*WikiDumpConflitcAnalyzer, error) { //TODO
	return new(WikiDumpConflitcAnalyzer), nil
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
	startDate string, endDate string, specialPageList string, nRevert, topNWordsPages, topNGlobalWords, topNTopicWords int, compress bool, verbouseMode bool) {

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

	wd.TopNWords.TopNWordsPages = topNWordsPages
	wd.TopNWords.TopNGlobalWords = topNGlobalWords
	wd.TopNWords.TopNTopicWords = topNTopicWords

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
func (wd *WikiDumpConflitcAnalyzer) Process() error {
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
	err := wordmapper.WordMapperByPage(wd.ResultDir)
	if err != nil{
		return err
	}
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Word mapping by page end")
	}

	if wd.VerbouseMode{
		fmt.Println("Processing GlobalWordMap file start")
	}
	start = time.Now()
	err = wordmapper.GlobalWordMapper(wd.ResultDir)
	if err != nil{
		return err
	}
	if wd.VerbouseMode{
		fmt.Println("Processing GlobalWordMap file start")
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
	}

	if wd.VerbouseMode{
		fmt.Println("Processing GlobalStem file start")
	}
	start = time.Now()
	err = wordmapper.StemRevAggregator(wd.ResultDir)
	if err != nil{
		return err
	}
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Processing GlobalStem file end")
	}

	if wd.VerbouseMode{
		fmt.Println("Processing GlobalPage file start")
	}
	start = time.Now()
	err = wordmapper.PageMapAggregator(wd.ResultDir)
	if err != nil{
		return err
	}
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Processing GlobalPage file end")
	}

	if wd.VerbouseMode{
		fmt.Println("Processing TFIDF file start")
	}
	start = time.Now()
	err = tfidf.ComputeTFIDF(wd.ResultDir)
	if err != nil{
		return err
	}
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
	topNWordsPageExtractor := exec.Command("python3", "./internals/topwordspageextractor/runTopNWordsPageExtractor.py", wd.ResultDir, strconv.Itoa(wd.TopNWords.TopNWordsPages))
	_ = topNWordsPageExtractor.Run()
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Processing top N words per page end")
	}

	if wd.VerbouseMode{
		fmt.Println("Processing topic words start")
	}
	start = time.Now()
	err = topicwords.TopicWords(wd.ResultDir)
	if err != nil{
		return err
	}
	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Processing topic words end")
	}

	if wd.VerbouseMode{
		fmt.Println("Processing Badwords report start")
	}
	start = time.Now()
	err = badwords.BadWords(wd.Lang, wd.ResultDir)
	if err != nil{
		return err
	}

	if wd.VerbouseMode{
		fmt.Println("Duration: (h) ", time.Now().Sub(start).Hours())
		fmt.Println("Processing Badwords report end")
	}
	return nil
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

func (wd *WikiDumpConflitcAnalyzer) CheckErrors() {
	if wd.Error != nil{
		log.Fatal(wd.Error)
	}
}

func (wd *WikiDumpConflitcAnalyzer) GlobalWordExporter(ctx context.Context) chan map[string]uint32 {
	ch := make(chan map[string]uint32)

	globalWord, err := utils.GetGlobalWordsTopN(wd.ResultDir, wd.TopNWords.TopNGlobalWords)
	if err != nil {
		wd.Error = errors.Wrap(err, "")
		return nil
	}

	go func(){
		defer close(ch)

		for key, value := range globalWord{
			if ctx != nil {
				select {
				case <-ctx.Done():
					return
				default:

				}
			}
			x := make(map[string]uint32)
			x[key] = value
			ch <- x
		}
	}()
	return ch
}

func (wd *WikiDumpConflitcAnalyzer) GlobalPagesExporter(ctx context.Context) chan map[string]structures.TfidfTopNWordPage {
	ch := make(chan map[string]structures.TfidfTopNWordPage)

	globalPage, err := os.Open(wd.ResultDir+"GlobalPagesTFIDF_top"+strconv.Itoa(wd.TopNWords.TopNWordsPages)+".json")

	if err != nil {
		log.Fatal("Error happened while trying to open GlobalPages.json file:GlobalPages.json", err)
	}
	globalPageReader := bufio.NewReader(globalPage)

	go func(){
		defer close(ch)
		defer globalPage.Close()

		for{
			line, err := globalPageReader.ReadString('\n')
			println(line)
			if err != nil {
				break
			}
			if line == "}" {
				break
			}

			var page map[string]structures.TfidfTopNWordPage

			if line[:1] != "{" {
				line = "{" + line
			}

			line = line[:len(line)-2] + "}"
			err = json.Unmarshal([]byte(line), &page)
			if err != nil {
				wd.Error = errors.Wrapf(err, "Error while unmarshalling json.")
				return
			}
			if ctx != nil {
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
			ch <- page
		}

	}()
	return ch
}

func (wd *WikiDumpConflitcAnalyzer) GlobalTopicsExporter(ctx context.Context) chan map[string]map[string]uint32 {
	ch := make(chan map[string]map[string]uint32)

	globalTopic, err := os.Open(wd.ResultDir+"GlobalTopics_top"+strconv.Itoa(wd.TopNWords.TopNWordsPages)+".json")

	if err != nil {
		log.Fatal("Error happened while trying to open GlobalTopics_top.json ", err)
	}
	globalPageReader := bufio.NewReader(globalTopic)

	go func(){
		defer close(ch)
		defer globalTopic.Close()

		for{
			line, err := globalPageReader.ReadString('\n')
			println(line)
			if err != nil {
				break
			}
			if line == "}" {
				break
			}

			var topic map[string]map[string]uint32

			if line[:1] != "{" {
				line = "{" + line
			}

			line = line[:len(line)-2] + "}"
			err = json.Unmarshal([]byte(line), &topic)
			if err != nil {
				wd.Error = errors.Wrapf(err, "Error while unmarshalling json.")
				return
			}
			if ctx != nil {
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
			ch <- topic
		}

	}()
	return ch
}

