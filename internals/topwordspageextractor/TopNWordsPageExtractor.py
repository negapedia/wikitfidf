#!/usr/bin/env python
# -*- coding: utf-8 -*-

import json
import sys
from collections import Counter


def _top_n_getter(words_dict: dict, n: int):
    top_n = Counter(words_dict).most_common(n)
    words_dict = {}
    for key, value in top_n:
        words_dict[key] = value
    return words_dict


def _get_top_n_words_dict(pageDict: dict, n: int):
    words_dict = {}
    for page in pageDict:
        for word in pageDict[page]["Words"]:
            words_dict[word] = pageDict[page]["Words"][word]["tfidf"]

        top_n_page = {}
        if len(words_dict) > n:
            top_n_page[page] = {"TopicID": pageDict[page]["TopicID"], "Tot": pageDict[page]["Tot"],
                                "Words": _top_n_getter(words_dict, n)}
        else:
            top_n_page[page] = {"TopicID": pageDict[page]["TopicID"], "Tot": pageDict[page]["Tot"],
                                "Words": words_dict}

        return top_n_page


def top_n_Words_Page_Extractor(result_dir: str, n):
    """
    top_N_Words_Page_Extractor given the result dir compute the n most important words for each page in GlobalPageTFIDF
    :param result_dir: result dir path
    :param n: amount of most important words to calculate
    """
    globalTopNTFIDF = open(result_dir + "GlobalPagesTFIDF_top" + n + ".json", "w")

    gloabalTFIDF = open(result_dir + "GlobalPagesTFIDF.json", "r");
    globalTFIDF_it = iter(gloabalTFIDF.readline, "")

    n = int(n)

    counter = 0
    for line in globalTFIDF_it:
        if line == "}":
            break

        if line[:1] != "{":
            line = "{" + line

        line = line[:len(line) - 2] + "}"

        pageDict = json.loads(line)
        pageDict = _get_top_n_words_dict(pageDict, n)

        if counter == 0:
            page_json = json.dumps(pageDict)
            page_json = page_json[:len(page_json) - 1] + ",\n"
            globalTopNTFIDF.write(page_json)
        elif counter >= 0:
            page_json = json.dumps(pageDict)
            page_json = page_json[1:len(page_json) - 1] + ",\n"
            globalTopNTFIDF.write(page_json)
        globalTopNTFIDF.flush()
        counter += 1

    globalTopNTFIDF.write("}")
    globalTopNTFIDF.flush()
    globalTopNTFIDF.close()
    gloabalTFIDF.close()


if __name__ == "__main__":
    top_n_Words_Page_Extractor(sys.argv[1], sys.argv[2])
