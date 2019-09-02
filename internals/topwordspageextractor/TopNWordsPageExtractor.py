#!/usr/bin/env python
# -*- coding: utf-8 -*-

import gzip
import json
import os
from collections import Counter


def _top_n_getter(words_dict: dict, n: int):
    top_n = Counter(words_dict).most_common(n)
    words_dict = {}
    for key, value in top_n:
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
        if word == "@Total Word" or word == "@Total Page":
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
    global_top_ntfidf = gzip.GzipFile(filename=result_dir + "GlobalPagesTFIDF_top" + n + ".json.gz", mode="w",
                                      compresslevel=9)

    gloabal_tfidf = open(result_dir + "GlobalPagesTFIDF.json", "r")
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
            global_top_ntfidf.write(page_json)
        elif counter >= 0:
            page_json = json.dumps(page_dict)
            page_json = page_json[1:len(page_json) - 1] + ",\n"
            global_top_ntfidf.write(page_json)
        global_top_ntfidf.flush()
        counter += 1

    global_top_ntfidf.write("}")
    global_top_ntfidf.flush()
    global_top_ntfidf.close()
    gloabal_tfidf.close()

    if delete:
        os.remove(result_dir + "GlobalPagesTFIDF.json")


def top_n_global_words_extractor(result_dir: str, n, delete: bool):
    """
    top_n_Global_Words_Extractor given the result dir compute the n most frequent word in GlobalWord.
    After processing, original file is deleted if delete is true
    :param result_dir: result dir path
    :param n: amount of most important words to calculate
    :param delete: if true after processing original file is deleted
    """

    global_word_top_n = gzip.GzipFile(filename=result_dir + "GlobalWords_top" + n + ".json.gz", mode="w",
                                      compresslevel=9)

    with open(result_dir + "GlobalWords.json", "r") as gloabal_words:
        global_words_dict = json.load(gloabal_words)

    global_words_dict = _get_global_words(global_words_dict)
    global_word_top_n.write(json.dumps(_top_n_getter(global_words_dict, int(n))).encode())
    global_word_top_n.flush()
    global_word_top_n.close()

    if delete:
        os.remove(result_dir + "GlobalWords.json")


def top_n_topic_words_extractor(result_dir: str, n, delete: bool):
    """
    top_n_Global_Words_Extractor given the result dir compute the n most frequent word in GlobalWord.
    After processing, original file is deleted if delete is true
    :param result_dir: result dir path
    :param n: amount of most important words to calculate
    :param delete: if true after processing original file is deleted
    """
    global_word_top_n = gzip.GzipFile(filename=result_dir + "GlobalTopicsWords_top" + n + ".json.gz", mode="w",
                                      compresslevel=9)

    global_topic = None
    try:
        global_topic = open(result_dir + "GlobalTopicsWords.json", "r")
    except IOError:
        global_word_top_n.write("Error while opening GlobalTopicsWords.json!")
        global_word_top_n.flush()
        global_word_top_n.close()

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
            global_word_top_n.write(page_json.encode())
        elif counter >= 0:
            page_json = json.dumps(top_words)
            page_json = page_json[1:len(page_json) - 1] + ",\n"
            global_word_top_n.write(page_json.encode())
        global_word_top_n.flush()
        counter += 1

    global_word_top_n.write("}".encode())
    global_word_top_n.flush()
    global_word_top_n.close()
    global_topic.close()

    if delete:
        os.remove(result_dir + "GlobalTopicsWords.json")
