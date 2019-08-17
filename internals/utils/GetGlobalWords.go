package utils

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"strconv"
)

// GetGlobalWordsTopN return GlobalWord_TopN map and a error
func GetGlobalWordsTopN(resultDir string, topN int) (map[string]uint32, error) {
	top := strconv.Itoa(topN)

	byteValue, err := ioutil.ReadFile(resultDir + "GlobalWords_top"+top+".json")
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to open GlobalWords.json file:"+resultDir + "GlobalWords.json")
	}

	var globalWord map[string]uint32

	err = json.Unmarshal(byteValue, &globalWord)
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to unmarshal GlobalWords.json file:"+resultDir + "GlobalWords.json")
	}

	return globalWord, nil
}
