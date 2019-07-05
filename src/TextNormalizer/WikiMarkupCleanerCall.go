package TextNormalizer

import "os/exec"

func CallCleanMarkup(resultDir string) {
	wikiMarkupRemoval := exec.Command("python3", "-m", "cProfile", "-o", "./WikiMarkProf.txt", "-s", "cumulative", "./TextNormalizer/WikiMarkupCleaner.py", "../"+resultDir, ">", "WikiMark_Profile.txt")
	err := wikiMarkupRemoval.Run()
	if err != nil {
		panic(err)
	}
}
