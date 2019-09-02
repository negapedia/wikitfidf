package utils

import (
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

// GetGlobalWordsTopN return GlobalWord_TopN map and a error
func GetGlobalWordsTopN(resultDir string, topN int) (map[string]uint32, error) {
	top := strconv.Itoa(topN)

	file, err := os.Open(resultDir + "GlobalWords_top" + top + ".json.gz")
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to opening GlobalWords_topN.json.gz file:"+resultDir+"GlobalWordstopN.json.gz")
	}
	defer file.Close()
	fileReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to create gzip reader")
	}
	defer fileReader.Close()

	byteValue, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to read GlobalWords.json file:"+resultDir+"GlobalWords.json")
	}

	var globalWord map[string]uint32

	err = json.Unmarshal(byteValue, &globalWord)
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to unmarshal GlobalWords.json file:"+resultDir+"GlobalWords.json")
	}

	return globalWord, nil
}
