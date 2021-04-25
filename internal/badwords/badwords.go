package badwords

import (
	"bufio"
	//"bytes"
	"compress/gzip"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	//"github.com/negapedia/wikitfidf/internal/assets"

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
		"simple": "english"}
	language, isIn := languages[lang]
	return language, isIn
}

func badWordsListGetter(lang string) (badwordsList map[string]bool, err error) {
	fpath := filepath.Join("/go/src/github.com/negapedia/wikitfidf/internal/data", lang)

	file, err := os.Open(fpath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	badwordsList = make(map[string]bool)
	scanner := bufio.NewScanner(file)
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
			return errors.Wrapf(err, "Failed while trying to create :"+resultDir+"BadWordsReport.json")
		}
		encWriter, err := gzip.NewWriterLevel(outFile, gzip.BestCompression)
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to create gzip writer for BadWordsReport.json")
		}
		defer func() {
			if e := encWriter.Close(); e != nil && err == nil {
				err = errors.Wrapf(err, "Error while closing gzip writer of %v", outFile.Name())
			}
		}()

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
					return errors.Wrapf(err, "Failed while trying to write line in :"+resultDir+"BadWordsReport.json")
				}
				err = encWriter.Flush()
				if err != nil {
					return errors.Wrapf(err, "Failed while trying to flush:"+resultDir+"BadWordsReport.json")
				}
			}
			i++

		}

		_, err = encWriter.Write([]byte("}"))
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to write line in :"+resultDir+"BadWordsReport.json")
		}
		err = encWriter.Flush()
		if err != nil {
			return errors.Wrapf(err, "Failed while trying to flush:"+resultDir+"BadWordsReport.json")
		}
	}
	return nil
}
