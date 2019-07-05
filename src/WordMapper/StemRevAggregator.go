package WordMapper

import (
	"../Utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func StemRevAggregator(resultDir string) {
	fileList := Utils.FilesInDir(resultDir, ".json", "StemRev_")
	nFile := len(fileList)

	globalStemRev := make(map[string]string)

	for i, file := range fileList {
		fmt.Printf("\rOn %d/%d", i+1, nFile)

		jsonFile, err := os.Open(file)
		// if we os.Open returns an error then handle it
		if err != nil {
			panic(err)
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)
		_ = jsonFile.Close()
		_ = os.Remove(file)

		var StemDict map[string]string

		_ = json.Unmarshal(byteValue, &StemDict)

		for StemWord, RealWord := range StemDict {
			if _, ok := globalStemRev[StemWord]; ok {
				if len(RealWord) <  len(globalStemRev[StemWord]) {
					globalStemRev[StemWord] = RealWord
				}
			} else {
				globalStemRev[StemWord] = RealWord
			}
		}
	}

	Utils.WriteGlobalStem(resultDir, &globalStemRev)
}
