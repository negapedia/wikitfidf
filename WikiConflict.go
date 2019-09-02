package wikiconflict

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/negapedia/wikiconflict/internals/badwords"

	"github.com/negapedia/wikiconflict/internals/topicwords"

	"github.com/negapedia/wikibrief"
	"github.com/negapedia/wikiconflict/internals/dumpreducer"
	"github.com/negapedia/wikiconflict/internals/structures"
	"github.com/negapedia/wikiconflict/internals/tfidf"
	"github.com/negapedia/wikiconflict/internals/utils"
	"github.com/negapedia/wikiconflict/internals/wordmapper"
	"github.com/pkg/errors"
)

type nTopWords struct {
	TopNWordsPages  int
	TopNGlobalWords int
	TopNTopicWords  int
}

// Wikiconflict represent the main specific of desiderd Wikipedia dumps
// and some options for the elaboration process
type Wikiconflict struct {
	Lang      string
	ResultDir string
	date      string

	Nrevert   int
	TopNWords nTopWords

	StartDate       time.Time
	EndDate         time.Time
	SpecialPageList []string
	Logger          io.Writer

	Error error
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

	if _, isIn := languages[lang]; !isIn {
		return errors.New(lang + " is not an available language!")
	}
	return nil
}

// New admits to initialize with parameters a Wikiconflict.
func New(lang string, resultDir string,
	startDate string, endDate string, specialPageList string,
	nRevert, topNWordsPages, topNGlobalWords, topNTopicWords int,
	verbouseMode bool) (*Wikiconflict, error) {

	if lang == "" {
		return nil, errors.New("Langugage not set")
	} else if topNWordsPages == 0 || topNGlobalWords == 0 || topNTopicWords == 0 {
		return nil, errors.New("Number of topwords to calculate are setted to 0")
	}

	err := CheckAvailableLanguage(lang)
	if err != nil {
		return nil, errors.New("Language required not available")
	}

	wd := new(Wikiconflict)
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

	wd.TopNWords = nTopWords{topNWordsPages, topNGlobalWords, topNTopicWords}

	wd.SpecialPageList = func(specialPageList string) []string {
		if specialPageList == "" {
			return nil
		}
		return strings.Split(specialPageList, "-")
	}(specialPageList)

	if verbouseMode {
		wd.Logger = os.Stdout
	} else {
		wd.Logger = ioutil.Discard
	}

	if _, err := os.Stat(wd.ResultDir + "Stem"); os.IsNotExist(err) {
		err = os.MkdirAll(wd.ResultDir+"Stem", 0700) //0755
		if err != nil {
			log.Fatal("Error happened while trying to create", wd.ResultDir, "and", wd.ResultDir+"Stem")
		}
	}

	return wd, nil
}

// Preprocess given a wikibrief.EvolvingPage channel reduce the amount of information in pages and save them
func (wd *Wikiconflict) Preprocess(channel <-chan wikibrief.EvolvingPage) {
	_, _ = fmt.Fprintln(wd.Logger, "Parse and reduction start")
	start := time.Now()
	dumpreducer.DumpReducer(channel, wd.ResultDir, time.Time{}, time.Time{}, nil, wd.Nrevert) //("../103KB_test.7z", wd.ResultDir, wd.startDate, wd.endDate, wd.SpecialPageList)// //startDate and endDate must be in the same format of dump timestamp!
	_, _ = fmt.Fprintln(wd.Logger, "Duration: (h) ", time.Now().Sub(start).Hours())
	_, _ = fmt.Fprintln(wd.Logger, "Parse and reduction end")
}

// Process is the main procedure where the data process happen. In this method page will be cleaned by wikitext,
// will be performed tokenization, stopwords cleaning and stemming, files aggregation and then files de-stemming
func (wd *Wikiconflict) Process() error {
	_, _ = fmt.Fprintln(wd.Logger, "WikiMarkup cleaning start")
	start := time.Now()
	wikiMarkupClean := exec.Command("java", "-jar", "./internals/textnormalizer/WikipediaMarkupCleaner.jar", wd.ResultDir)
	_ = wikiMarkupClean.Run()
	_, _ = fmt.Fprintln(wd.Logger, "Duration: (h) ", time.Now().Sub(start).Hours())
	_, _ = fmt.Fprintln(wd.Logger, "WikiMarkup cleaning end")

	_, _ = fmt.Fprintln(wd.Logger, "Stopwords cleaning and stemming start")
	start = time.Now()
	stopwordsCleanerStemming := exec.Command("python3", "./internals/textnormalizer/runStopwClean.py", wd.ResultDir, wd.Lang)
	_ = stopwordsCleanerStemming.Run()
	_, _ = fmt.Fprintln(wd.Logger, "Duration: (h) ", time.Now().Sub(start).Hours())
	_, _ = fmt.Fprintln(wd.Logger, "Stopwords cleaning and stemming end")

	_, _ = fmt.Fprintln(wd.Logger, "Word mapping by page start")
	start = time.Now()
	err := wordmapper.ByPage(wd.ResultDir)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(wd.Logger, "Duration: (h) ", time.Now().Sub(start).Hours())
	_, _ = fmt.Fprintln(wd.Logger, "Word mapping by page end")

	_, _ = fmt.Fprintln(wd.Logger, "Processing GlobalWordMap file start")
	start = time.Now()
	err = wordmapper.GlobalWordMapper(wd.ResultDir)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(wd.Logger, "Processing GlobalWordMap file start")
	_, _ = fmt.Fprintln(wd.Logger, "Duration: (h) ", time.Now().Sub(start).Hours())

	_, _ = fmt.Fprintln(wd.Logger, "Processing GlobalStem file start")
	start = time.Now()
	err = wordmapper.StemRevAggregator(wd.ResultDir)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(wd.Logger, "Duration: (h) ", time.Now().Sub(start).Hours())
	_, _ = fmt.Fprintln(wd.Logger, "Processing GlobalStem file end")

	_, _ = fmt.Fprintln(wd.Logger, "Processing GlobalPage file start")
	start = time.Now()
	err = wordmapper.PageMapAggregator(wd.ResultDir)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(wd.Logger, "Duration: (h) ", time.Now().Sub(start).Hours())
	_, _ = fmt.Fprintln(wd.Logger, "Processing GlobalPage file end")

	_, _ = fmt.Fprintln(wd.Logger, "Processing TFIDF file start")
	start = time.Now()
	err = tfidf.ComputeTFIDF(wd.ResultDir)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(wd.Logger, "Duration: (h) ", time.Now().Sub(start).Hours())
	_, _ = fmt.Fprintln(wd.Logger, "Processing TFIDF file end")

	_, _ = fmt.Fprintln(wd.Logger, "Performing Destemming start")
	start = time.Now()
	deStemming := exec.Command("python3", "./internals/destemmer/runDeStemming.py", wd.ResultDir)
	_ = deStemming.Run()
	_, _ = fmt.Fprintln(wd.Logger, "Duration: (h) ", time.Now().Sub(start).Hours())
	_, _ = fmt.Fprintln(wd.Logger, "Performing Destemming file end")

	_, _ = fmt.Fprintln(wd.Logger, "Processing topic words start")
	start = time.Now()
	err = topicwords.TopicWords(wd.ResultDir)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(wd.Logger, "Duration: (h) ", time.Now().Sub(start).Hours())
	_, _ = fmt.Fprintln(wd.Logger, "Processing topic words end")

	_, _ = fmt.Fprintln(wd.Logger, "Processing Badwords report start")
	start = time.Now()
	err = badwords.BadWords(wd.Lang, wd.ResultDir)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(wd.Logger, "Duration: (h) ", time.Now().Sub(start).Hours())
	_, _ = fmt.Fprintln(wd.Logger, "Processing Badwords report end")

	_, _ = fmt.Fprintln(wd.Logger, "Processing top N words start")
	start = time.Now()
	topNWordsPageExtractor := exec.Command("python3", "./internals/topwordspageextractor/runTopNWordsPageExtractor.py", wd.ResultDir,
		strconv.Itoa(wd.TopNWords.TopNWordsPages), strconv.Itoa(wd.TopNWords.TopNWordsPages), strconv.Itoa(wd.TopNWords.TopNTopicWords))
	_ = topNWordsPageExtractor.Run()
	_, _ = fmt.Fprintln(wd.Logger, "Duration: (h) ", time.Now().Sub(start).Hours())
	_, _ = fmt.Fprintln(wd.Logger, "Processing top N words end")

	return nil
}

// CheckErrors check if errors happened during export process
func (wd *Wikiconflict) CheckErrors() {
	if wd.Error != nil {
		log.Fatal(wd.Error)
	}
}

// Clean delete files from result directory
func (wd *Wikiconflict) Clean() error {
	err := os.RemoveAll(wd.ResultDir)
	if err != nil {
		return errors.Wrap(err, "Error happened while trying to delete result dir")
	}
	return nil
}

// GlobalWordsExporter returns a channel with the data of GlobalWord (top N words)
func (wd *Wikiconflict) GlobalWordsExporter() map[string]uint32 {
	if wd.Error != nil {
		return nil
	}
	globalWord, err := utils.GetGlobalWordsTopN(wd.ResultDir, wd.TopNWords.TopNGlobalWords)
	if err != nil {
		wd.Error = errors.Wrap(err, "Errors happened while handling GlobalWords file")
		return nil
	}

	return globalWord
}

// PageTFIF represents a single page with its data: ID, TopicID, Total number of words,
// dictionary with the top N words in the following format: "word": tfidf_value
type PageTFIF struct {
	ID      uint32
	TopicID uint32
	Tot     uint32
	Words   map[string]float64
}

// GlobalPagesExporter returns a channel with the data of GlobalPagesTFIDF (top N words per page)
func (wd *Wikiconflict) PagesExporter(ctx context.Context) chan PageTFIF {
	ch := make(chan PageTFIF)
	if wd.Error != nil {
		close(ch)
		return ch
	}

	filename := wd.ResultDir + "GlobalPagesTFIDF_top" + strconv.Itoa(wd.TopNWords.TopNWordsPages) + ".json.gz"
	globalPage, err := os.Open(filename)
	if err != nil {
		wd.Error = errors.Wrap(err, "Error happened while trying to open GlobalPages.json file:GlobalPages.json")
		return nil
	}

	globalPageReader, err := gzip.NewReader(globalPage)
	if err != nil {
		wd.Error = errors.Wrap(err, "Error happened while trying to create gzip reader")
		close(ch)
		return ch
	}
	lineReader := bufio.NewReader(globalPageReader)

	go func() {
		defer close(ch)
		defer globalPage.Close()
		defer globalPageReader.Close()

		for {
			line, err := lineReader.ReadString('\n')
			if err != nil {
				wd.Error = errors.Wrapf(err, "Error while reading line")
				return
			}
			if line == "}" {
				break
			}

			var page map[uint32]structures.TfidfTopNWordPage

			if line[:1] != "{" {
				line = "{" + line
			}

			line = line[:len(line)-2] + "}"
			err = json.Unmarshal([]byte(line), &page)
			if err != nil {
				wd.Error = errors.Wrapf(err, "Error while unmarshalling json.")
				return
			}
			for id := range page {
				select {
				case <-ctx.Done():
					return
				case ch <- PageTFIF{ID: id, TopicID: page[id].TopicID, Tot: page[id].Tot, Words: *page[id].Words}:
				}
			}
		}

	}()
	return ch
}

// Topic represents a single topic with TopicID and the list of top N words in it in
// the following format: "word": number_of_occurrence
type Topic struct {
	TopicID uint32
	Words   map[string]uint32
}

// TopicsExporter returns a channel with the data of GlobalTopic (top N words per topic)
func (wd *Wikiconflict) TopicsExporter(ctx context.Context) chan Topic {
	ch := make(chan Topic)

	if wd.Error != nil {
		close(ch)
		return ch
	}

	filename := wd.ResultDir + "GlobalTopicsWords_top" + strconv.Itoa(wd.TopNWords.TopNTopicWords) + ".json.gz"
	globalTopic, err := os.Open(filename)

	if err != nil {
		wd.Error = errors.Wrapf(err, "Error happened while trying to open GlobalTopics_top.json ")
		close(ch)
		return ch
	}
	globalPageReader, err := gzip.NewReader(globalTopic)
	if err != nil {
		wd.Error = errors.Wrap(err, "Error happened while trying to create gzip reader")
		close(ch)
		return ch
	}
	lineReader := bufio.NewReader(globalPageReader)

	go func() {
		defer close(ch)
		defer globalTopic.Close()
		defer globalPageReader.Close()

		for {
			line, err := lineReader.ReadString('\n')
			if err != nil {
				wd.Error = errors.Wrapf(err, "Error while reading line")
				return
			}
			if err != nil {
				break
			}
			if line == "}" {
				break
			}

			var topic map[uint32]map[string]uint32

			if line[:1] != "{" {
				line = "{" + line
			}

			line = line[:len(line)-2] + "}"
			err = json.Unmarshal([]byte(line), &topic)
			if err != nil {
				wd.Error = errors.Wrapf(err, "Error while unmarshalling json.")
				return
			}

			for topicId := range topic {
				select {
				case <-ctx.Done():
					return
				case ch <- Topic{TopicID: topicId, Words: topic[topicId]}:
				}
			}

		}

	}()
	return ch
}

// BadWordsPage represents a single page with badwords data: PageID, TopicID, Absolute number of badwords in page,
// Relative number of badwords in page (tot/abs) and the list of the badwords in the following format: "badWord": number_of_occurrence
type BadWordsPage struct {
	PageID  uint32
	TopicID uint32
	Abs     uint32
	Rel     float64
	BadW    map[string]int
}

// BadwordsReportExporter returns a channel with the data of BadWords Report
func (wd *Wikiconflict) BadwordsReportExporter(ctx context.Context) chan BadWordsPage {
	ch := make(chan BadWordsPage)

	if wd.Error != nil {
		close(ch)
		return ch
	}

	filename := wd.ResultDir + "BadWordsReport.json.gz"
	globalTopic, err := os.Open(filename)

	if err != nil {
		wd.Error = errors.Wrap(err, "Error happened while trying to open BadWordsReport.json ")
		close(ch)
		return ch
	}
	globalPageReader, err := gzip.NewReader(globalTopic)
	if err != nil {
		wd.Error = errors.Wrap(err, "Error happened while trying to create gzip reader")
		close(ch)
		return ch
	}
	lineReader := bufio.NewReader(globalPageReader)

	go func() {
		defer close(ch)
		defer globalTopic.Close()
		defer globalPageReader.Close()

		for {
			line, err := lineReader.ReadString('\n')
			if err != nil {
				wd.Error = errors.Wrapf(err, "Error while reading line")
				return
			}
			if err != nil {
				break
			}
			if line == "}" {
				break
			}

			var page map[uint32]structures.BadWordsReport

			if line[:1] != "{" {
				line = "{" + line
			}

			line = line[:len(line)-2] + "}"
			err = json.Unmarshal([]byte(line), &page)
			if err != nil {
				wd.Error = errors.Wrapf(err, "Error while unmarshalling json.")
				return
			}

			for id := range page {
				select {
				case <-ctx.Done():
					return
				case ch <- BadWordsPage{PageID: id, TopicID: page[id].TopicID,
					Abs: page[id].Abs, Rel: page[id].Rel, BadW: page[id].BadW}:
				}
			}
		}

	}()
	return ch
}
