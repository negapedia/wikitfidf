package tfidf

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/negapedia/wikitfidf/internal/structures"
)

// GetGlobalWord return full GlobalWord map and a error
func GetGlobalWord(resultDir string) (map[string]map[string]float64, error) {
	jsonFile, err := os.Open(filepath.Join(resultDir, "GlobalWords.json"))
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
func ComputeTFIDF(resultDir string) (err error) {
	globalWord, err := GetGlobalWord(resultDir)
	if err != nil {
		return err
	}

	totalPage := globalWord["@Total Page"]["tot"]

	outFile, err := os.Create(filepath.Join(resultDir, "GlobalPagesTFIDF.json"))
	if err != nil {
		return errors.Wrapf(err, "Error happened while trying to create GlobalPagesTFIDF.json file:"+resultDir+"GlobalPagesTFIDF.json")
	}
	defer func() {
		if e := outFile.Close(); e != nil && err == nil {
			err = errors.Wrapf(e, "Error while closing file %v", outFile.Name())
		}
	}()
	encWriter := bufio.NewWriter(outFile)

	globalPage, err := os.Open(filepath.Join(resultDir, "GlobalPages.json"))
	defer func() {
		if e := globalPage.Close(); e != nil && err == nil {
			err = errors.Wrapf(e, "Error while closing file %v", globalPage.Name())
		}
	}()
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

		var page map[uint32]structures.AggregatedPage

		if line[:1] != "{" {
			line = "{" + line
		}

		line = line[:len(line)-2] + "}"
		err = json.Unmarshal([]byte(line), &page)
		if err != nil {
			return errors.Wrapf(err, "Error while unmarshalling json.")
		}

		newPageWords := make(map[string]map[string]float64)
		var newPage = make(map[uint32]structures.TfidfAggregatedPage)
		for id := range page {
			for word, wordFreq := range page[id].Words {
				tf := float64(wordFreq) / float64(page[id].Tot)
				appearIn := globalWord[word]["i"]
				idf := math.Log10(totalPage / appearIn)
				tfidf := math.Round((tf*idf)*10000) / 10000

				newPageWords[word] = make(map[string]float64)
				newPageWords[word]["abs"] = float64(wordFreq)
				newPageWords[word]["tfidf"] = tfidf
			}
			newPage[id] = structures.TfidfAggregatedPage{TopicID: page[id].TopicID, Tot: page[id].Tot, Words: &newPageWords}
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

	err = os.Remove(filepath.Join(resultDir, "GlobalPages.json"))
	if err != nil {
		return errors.Wrapf(err, "Failed while trying to delete file:"+resultDir+"GlobalPages.json")
	}
	return nil
}
