#!/usr/bin/env python
#cython: language_level=3
# -*- coding: utf-8 -*-

#IF YOU MODIFY THIS FILE, YOU NEED TO RUN "go generate" IN "assets" FOR CHANGES TO TAKE EFFECT.

import gzip
import json
import os
import sys
from collections import Counter
from os.path import join

MIN_WORD_LENGTH = 1  # Min lenght for words (might be changed in-program wrt language)
MAX_WORD_LENGTH = 33  # Max lenght for words


def _top_n_getter(words_dict: dict, n: int):
    top_n = Counter(words_dict).most_common(n)
    words_dict = {}
    for key, value in top_n:
        if MAX_WORD_LENGTH >= len(key) >= MIN_WORD_LENGTH:
            words_dict[key] = value
    return words_dict


def _get_top_n_words_pages_dict(page_dict: dict, n: int):
    words_dict = {}
    for page in page_dict:
        for word in page_dict[page]["Words"]:
            words_dict[word] = page_dict[page]["Words"][word]["tfidf"]

        top_n_page = {}
        if len(words_dict) > n:
            top_n_page[page] = {"TopicID": page_dict[page]["TopicID"], "Tot": page_dict[page]["Tot"],
                                "Words": _top_n_getter(words_dict, n)}
        else:
            top_n_page[page] = {"TopicID": page_dict[page]["TopicID"], "Tot": page_dict[page]["Tot"],
                                "Words": words_dict}

        return top_n_page


def _get_global_words(global_dict: dict):
    new_global_dict = {}
    for word in global_dict:
        if word == "@Total Word" or word == "@Total Page" or not (MAX_WORD_LENGTH >= len(word) >= MIN_WORD_LENGTH) :
            continue
        new_global_dict[word] = global_dict[word]["a"]

    return new_global_dict


def top_n_words_page_extractor(result_dir: str, n, delete: bool):
    """
    top_N_Words_Page_Extractor given the result dir compute the n most important words for each page in GlobalPageTFIDF.
    After processing, original file is deleted if delete is true
    :param result_dir: result dir path
    :param n: amount of most important words to calculate
    :param delete: if true after processing original file is deleted
    """
    global_top_ntfidf = gzip.GzipFile(filename=join(result_dir, "GlobalPagesTFIDF_topN.json.gz"), mode="w",
                                      compresslevel=9)

    gloabal_tfidf = open(join(result_dir, "GlobalPagesTFIDF.json"), "r", encoding='utf-8')
    global_tfidf_it = iter(gloabal_tfidf.readline, "")

    n = int(n)

    counter = 0
    for line in global_tfidf_it:
        if line == "}":
            break

        if line[:1] != "{":
            line = "{" + line

        line = line[:len(line) - 2] + "}"

        page_dict = json.loads(line)
        page_dict = _get_top_n_words_pages_dict(page_dict, n)

        if counter == 0:
            page_json = json.dumps(page_dict)
            page_json = page_json[:len(page_json) - 1] + ",\n"
            global_top_ntfidf.write(page_json.encode('utf-8'))
        elif counter >= 0:
            page_json = json.dumps(page_dict)
            page_json = page_json[1:len(page_json) - 1] + ",\n"
            global_top_ntfidf.write(page_json.encode('utf-8'))
        global_top_ntfidf.flush()
        counter += 1

    global_top_ntfidf.write("}".encode('utf-8'))
    global_top_ntfidf.close()
    gloabal_tfidf.close()

    if delete:
        os.remove(join(result_dir, "GlobalPagesTFIDF.json"))


def top_n_global_words_extractor(result_dir: str, n, delete: bool):
    """
    top_n_Global_Words_Extractor given the result dir compute the n most frequent word in GlobalWord.
    After processing, original file is deleted if delete is true
    :param result_dir: result dir path
    :param n: amount of most important words to calculate
    :param delete: if true after processing original file is deleted
    """

    global_word_top_n = gzip.GzipFile(filename=join(result_dir, "GlobalWords_topN.json.gz"), mode="w",
                                      compresslevel=9)

    with open(join(result_dir, "GlobalWords.json"), "r", encoding='utf-8') as file:
        global_words_dict = json.load(file)
        global_words_dict = _get_global_words(global_words_dict)
        global_word_top_n.write(json.dumps(_top_n_getter(global_words_dict, int(n))).encode('utf-8'))
        global_word_top_n.close()
        file.close() #overzealous

    if delete:
        os.remove(join(result_dir, "GlobalWords.json"))


def top_n_topic_words_extractor(result_dir: str, n, delete: bool):
    """
    top_n_Global_Words_Extractor given the result dir compute the n most frequent word in GlobalWord.
    After processing, original file is deleted if delete is true
    :param result_dir: result dir path
    :param n: amount of most important words to calculate
    :param delete: if true after processing original file is deleted
    """
    global_word_top_n = gzip.GzipFile(filename=join(result_dir, "GlobalTopicsWords_topN.json.gz"), mode="w",
                                      compresslevel=9)

    global_topic = open(join(result_dir, "GlobalTopicsWords.json"), "r", encoding='utf-8')

    global_topic_iter = iter(global_topic.readline, "")

    n = int(n)

    counter = 0
    for line in global_topic_iter:
        if line == "}":
            break

        if line[:1] != "{":
            line = "{" + line

        line = line[:len(line) - 2] + "}"

        topic_dict = json.loads(line)

        for topic in topic_dict:
            top_words = {topic: _top_n_getter(topic_dict[topic], n)}

        if counter == 0:
            page_json = json.dumps(top_words)
            page_json = page_json[:len(page_json) - 1] + ",\n"
            global_word_top_n.write(page_json.encode('utf-8'))
        elif counter >= 0:
            page_json = json.dumps(top_words)
            page_json = page_json[1:len(page_json) - 1] + ",\n"
            global_word_top_n.write(page_json.encode('utf-8'))
        global_word_top_n.flush()
        counter += 1

    global_word_top_n.write("}".encode('utf-8'))
    global_word_top_n.close()
    global_topic.close()

    if delete:
        os.remove(join(result_dir, "GlobalTopicsWords.json"))


def _get_min_word_length(lang):  # Returns min admitted word length for the language (sync in textnormalizer)
    if lang in ["gan", "ja", "ko", "vi",  "wuu", "zh", "zh-classical", "zh-yue"]:
        return 1  # Hang, Hans, Hant scripts
    elif lang == "vi":
        return 2  # Hybrid case of Chu Nom in Latn
    else:
        return 3


def main():
    global MIN_WORD_LENGTH
    MIN_WORD_LENGTH = _get_min_word_length(sys.argv[5])
    top_n_words_page_extractor(sys.argv[1], sys.argv[2], True)
    top_n_global_words_extractor(sys.argv[1], sys.argv[3], True)
    top_n_topic_words_extractor(sys.argv[1], sys.argv[4], True)

if __name__ == "__main__":
    main()