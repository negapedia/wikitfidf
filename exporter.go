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

// GlobalWords returns a dictionary with the top N words of GlobalWord in the following format: "word": occurencies
func (exporter Exporter) GlobalWords() (word2Occurencies map[string]uint32, err error) {
	file, err := os.Open(filepath.Join(exporter.ResultDir, globalTopicsWordsName))
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to opening %v", globalTopicsWordsName)
	}
	defer file.Close()
	fileReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to create gzip reader for %v", globalTopicsWordsName)
	}
	defer fileReader.Close()

	byteValue, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to read %v", globalTopicsWordsName)
	}

	var globalWord map[string]uint32

	err = json.Unmarshal(byteValue, &globalWord)
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to unmarshal %v", globalTopicsWordsName)
	}

	return globalWord, nil
}

// PageTFIF represents a single page with its data: ID, TopicID, Total number of words,
// dictionary with the top N words in the following format: "word": tfidf_value
type PageTFIF struct {
	ID      uint32
	TopicID uint32
	Tot     uint32
	Words   map[string]float64
}

// Pages returns a channel with the data of PageTFIF (top N words per page), pages sent in channel are ascending order.
func (exporter Exporter) Pages(ctx context.Context, fail func(error) error) chan PageTFIF {
	ch := make(chan PageTFIF)

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
	lineReader := bufio.NewReader(globalPageReader)

	go func() {
		defer close(ch)
		defer globalPage.Close()
		defer globalPageReader.Close()

		for {
			line, err := lineReader.ReadString('\n')
			if err != nil {
				fail(errors.Wrapf(err, "Error while reading line in %v", globalPagesTFIDFName))
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

			if err = json.Unmarshal([]byte(line), &page); err != nil {
				fail(errors.Wrapf(err, "Error while unmarshalling json in %v", globalPagesTFIDFName))
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
	lineReader := bufio.NewReader(globalPageReader)

	go func() {
		defer close(ch)
		defer globalTopic.Close()
		defer globalPageReader.Close()

		for {
			line, err := lineReader.ReadString('\n')
			if err != nil {
				fail(errors.Wrapf(err, "Error while reading line in %v", globalTopicsWordsName))
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

			if err = json.Unmarshal([]byte(line), &topic); err != nil {
				fail(errors.Wrapf(err, "Error while unmarshalling json in %v", globalTopicsWordsName))
				return
			}

			for topicID := range topic {
				select {
				case <-ctx.Done():
					return
				case ch <- Topic{TopicID: topicID, Words: topic[topicID]}:
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

// BadwordsReport returns a channel with the data of BadWords Report
// pages sent in channel are descending ordered
func (exporter Exporter) BadwordsReport(ctx context.Context, fail func(error) error) chan BadWordsPage {
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
	lineReader := bufio.NewReader(globalPageReader)

	go func() {
		defer close(ch)
		defer badWords.Close()
		defer globalPageReader.Close()

		for {
			line, err := lineReader.ReadString('\n')
			if err != nil {
				fail(errors.Wrapf(err, "Error while reading line in %v", badWordsReportName))
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

			if err = json.Unmarshal([]byte(line), &page); err != nil {
				fail(errors.Wrapf(err, "Error while unmarshalling json in %v", badWordsReportName))
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
