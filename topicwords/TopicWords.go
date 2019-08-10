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
	"log"
	"os"
	"strings"

	"github.com/MarcoChilese/Wikipedia-Conflict-Analyzer/structures"
	"github.com/MarcoChilese/Wikipedia-Conflict-Analyzer/utils"
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

func topicWordsWriter(resultDir string) {
	globalPageTFIDF, err := os.Open(resultDir + "GlobalPagesTFIDF.json")
	defer globalPageTFIDF.Close()
	// if we os.Open returns an error then handle it
	if err != nil {
		log.Fatal(err)
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
			panic(err)
		}
		for i := range page {
			for word := range *page[i].Words {
				writeWord(topicWordWriters, resultDir, page[i].TopicID, word)
			}
		}
	}

	closeAll(topicWordWriters)
}

func mapWordsInFile(file string) *map[string]uint32 {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fileReader, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	wordMap := make(map[string]uint32)

	for _, word := range strings.Split(string(fileReader), "\n") {
		if _, ok := wordMap[word]; !ok {
			wordMap[word] = 1
		} else {
			wordMap[word] += 1
		}
	}

	return &wordMap
}

func getJSONBytes(topicFile string, words *map[string]uint32) *[]byte {
	topicID := topicFile[len(topicFile)-10:]

	topicMap := make(map[string]*map[string]uint32)
	topicMap[topicID] = words

	wordsDict, err := json.Marshal(topicMap)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	return &wordsDict
}

func topicWordsMapper(resultDir string) {
	topicFiles := utils.FilesInDir(resultDir, "T*")

	outFile, err := os.Create(resultDir + "GlobalTopicsWords.json")
	if err != nil {
		log.Fatalf("%+v", err)
	}
	writer := bufio.NewWriter(outFile)
	defer outFile.Close()

	for i, topicFile := range topicFiles {
		topicWords := mapWordsInFile(topicFile)
		jsonTopicWords := string(*getJSONBytes(topicFile, topicWords))

		if i == 0 {
			jsonTopicWords = jsonTopicWords[:len(jsonTopicWords)-1] + ",\n"
			_, _ = writer.Write([]byte(jsonTopicWords))

		} else if i > 0 {
			jsonTopicWords = jsonTopicWords[1:len(jsonTopicWords)-1] + ",\n"
			_, _ = writer.Write([]byte(jsonTopicWords))

		}
		_ = writer.Flush()

		_ = os.Remove(topicFile)
	}
	writer.Write([]byte("}"))
	writer.Flush()
}

// TopicWords given the result dir process topics files containing the set of words and their absolute frequency
// which appear on pages of that topic
func TopicWords(resultDir string) {
	topicWordsWriter(resultDir)
	topicWordsMapper(resultDir)
}
