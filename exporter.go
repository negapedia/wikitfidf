package wikitfidf

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/negapedia/wikitfidf/internal/badwords"
	"github.com/negapedia/wikitfidf/internal/structures"

	"github.com/pkg/errors"
)

//Exporter represents the TFIDF data calculated from New.
type Exporter struct {
	ResultDir, Lang string
}

const (
	globalPagesTFIDFName  = "GlobalPagesTFIDF_topN.json.gz"
	globalTopicsWordsName = "GlobalTopicsWords_topN.json.gz"
	globalWordsName       = "GlobalWords_topN.json.gz"
	badWordsReportName    = "BadWordsReport.json.gz"
)

// From returns an exporter from existing data, it check if files that have to be exported exists.
// If not, returns an error with the specified missing file.
func From(lang, resultDir string) (exporter Exporter, err error) {
	err = CheckAvailableLanguage(lang)
	if err != nil {
		return
	}

	files := []string{globalPagesTFIDFName,
		globalTopicsWordsName,
		globalWordsName,
	}
	if _, ok := badwords.AvailableLanguage(lang); ok {
		files = append(files, badWordsReportName)
	}

	for _, file := range files {
		if _, err = os.Stat(filepath.Join(resultDir, file)); os.IsNotExist(err) {
			return
		}
	}

	return Exporter{resultDir, lang}, nil
}

// Delete deletes files from result directory
func (exporter Exporter) Delete() (err error) {
	files := []string{globalPagesTFIDFName,
		globalTopicsWordsName,
		globalWordsName,
	}
	if _, ok := badwords.AvailableLanguage(exporter.Lang); ok {
		files = append(files, badWordsReportName)
	}

	for _, file := range files {
		if currErr := os.Remove(filepath.Join(exporter.ResultDir, file)); currErr != nil {
			err = currErr
		}
	}
	return
}

// WikiWords represents the top N words in Wikipedia with the total number of words in it
type WikiWords struct {
	TotalWords  uint32
	Words2Occur map[string]uint32
}

// GlobalWords returns a dictionary with the top N words of GlobalWord in the following format: "word": occurencies
func (exporter Exporter) GlobalWords() (word2Occurencies *WikiWords, err error) {
	file, err := os.Open(filepath.Join(exporter.ResultDir, globalWordsName))
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to opening %v", globalWordsName)
	}
	defer file.Close()
	fileReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to create gzip reader for %v", globalWordsName)
	}
	defer fileReader.Close()

	byteValue, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to read %v", globalWordsName)
	}

	var globalWord map[string]uint32

	err = json.Unmarshal(byteValue, &globalWord)
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to unmarshal %v", globalWordsName)
	}

	totWords := globalWord["@TOTAL Words"]
	delete(globalWord, "@TOTAL Words")

	return &WikiWords{TotalWords: totWords, Words2Occur: globalWord}, nil
}

// PageTFIDF represents a single page with its data: ID, TopicID, Total number of words,
// dictionary with the top N words in the following format: "word": tfidf_value
type PageTFIDF struct {
	ID         uint32
	TotWords   uint32
	Word2TFIDF map[string]float64
}

// Pages returns a channel with the data of PageTFIDF (top N words per page), pages sent in channel are ascending order.
func (exporter Exporter) Pages(ctx context.Context, fail func(error) error) chan PageTFIDF {
	ch := make(chan PageTFIDF)

	globalPage, err := os.Open(filepath.Join(exporter.ResultDir, globalPagesTFIDFName))
	if err != nil {
		fail(errors.Wrapf(err, "Error happened while trying to open %v", globalPagesTFIDFName))
		close(ch)
		return ch
	}

	globalPageReader, err := gzip.NewReader(globalPage)
	if err != nil {
		fail(errors.Wrapf(err, "Error happened while trying to create gzip reader for %v", globalPagesTFIDFName))
		close(ch)
		return ch
	}
	lineReader := bufio.NewScanner(globalPageReader)

	go func() {
		defer close(ch)
		defer globalPage.Close()
		defer globalPageReader.Close()

		for lineReader.Scan() {
			line := lineReader.Text()

			if line == "}" {
				break
			}

			var page map[uint32]structures.PageTopNWords

			if line[:1] != "{" {
				line = "{" + line
			}

			line = line[:len(line)-1] + "}"

			if err = json.Unmarshal([]byte(line), &page); err != nil {
				fail(errors.Wrapf(err, "Error while unmarshalling json in %v \n\t at line: %v", globalPagesTFIDFName, line))
				return
			}
			for id := range page {
				select {
				case <-ctx.Done():
					return
				case ch <- PageTFIDF{ID: id, TotWords: page[id].TotWords, Word2TFIDF: page[id].Words}:
				}
			}
		}
	}()
	return ch
}

// Topic represents a single topic with TopicID and the list of top N words in it in
// the following format: "word": number_of_occurrence
type Topic struct {
	TopicID  uint32
	TotWords uint32
	Words    map[string]uint32
}

// Topics returns a channel with the data of GlobalTopic (top N words per topic)
func (exporter Exporter) Topics(ctx context.Context, fail func(error) error) chan Topic {
	ch := make(chan Topic)

	globalTopic, err := os.Open(filepath.Join(exporter.ResultDir, globalTopicsWordsName))
	if err != nil {
		fail(errors.Wrapf(err, "Error happened while trying to open %v", globalTopicsWordsName))
		close(ch)
		return ch
	}
	globalPageReader, err := gzip.NewReader(globalTopic)
	if err != nil {
		fail(errors.Wrapf(err, "Error happened while trying to create gzip reader for %v", globalTopicsWordsName))
		close(ch)
		return ch
	}
	lineReader := bufio.NewScanner(globalPageReader)

	go func() {
		defer close(ch)
		defer globalTopic.Close()
		defer globalPageReader.Close()

		for lineReader.Scan() {
			line := lineReader.Text()

			if line == "}" {
				break
			}

			var topic map[uint32]map[string]uint32 //ID: {words: {w: y, w2: z..., @Tot: k}

			if line[:1] != "{" {
				line = "{" + line
			}

			line = line[:len(line)-1] + "}"

			if err = json.Unmarshal([]byte(line), &topic); err != nil {
				fail(errors.Wrapf(err, "Error while unmarshalling json in %v : %v", globalTopicsWordsName, line))
				return
			}

			for topicID := range topic {
				totalWords := topic[topicID]["@TOT"]
				delete(topic[topicID], "@TOT")
				select {
				case <-ctx.Done():
					return
				case ch <- Topic{TopicID: topicID, TotWords: totalWords, Words: topic[topicID]}:
				}
			}

		}

	}()
	return ch
}

// BadWordsPage represents a single page with badwords data: PageID, TopicID, Absolute number of badwords in page,
// Relative number of badwords in page (tot/abs) and the list of the badwords in the following format: "badWord": number_of_occurrence
type BadWordsPage struct {
	PageID uint32
	Abs    uint32
	Rel    float64
	BadW   map[string]uint32
}

// PageBadwords returns a channel with the data of BadWords Report
// pages sent in channel are descending ordered
func (exporter Exporter) PageBadwords(ctx context.Context, fail func(error) error) chan BadWordsPage {
	ch := make(chan BadWordsPage)

	if _, exists := badwords.AvailableLanguage(exporter.Lang); !exists {
		close(ch)
		return ch
	}

	badWords, err := os.Open(filepath.Join(exporter.ResultDir, badWordsReportName))
	if err != nil {
		fail(errors.Wrapf(err, "Error happened while trying to open %v", badWordsReportName))
		close(ch)
		return ch
	}

	globalPageReader, err := gzip.NewReader(badWords)
	if err != nil {
		fail(errors.Wrapf(err, "Error happened while trying to create gzip reader for %v", badWordsReportName))
		close(ch)
		return ch
	}
	lineReader := bufio.NewScanner(globalPageReader)

	go func() {
		defer close(ch)
		defer badWords.Close()
		defer globalPageReader.Close()

		for lineReader.Scan() {
			line := lineReader.Text()

			if line == "}" {
				break
			}

			var page map[uint32]structures.BadWordsReport

			if line[:1] != "{" {
				line = "{" + line
			}

			line = line[:len(line)-1] + "}"

			if err = json.Unmarshal([]byte(line), &page); err != nil {
				fail(errors.Wrapf(err, "Error while unmarshalling json in %v", badWordsReportName))
				return
			}

			for id := range page {
				select {
				case <-ctx.Done():
					return
				case ch <- BadWordsPage{PageID: id, Abs: page[id].Abs, Rel: page[id].Rel, BadW: page[id].BadW}:
				}
			}
		}

	}()
	return ch
}
