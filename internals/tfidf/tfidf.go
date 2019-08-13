package tfidf

import (
	"bufio"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"math"
	"os"

	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/structures"
)

func GetGlobalWord(resultDir string) (map[string]map[string]float64, error) {
	jsonFile, err := os.Open(resultDir + "GlobalWords.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to open GlobalWords.json file:"+resultDir + "GlobalWords.json")
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	_ = jsonFile.Close()

	var globalWord map[string]map[string]float64

	err = json.Unmarshal(byteValue, &globalWord)
	if err != nil {
		log.Fatal("Error while unmarshalling json.",err)
	}

	return globalWord, nil
}

// ComputeTFIDF given the result dir, compute the TFIDF for all available pages
func ComputeTFIDF(resultDir string) error {
	globalWord, err := GetGlobalWord(resultDir)
	if err != nil {
		return err
	}

	totalPage := globalWord["@Total Page"]["tot"]

	outFile, _ := os.Create(resultDir + "GlobalPagesTFIDF.json")
	defer outFile.Close()
	encWriter := bufio.NewWriter(outFile)

	globalPage, err := os.Open(resultDir + "GlobalPages.json")
	defer globalPage.Close()
	if err != nil {
		return errors.Wrapf(err, "Error happened while trying to open GlobalPages.json file:"+ resultDir + "GlobalPages.json")
	}
	globalPageReader := bufio.NewReader(globalPage)
	i := 0
	for {
		line, err := globalPageReader.ReadString('\n')

		if err != nil {
			break
		}
		if line == "}" {
			break
		}

		var page map[string]structures.AggregatedPage

		if line[:1] != "{" {
			line = "{" + line
		}

		line = line[:len(line)-2] + "}"
		err = json.Unmarshal([]byte(line), &page)
		if err != nil {
			return errors.Wrapf(err, "Error while unmarshalling json.")
		}

		newPageWords := make(map[string]map[string]float64)
		var newPage = make(map[string]structures.TfidfAggregatedPage)
		for i := range page {
			for word, wordFreq := range page[i].Words {
				tf := float64(wordFreq) / float64(page[i].Tot)
				appearIn := globalWord[word]["i"]
				idf := math.Log10(totalPage / appearIn)
				tfidf := math.Round((tf*idf)*10000) / 10000

				newPageWords[word] = make(map[string]float64)
				newPageWords[word]["abs"] = float64(wordFreq)
				newPageWords[word]["tfidf"] = tfidf
			}
			newPage[i] = structures.TfidfAggregatedPage{TopicID: page[i].TopicID, Tot: page[i].Tot, Words: &newPageWords}
		}

		if i == 0 {
			marshalledPage, _ := json.Marshal(newPage)
			pageAsString := string(marshalledPage)
			pageAsString = pageAsString[:len(pageAsString)-1] + ",\n"
			_, _ = encWriter.Write([]byte(pageAsString))

		} else if i > 0 {
			marshalledPage, _ := json.Marshal(newPage)
			pageAsString := string(marshalledPage)
			pageAsString = pageAsString[1:len(pageAsString)-1] + ",\n"
			_, _ = encWriter.Write([]byte(pageAsString))

		}
		_ = encWriter.Flush()
		i++
	}

	_, _ = encWriter.Write([]byte("}"))
	_ = encWriter.Flush()

	_ = os.Remove(resultDir + "GlobalPages.json")
	return nil
}
