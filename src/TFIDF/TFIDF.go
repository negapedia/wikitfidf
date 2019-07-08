package TFIDF

import (
	"../DataStructure"
	"bufio"
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
)

func getGlobalWord(resultDir string) map[string]map[string]float64 {
	jsonFile, err := os.Open(resultDir+"GlobalWord.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		panic(err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	_ = jsonFile.Close()

	var globalWord map[string]map[string]float64

	err = json.Unmarshal(byteValue, &globalWord)
	if err != nil {
		panic(err)
	}

	return globalWord
}

func ComputeTFIDF(resultDir string) {
	globalWord := getGlobalWord(resultDir)

	totalPage := globalWord["@Total Page"]["tot"]

	outFile, _ := os.Create(resultDir + "GlobalPageTFIDF.json")
	encWriter := bufio.NewWriter(outFile)

	globalPage, err := os.Open(resultDir+"GlobalPage.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		panic(err)
	}
	globalPageReader := bufio.NewReader(globalPage)
	i := 0
	for {
		line, err := globalPageReader.ReadString('\n')

		if err != nil {
			break
		}
		if line == "}"{
			break
		}

		var page map[string]DataStructure.AggregatedPage

		if line[:1] != "{"{
			line = "{"+line
		}

		line = line[:len(line)-2]+"}"
		err = json.Unmarshal([]byte(line), &page)

		newPageWords := make(map[string]map[string]float64)
		var newPage = make(map[string]DataStructure.TfidfAggregatedPage)
		for i := range page {
			for word, wordFreq := range page[i].Words {
				tf := wordFreq / page[i].Tot
				appearIn := globalWord[word]["in"]
				idf := math.Log10(totalPage / appearIn)
				tfidf := tf * idf

				newPageWords[word] = make(map[string]float64)
				newPageWords[word]["abs"] = wordFreq
				newPageWords[word]["tfidf"] = tfidf
			}
			newPage[i] = DataStructure.TfidfAggregatedPage{Title:page[i].Title, Tot:page[i].Tot, Words: &newPageWords}
		}


		if i == 0 {
			marshalledPage, _ := json.Marshal(newPage)
			pageAsString := string(marshalledPage)
			pageAsString = pageAsString[:len(pageAsString)-1] + ",\n"
			_, _ = encWriter.Write([]byte(pageAsString))

		} else if  i > 0 {
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

	_ = os.Remove(resultDir + "GlobalPage.json")
	_ = globalPage.Close()
}
