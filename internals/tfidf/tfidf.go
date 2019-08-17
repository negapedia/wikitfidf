package tfidf

import (
	"bufio"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"math"
	"os"

	"github.com/negapedia/wikiconflict/internals/structures"
)

// GetGlobalWords return full GlobalWord map and a error
func GetGlobalWord(resultDir string) (map[string]map[string]float64, error) {
	jsonFile, err := os.Open(resultDir + "GlobalWords.json")
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to open GlobalWords.json file:"+resultDir+"GlobalWords.json")
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to read GlobalWords.json file:"+resultDir+"GlobalWords.json")
	}

	err = jsonFile.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to close GlobalWords.json file:"+resultDir+"GlobalWords.json")
	}

	var globalWord map[string]map[string]float64

	err = json.Unmarshal(byteValue, &globalWord)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while unmarshalling json.")
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

	outFile, err := os.Create(resultDir + "GlobalPagesTFIDF.json")
	if err != nil {
		return errors.Wrapf(err, "Error happened while trying to create GlobalPagesTFIDF.json file:"+resultDir+"GlobalPagesTFIDF.json")
	}
	defer outFile.Close()
	encWriter := bufio.NewWriter(outFile)

	globalPage, err := os.Open(resultDir + "GlobalPages.json")
	defer globalPage.Close()
	if err != nil {
		return errors.Wrapf(err, "Error happened while trying to open GlobalPages.json file:"+resultDir+"GlobalPages.json")
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
			_, err = encWriter.Write([]byte(pageAsString))
		} else if i > 0 {
			marshalledPage, _ := json.Marshal(newPage)
			pageAsString := string(marshalledPage)
			pageAsString = pageAsString[1:len(pageAsString)-1] + ",\n"
			_, err = encWriter.Write([]byte(pageAsString))
		}
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to write line in :"+resultDir+"GlobalPagesTFIDF.json")
		}
		err = encWriter.Flush()
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to flush:"+resultDir+"GlobalPagesTFIDF.json")
		}
		i++
	}

	_, err = encWriter.Write([]byte("}"))
	if err != nil {
		return errors.Wrapf(err, "Failed while trying to write line in :"+resultDir+"GlobalPagesTFIDF.json")
	}
	err = encWriter.Flush()
	if err != nil {
		return errors.Wrapf(err, "Failed while trying to flush:"+resultDir+"GlobalPagesTFIDF.json")
	}

	err = os.Remove(resultDir + "GlobalPages.json")
	if err != nil {
		return errors.Wrapf(err, "Failed while trying to delete file:"+resultDir+"GlobalPages.json")
	}
	return nil
}
