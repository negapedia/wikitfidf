package utils

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func GetGlobalWordsTopN(resultDir string, topN int) (map[string]uint32, error) {
	top := strconv.Itoa(topN)

	jsonFile, err := os.Open(resultDir + "GlobalWords_top"+top+".json")
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to open GlobalWords.json file:"+resultDir + "GlobalWords.json")
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	_ = jsonFile.Close()

	var globalWord map[string]uint32

	err = json.Unmarshal(byteValue, &globalWord)
	if err != nil {
		log.Fatal("Error while unmarshalling json.",err)
	}

	return globalWord, nil
}
