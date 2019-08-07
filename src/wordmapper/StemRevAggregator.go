package wordmapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"../utils"
)

// StemRevAggregator given the result directory, will aggregate all Stem files into a single global file
func StemRevAggregator(resultDir string) {
	fileList := utils.FilesInDir(resultDir+"Stem/", "StemRev_*")
	nFile := len(fileList)
	globalStemRev := make(map[string]string)

	for i, file := range fileList {
		fmt.Printf("\rOn %d/%d\n", i+1, nFile)

		jsonFile, err := os.Open(file)
		// if we os.Open returns an error then handle it
		if err != nil {
			panic(err)
		}

		byteValue, err := ioutil.ReadAll(jsonFile)
		jsonFile.Close()
		if err != nil {
			panic(err)
		}

		err = os.Remove(file)
		if err != nil {
			panic(err)
		}

		var StemDict map[string]string

		err = json.Unmarshal(byteValue, &StemDict)
		if err != nil {
			panic(err)
		}

		for StemWord, RealWord := range StemDict {
			if _, ok := globalStemRev[StemWord]; ok {
				if len(RealWord) < len(globalStemRev[StemWord]) {
					globalStemRev[StemWord] = RealWord
				}
			} else {
				globalStemRev[StemWord] = RealWord
			}
		}
	}

	println(len(globalStemRev))
	utils.WriteGlobalStem(resultDir, &globalStemRev)
}
