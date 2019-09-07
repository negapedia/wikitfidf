/*
 * Developed by Marco Chilese.
 * Last modified 04/08/2019, 11:49
 *
 */

package topicwords

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/negapedia/wikitfidf/internal/structures"
	"github.com/negapedia/wikitfidf/internal/utils"
)

type topicWriter struct {
	Writer *bufio.Writer
	File   *os.File
}

func writeWord(topicWriters map[uint32]*topicWriter, resultDir string, topicID uint32, word string) (err error) {
	if _, ok := topicWriters[topicID]; !ok {
		outFile, err := os.Create(filepath.Join(resultDir, fmt.Sprint("T", topicID)))
		if err != nil {
			return errors.WithStack(err)
		}
		topicWriters[topicID] = &topicWriter{Writer: bufio.NewWriter(outFile), File: outFile}
	}

	_, err = topicWriters[topicID].Writer.Write([]byte(word + "\n"))
	if err != nil {
		return errors.WithStack(err)
	}

	err = topicWriters[topicID].Writer.Flush()
	if err != nil {
		return errors.WithStack(err)
	}
	return
}

func topicWordsWriter(resultDir string) (err error) {
	globalPageTFIDF, err := os.Open(filepath.Join(resultDir, "GlobalPagesTFIDF.json"))
	defer func() {
		if e := globalPageTFIDF.Close(); e != nil && err == nil {
			err = errors.Wrapf(e, "Error while closing file %v", globalPageTFIDF.Name())
		}
	}()

	if err != nil {
		return errors.Wrapf(err, "Error happened while trying to open GlobalPagesTFIDF.json file:"+resultDir+"GlobalPagesTFIDF.json")
	}
	globalPageReader := bufio.NewReader(globalPageTFIDF)

	topicWordWriters := make(map[uint32]*topicWriter)
	defer func() {
		for _, writer := range topicWordWriters {
			if e := writer.File.Close(); e != nil && err == nil {
				err = errors.Wrapf(e, "Error while closing file %v", writer.File.Name())
			}
		}
	}()

	for {
		line, err := globalPageReader.ReadString('\n')

		if err != nil {
			break
		}
		if line == "}" {
			break
		}

		var page map[uint32]structures.TfidfAggregatedPage

		if line[:1] != "{" {
			line = "{" + line
		}

		line = line[:len(line)-2] + "}"
		err = json.Unmarshal([]byte(line), &page)
		if err != nil {
			return errors.Wrapf(err, "Error while unmarshalling json.")
		}
		for i := range page {
			for word := range *page[i].Words {
				err = writeWord(topicWordWriters, resultDir, page[i].TopicID, word)
				if err != nil {
					return errors.Wrap(err, "Error while writing json.")
				}
			}
		}
	}

	return nil
}

func mapWordsInFile(file string) (*map[string]uint32, error) {
	fileReader, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while trying to read file.")
	}

	wordMap := make(map[string]uint32)

	var tot uint32
	for _, word := range strings.Split(string(fileReader), "\n") {
		tot += wordMap[word]
		if _, ok := wordMap[word]; !ok {
			wordMap[word] = 1
		} else {
			wordMap[word]++
		}
	}

	wordMap["@TOT"] = tot
	return &wordMap, nil
}

func getJSONBytes(topicFile string, words *map[string]uint32) (*[]byte, error) {
	topicID := topicFile[len(topicFile)-10:]

	topicMap := make(map[string]*map[string]uint32)
	topicMap[topicID] = words

	wordsDict, err := json.Marshal(topicMap)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while unmarshalling "+topicFile)
	}

	return &wordsDict, nil
}

func topicWordsMapper(resultDir string) (err error) {
	topicFiles, err := utils.FilesInDir(resultDir, "T*")
	if err != nil {
		return err
	}

	outFile, err := os.Create(filepath.Join(resultDir, "GlobalTopicsWords.json"))
	if err != nil {
		return errors.Wrapf(err, "Error while creating file")
	}
	writer := bufio.NewWriter(outFile)
	defer func() {
		if e := outFile.Close(); e != nil && err == nil {
			err = errors.Wrapf(e, "Error while closing file %v", outFile.Name())
		}
	}()

	for i, topicFile := range topicFiles {
		topicWords, err := mapWordsInFile(topicFile)
		if err != nil {
			return err
		}

		jsonBytes, err := getJSONBytes(topicFile, topicWords)
		if err != nil {
			return err
		}
		jsonTopicWords := string(*jsonBytes)

		if i == 0 {
			jsonTopicWords = jsonTopicWords[:len(jsonTopicWords)-1] + ",\n"
			_, err = writer.Write([]byte(jsonTopicWords))

		} else if i > 0 {
			jsonTopicWords = jsonTopicWords[1:len(jsonTopicWords)-1] + ",\n"
			_, err = writer.Write([]byte(jsonTopicWords))
		}
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to write line in :"+resultDir+"GlobalTopicsWords.json")
		}
		err = writer.Flush()
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to flush:"+resultDir+"GlobalTopicsWords.json")
		}

		err = os.Remove(topicFile)
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to delete file:"+topicFile)
		}
	}
	_, err = writer.Write([]byte("}"))
	if err != nil {
		return errors.Wrapf(err, "Failed while trying to write line in :"+resultDir+"GlobalTopicsWords.json")
	}
	err = writer.Flush()
	if err != nil {
		return errors.Wrapf(err, "Failed while trying to flush:"+resultDir+"GlobalTopicsWords.json")
	}
	return nil
}

// TopicWords given the result dir process topics files containing the set of words and their absolute frequency
// which appear on pages of that topic
func TopicWords(resultDir string) error {
	err := topicWordsWriter(resultDir)
	if err != nil {
		return err
	}
	err = topicWordsMapper(resultDir)
	if err != nil {
		return err
	}
	return nil
}
