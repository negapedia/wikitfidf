package TextNormalizer

import "os/exec"

func CallStopwStemm(resultDir string, lang string) {
	stopwordsCleanerStemming := exec.Command("python3", "-m", "cProfile", "-o", "./StopWStemProf.txt", "-s", "cumulative", "./TextNormalizer/StopwordsCleaner_Stemming.py", "../"+resultDir, lang)
	err := stopwordsCleanerStemming.Run()
	if err != nil {
		panic(err)
	}
}
