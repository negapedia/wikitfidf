package wordmapper

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/negapedia/wikiconflict/internals/utils"
)

// StemRevAggregator given the result directory, will aggregate all Stem files into a single global file
func StemRevAggregator(resultDir string) error{
	fileList, err := utils.FilesInDir(resultDir+"Stem/", "StemRev_*")
	if err != nil {
		return err
	}
	nFile := len(fileList)
	globalStemRev := make(map[string]string)

	for i, file := range fileList {
		fmt.Printf("\rOn %d/%d", i+1, nFile)

		jsonFile, err := os.Open(file)
		// if we os.Open returns an error then handle it
		if err != nil {
			return errors.Wrapf(err, "error while opening file")
		}

		byteValue, err := ioutil.ReadAll(jsonFile)
		jsonFile.Close()
		if err != nil {
			return errors.Wrapf(err, "error while reading file")
		}

		err = os.Remove(file)
		if err != nil {
			log.Fatal(err)
		}

		var StemDict map[string]string

		err = json.Unmarshal(byteValue, &StemDict)
		if err != nil {
			return errors.Wrapf(err, "error while unmarshalling json")
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
	fmt.Println()

	println(len(globalStemRev))
	utils.WriteGlobalStem(resultDir, &globalStemRev)
	return nil
}
