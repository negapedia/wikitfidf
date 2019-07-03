package WordMapper


import (
	"../DataStructure"
	"../Utils"
	"encoding/json"
	"io/ioutil"
	"os"
)

func getMappedPage(page *DataStructure.StemmedPageJson) DataStructure.PageElement {
	var mappedText = make(map[string]int)

	for _, rev := range page.Revision {
		for _, word := range rev.Text {
			if _, ok := mappedText[word]; ok {
				mappedText[word] += 1
			} else {
				mappedText[word] = 1
			}
		}
	}
	return DataStructure.PageElement{PageId: page.PageID, Title: page.Title, Word: mappedText}
}


func WordMapperByPage(resultDir string) {
	fileList := Utils.FilesInDir(resultDir, ".json", "WS")

	for _, file := range fileList {
		println(file)

		jsonFile, err := os.Open(file)
		// if we os.Open returns an error then handle it
		if err != nil {
			panic(err)
		}
		// defer the closing of our jsonFile so that we can parse it later on

		// read our opened xmlFile as a byte array.
		byteValue, _ := ioutil.ReadAll(jsonFile)

		jsonFile.Close()

		// we initialize our Users array
		var page DataStructure.StemmedPageJson

		// we unmarshal our byteArray which contains our
		// jsonFile's content into 'users' which we defined above
		_ = json.Unmarshal(byteValue, &page)

		mappedPage := getMappedPage(&page)
		_ = os.Remove(file)
		if len(mappedPage.Word) > 0{
			Utils.WriteMappedPage(resultDir, &mappedPage)
		}
	}
}