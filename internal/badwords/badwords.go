package badwords

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/negapedia/wikitfidf/internal/assets"

	"github.com/pkg/errors"

	"github.com/negapedia/wikitfidf/internal/structures"
)

//AvailableLanguage checks if bad words are avaible for a language
func AvailableLanguage(lang string) (string, bool) {
	languages := map[string]string{
		"en":     "english",
		"ar":     "arabic",
		"da":     "danish",
		"nl":     "dutch",
		"fi":     "finnish",
		"fr":     "french",
		"de":     "german",
		"hu":     "hungarian",
		"it":     "italian",
		"no":     "norwegian",
		"pt":     "portuguese",
		"es":     "spanish",
		"sv":     "swedish",
		"zh":     "chinese",
		"cs":     "czech",
		"hi":     "hindi",
		"ja":     "japanese",
		"ko":     "korean",
		"fa":     "persian",
		"pl":     "polish",
		"th":     "thai",
		"simple": "english",
		"vec":    "italian", // only as test
	}
	language, isIn := languages[lang]
	return language, isIn
}

func badWordsListGetter(lang string) (badwordsList map[string]bool, err error) {
	fpath := filepath.Join("badwords", "data", lang)
	assetData, err := assets.Asset(fpath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error happened while trying to open badwords asset file:%v", fpath)
	}

	badwordsList = make(map[string]bool)
	scanner := bufio.NewScanner(bytes.NewBuffer(assetData))
	for scanner.Scan() {
		badwordsList[scanner.Text()] = true
	}

	return badwordsList, nil
}

// BadWords create the badwords report for the given language, if available, and the given result dir
func BadWords(lang, resultDir string) (err error) {
	if language, isAvailable := AvailableLanguage(lang); isAvailable {
		badWordsMap, err := badWordsListGetter(language)
		if err != nil {
			return err
		}

		outFile, err := os.Create(filepath.Join(resultDir, "BadWordsReport.json.gz"))
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to create :"+resultDir+"BadWordsReport.json.gz")
		}
		encWriter, err := gzip.NewWriterLevel(outFile, gzip.BestCompression)
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to create gzip writer for BadWordsReport.json.gz")
		}
		defer func() {
			if e := outFile.Close(); e != nil && err == nil {
				err = errors.Wrapf(err, "Error while closing file %v", outFile.Name())
			}
		}()

		globalPage, err := os.Open(filepath.Join(resultDir, "GlobalPagesTFIDF.json"))
		defer func() {
			if e := globalPage.Close(); e != nil && err == nil {
				err = errors.Wrapf(err, "Error while closing file %v", globalPage.Name())
			}
		}()

		if err != nil {
			return errors.Wrapf(err, "Error happened while trying to open GlobalPagesTFIDF.json file:"+resultDir+"GlobalPagesTFIDF.json")
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

			var page map[uint32]structures.TfidfAggregatedPage

			if line[:1] != "{" {
				line = "{" + line
			}

			line = line[:len(line)-2] + "}"
			err = json.Unmarshal([]byte(line), &page)
			if err != nil {
				return errors.Wrapf(err, "error while unmarshalling json")
			}

			toIgnore := false
			newPage := make(map[uint32]structures.BadWordsReport)
			for p := range page {
				badwordInPage := make(map[string]uint32)
				var totalBadW uint32
				for word := range *page[p].Words {
					if _, isBadword := badWordsMap[word]; isBadword {
						totalBadW++
						if _, ok := badwordInPage[word]; ok {
							badwordInPage[word]++
						} else {
							badwordInPage[word] = 1
						}
					}
				}

				if len(badwordInPage) > 0 {
					newPage[p] = structures.BadWordsReport{TopicID: page[p].TopicID, Abs: totalBadW, Rel: float64(totalBadW) / float64(page[p].Tot), BadW: badwordInPage}
				} else {
					toIgnore = true // no badwords in this page
				}

			}

			if !toIgnore {
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
					return errors.Wrapf(err, "Failed while trying to write line in :"+resultDir+"BadWordsReport.json.gz")
				}
				err = encWriter.Flush()
				if err != nil {
					return errors.Wrapf(err, "Failed while trying to flush:"+resultDir+"BadWordsReport.json.gz")
				}
			}
			i++

		}

		_, err = encWriter.Write([]byte("}"))
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to write line in :"+resultDir+"BadWordsReport.json.gz")
		}
		err = encWriter.Flush()
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to flush:"+resultDir+"BadWordsReport.json.gz")
		}
		err = encWriter.Close()
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to close writer of:"+resultDir+"BadWordsReport.json.gz")
		}
	}
	return nil
}

// TopicBadWords built a file with the list of badwords in each topic with total number of badwords
// per topic and number of occurence for each badword
func TopicBadWords(lang, resultDir string) (err error) {
	if language, isAvailable := AvailableLanguage(lang); isAvailable {
		badWordsMap, err := badWordsListGetter(language)
		if err != nil {
			return err
		}

		outFile, err := os.Create(filepath.Join(resultDir, "TopicBadWords.json.gz"))
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to create :"+resultDir+"TopicBadWords.json.gz")
		}
		encWriter, err := gzip.NewWriterLevel(outFile, gzip.BestCompression)
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to create gzip writer for TopicBadWords.json.gz")
		}
		defer func() {
			if e := outFile.Close(); e != nil && err == nil {
				err = errors.Wrapf(err, "Error while closing file %v", outFile.Name())
			}
		}()

		globalTopic, err := os.Open(filepath.Join(resultDir, "GlobalTopicsWords.json"))
		defer func() {
			if e := globalTopic.Close(); e != nil && err == nil {
				err = errors.Wrapf(err, "Error while closing file %v", globalTopic.Name())
			}
		}()

		if err != nil {
			return errors.Wrapf(err, "Error happened while trying to open GlobalTopicsWords.json file:"+resultDir+"GlobalTopicsWords.json")
		}
		topicReader := bufio.NewReader(globalTopic)

		i := 0

		for {
			line, err := topicReader.ReadString('\n')

			if err != nil {
				break
			}
			if line == "}" {
				break
			}

			if line[:1] != "{" {
				line = "{" + line
			}

			line = line[:len(line)-2] + "}"

			var topic map[uint32]map[string]uint32

			println(line[:15])
			err = json.Unmarshal([]byte(line), &topic)
			if err != nil {
				return errors.Wrapf(err, "error while unmarshalling json")
			}

			toIgnore := false
			newTopic := make(map[uint32]structures.TopicBadWords)
			for p := range topic {
				badwordInTopic := make(map[string]uint32)
				var totalBadW uint32
				for word := range topic[p] {
					if _, isBadword := badWordsMap[word]; isBadword {
						totalBadW++
						if _, ok := badwordInTopic[word]; ok {
							badwordInTopic[word]++
						} else {
							badwordInTopic[word] = 1
						}
					}
				}

				if len(badwordInTopic) > 0 {
					newTopic[p] = structures.TopicBadWords{TotBadw: totalBadW, BadW: badwordInTopic}
				} else {
					toIgnore = true // no badwords in this page
				}

			}

			if !toIgnore {
				if i == 0 {
					marshalledPage, _ := json.Marshal(newTopic)
					topicAsString := string(marshalledPage)
					topicAsString = topicAsString[:len(topicAsString)-1] + ",\n"
					_, err = encWriter.Write([]byte(topicAsString))
				} else if i > 0 {
					marshalledPage, _ := json.Marshal(newTopic)
					topicAsString := string(marshalledPage)
					topicAsString = topicAsString[1:len(topicAsString)-1] + ",\n"
					_, err = encWriter.Write([]byte(topicAsString))
				}
				if err != nil {
					return errors.Wrapf(err, "Failed while trying to write line in :"+resultDir+"TopicBadWords.json.gz")
				}
				err = encWriter.Flush()
				if err != nil {
					return errors.Wrapf(err, "Failed while trying to flush:"+resultDir+"TopicBadWords.json.gz")
				}
			}
			i++

		}

		_, err = encWriter.Write([]byte("}"))
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to write line in :"+resultDir+"TopicBadWords.json.gz")
		}
		err = encWriter.Flush()
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to flush:"+resultDir+"TopicBadWords.json.gz")
		}
		err = encWriter.Close()
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to close writer of:"+resultDir+"TopicBadWords.json.gz")
		}
	}
	return nil
}
