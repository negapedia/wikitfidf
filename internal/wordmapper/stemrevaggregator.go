package wordmapper

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/negapedia/wikitfidf/internal/utils"
)

// StemRevAggregator given the result directory, will aggregate all Stem files into a single global file
func StemRevAggregator(resultDir string) error {
	fileList, err := utils.FilesInDir(filepath.Join(resultDir, "Stem/"), "StemRev_*")
	if err != nil {
		return err
	}
	//nFile := len(fileList)
	globalStemRev := make(map[string]string)

	for _, file := range fileList {
		//fmt.Printf("\rOn %d/%d", i+1, nFile)

		jsonFile, err := os.Open(file)
		// if we os.Open returns an error then handle it
		if err != nil {
			return errors.Wrapf(err, "Error while opening json %v", file)
		}

		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			return errors.Wrapf(err, "Error while reading json %v", file)
		}

		if err = jsonFile.Close(); err != nil {
			return errors.Wrapf(err, "Error while closing json %v", file)
		}

		err = os.Remove(file)
		if err != nil {
			return errors.Wrapf(err, "Error while removing json %v", file)
		}

		var StemDict map[string]string

		err = json.Unmarshal(byteValue, &StemDict)
		if err != nil {
			return errors.Wrapf(err, "Error while unmarshalling json")
		}

		for StemWord, RealWord := range StemDict {
			if _, ok := globalStemRev[StemWord]; ok { // already exists in globalStemRev
				if len(RealWord) < len(globalStemRev[StemWord]) { // if shorter, replace
					globalStemRev[StemWord] = RealWord
				}
			} else { // if not exists in globalStemRev
				globalStemRev[StemWord] = RealWord
			}
		}
	}

	return utils.Write2JSON(filepath.Join(resultDir, "GlobalStem.json"), globalStemRev)
}
