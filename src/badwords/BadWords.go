package badwords

import (
	"bufio"
	"encoding/json"
	"os"

	"../structures"
)

func availableLanguage(lang string) (string, bool) {
	languages := map[string]string{
		"en": "english",
		"ar": "arabic",
		"da": "danish",
		"nl": "dutch",
		"fi": "finnish",
		"fr": "french",
		"de": "german",
		"hu": "hungarian",
		"it": "italian",
		"no": "norwegian",
		"pt": "portuguese",
		"es": "spanish",
		"sv": "swedish",
		"zh": "chinese",
		"cs": "czech",
		"hi": "hindi",
		"ja": "japanese",
		"ko": "korean",
		"fa": "persian",
		"pl": "polish",
		"th": "thai",
	}
	language, isIn := languages[lang]
	return language, isIn
}

func badWordsListGetter(lang, path string) map[string]bool {
	file, err := os.Open(path + lang)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	badwordsList := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		badwordsList[scanner.Text()] = true
	}

	return badwordsList
}

// BadWords create the badwords report for the given language, if available, and the given result dir
func BadWords(lang, resultDir string) {
	if language, isAvailable := availableLanguage(lang); isAvailable {
		badWordsMap := badWordsListGetter(language, "/root/badwords_data/") // TODO path to /root/badwords_data/

		outFile, _ := os.Create(resultDir + "BadWordsReport.json")
		encWriter := bufio.NewWriter(outFile)
		defer outFile.Close()

		globalPage, err := os.Open(resultDir + "GlobalPageTFIDF.json")
		defer globalPage.Close()
		// if we os.Open returns an error then handle it
		if err != nil {
			panic(err)
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

			var page map[string]structures.TfidfAggregatedPage

			if line[:1] != "{" {
				line = "{" + line
			}

			line = line[:len(line)-2] + "}"
			err = json.Unmarshal([]byte(line), &page)
			if err != nil {
				panic(err)
			}

			toIgnore := false
			newPage := make(map[string]structures.BadWordsReport)
			for p := range page {
				badwordInPage := make(map[string]int)
				var totalBadW uint32
				for word := range *page[p].Words {
					if _, isBadword := badWordsMap[word]; isBadword {
						totalBadW++
						if _, ok := badwordInPage[word]; ok {
							badwordInPage[word] += 1
						} else {
							badwordInPage[word] = 1
						}
					}
				}

				if len(badwordInPage) > 0 {
					newPage[p] = structures.BadWordsReport{Abs: totalBadW, Rel: float64(totalBadW) / page[p].Tot, BadW: badwordInPage}
				} else {
					toIgnore = true
				}

			}

			if !toIgnore {
				if i == 0 {
					marshalledPage, _ := json.Marshal(newPage)
					pageAsString := string(marshalledPage)
					pageAsString = pageAsString[:len(pageAsString)-1] + ",\n"
					_, _ = encWriter.Write([]byte(pageAsString))

				} else if i > 0 {
					marshalledPage, _ := json.Marshal(newPage)
					pageAsString := string(marshalledPage)
					pageAsString = pageAsString[1:len(pageAsString)-1] + ",\n"
					_, _ = encWriter.Write([]byte(pageAsString))

				}
				_ = encWriter.Flush()
			}
			i++

		}

		_, _ = encWriter.Write([]byte("}"))
		_ = encWriter.Flush()
	}
}
