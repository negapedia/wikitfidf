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
	"strings"

	"github.com/pkg/errors"

	"github.com/negapedia/wikiconflict/internals/structures"
	"github.com/negapedia/wikiconflict/internals/utils"
)

type topicWriter struct {
	Writer *bufio.Writer
	File   *os.File
}

func writeWord(topicWriters map[uint32]*topicWriter, resultDir string, topicID uint32, word string) {
	if _, ok := topicWriters[topicID]; !ok {
		outFile, _ := os.Create(resultDir + "T" + fmt.Sprint(topicID))
		topicWriters[topicID] = &topicWriter{Writer: bufio.NewWriter(outFile), File: outFile}
	}

	_, _ = topicWriters[topicID].Writer.Write([]byte(word + "\n"))
	_ = topicWriters[topicID].Writer.Flush()
}

func closeAll(topicWriters map[uint32]*topicWriter) {
	for _, writer := range topicWriters {
		writer.File.Close()
	}
}

func topicWordsWriter(resultDir string) error {
	globalPageTFIDF, err := os.Open(resultDir + "GlobalPagesTFIDF.json")
	defer globalPageTFIDF.Close()
	if err != nil {
		return errors.Wrapf(err, "Error happened while trying to open GlobalPagesTFIDF.json file:"+resultDir+"GlobalPagesTFIDF.json")
	}
	globalPageReader := bufio.NewReader(globalPageTFIDF)

	topicWordWriters := make(map[uint32]*topicWriter)

	for {
		line, err := globalPageReader.ReadString('\n')

		if err != nil {
			break
		}
		if line == "}" {
			break
		}

		var page map[string]structures.TfidfAggregatedPage

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
				writeWord(topicWordWriters, resultDir, page[i].TopicID, word)
			}
		}
	}

	closeAll(topicWordWriters)
	return nil
}

func mapWordsInFile(file string) (*map[string]uint32, error) {
	fileReader, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while trying to read file.")
	}

	wordMap := make(map[string]uint32)

	for _, word := range strings.Split(string(fileReader), "\n") {
		if _, ok := wordMap[word]; !ok {
			wordMap[word] = 1
		} else {
			wordMap[word]++
		}
	}

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

func topicWordsMapper(resultDir string) error {
	topicFiles, err := utils.FilesInDir(resultDir, "T*")
	if err != nil {
		return err
	}

	outFile, err := os.Create(resultDir + "GlobalTopicsWords.json")
	if err != nil {
		return errors.Wrapf(err, "Error while creating file")
	}
	writer := bufio.NewWriter(outFile)
	defer outFile.Close()

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
