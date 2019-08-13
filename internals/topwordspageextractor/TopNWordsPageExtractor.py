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


def _get_top_n_words_pages_dict(pageDict: dict, n: int):
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

def _get_global_words(globalDict: dict):
    newGlobalDict = {}
    for word in globalDict:
        if word == "@Total Word" or word == "@Total Page":
            continue
        newGlobalDict[word] = globalDict[word]["a"]

    return newGlobalDict



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
        pageDict = _get_top_n_words_pages_dict(pageDict, n)

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


def top_n_Global_Words_Extractor(result_dir: str, n):
    """
    top_n_Global_Words_Extractor given the result dir compute the n most frequent word in GlobalWord
    :param result_dir: result dir path
    :param n: amount of most important words to calculate
    """
    globalWordTopN = open(result_dir + "GlobalWords_top" + n + ".json", "w")

    with open(result_dir + "GlobalWords.json", "r") as gloabalWords:
        globalWordsDict = json.load(gloabalWords)

    globalWordsDict = _get_global_words(globalWordsDict)
    json.dump(_top_n_getter(globalWordsDict, int(n)), globalWordTopN)
    globalWordTopN.flush()
    globalWordTopN.close()

def top_n_Topic_Words_Extractor(result_dir: str, n):
    """
    top_n_Global_Words_Extractor given the result dir compute the n most frequent word in GlobalWord
    :param result_dir: result dir path
    :param n: amount of most important words to calculate
    """
    globalWordTopN = open(result_dir + "GlobalTopicsWords_top" + n + ".json", "w")

    globalTopic = open(result_dir + "GlobalTopicsWords.json", "r");
    globalTopic_iter = iter(globalTopic.readline, "")

    n = int(n)

    counter = 0
    for line in globalTopic_iter:
        if line == "}":
            break

        if line[:1] != "{":
            line = "{" + line

        line = line[:len(line) - 2] + "}"

        topicDict = json.loads(line)

        for topic in topicDict:
            topWords = {topic: _top_n_getter(topicDict[topic], n)}

        if counter == 0:
            page_json = json.dumps(topWords)
            page_json = page_json[:len(page_json) - 1] + ",\n"
            globalWordTopN.write(page_json)
        elif counter >= 0:
            page_json = json.dumps(topWords)
            page_json = page_json[1:len(page_json) - 1] + ",\n"
            globalWordTopN.write(page_json)
        globalWordTopN.flush()
        counter += 1

    globalWordTopN.write("}")
    globalWordTopN.flush()
    globalWordTopN.close()
    globalTopic.close()

if __name__ == "__main__":
    #top_n_Words_Page_Extractor(sys.argv[1], sys.argv[2])
    #top_n_Global_Words_Extractor(sys.argv[1], sys.argv[2])
    top_n_Topic_Words_Extractor(sys.argv[1], sys.argv[2])
