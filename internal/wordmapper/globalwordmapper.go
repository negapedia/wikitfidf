package wordmapper

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/negapedia/wikitfidf/internal/structures"
	"github.com/negapedia/wikitfidf/internal/utils"
)

// GlobalWordMapper given the result dir, generate the file containing the global report about word frequency
func GlobalWordMapper(resultDir string) (err error) {
	fileList, err := utils.FilesInDir(resultDir, "M[0-9]*")
	if err != nil {
		return err
	}

	globalWord := make(map[string]map[string]uint32)
	var totalWord uint32
	var totalPage uint32

	for _, file := range fileList {

		jsonFile, err := os.Open(file)
		if err != nil {
			return errors.Wrapf(err, "Error happened while trying to open json %v", file)
		}

		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			return errors.Wrapf(err, "Error while reading json %v", file)
		}

		if err = jsonFile.Close(); err != nil {
			return errors.Wrapf(err, "Error while closing json %v", file)
		}

		var page structures.PageElement
		if err = json.Unmarshal(byteValue, &page); err != nil {
			return errors.Wrapf(err, "Error while unmarshalling json %v", file)
		}

		totalPage++

		for word, freq := range page.Word {
			if _, ok := globalWord[word]; ok {
				globalWord[word]["a"] += uint32(freq) // a --> abs, i --> in
				globalWord[word]["i"]++
			} else {
				globalWord[word] = make(map[string]uint32)
				globalWord[word]["a"] = uint32(freq)
				globalWord[word]["i"] = 1
			}
			totalWord += uint32(freq)
		}
	}

	globalWord["@Total Word"] = make(map[string]uint32)
	globalWord["@Total Word"]["tot"] = totalWord
	globalWord["@Total Page"] = make(map[string]uint32)
	globalWord["@Total Page"]["tot"] = totalPage

	return utils.Write2JSON(filepath.Join(resultDir, "GlobalWords.json"), globalWord)
}
