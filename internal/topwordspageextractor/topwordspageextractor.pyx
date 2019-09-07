#!/usr/bin/env python
# cython: language_level=3
# -*- coding: utf-8 -*-

# IF YOU MODIFY THIS FILE, YOU NEED TO RUN "go generate" IN "assets" FOR CHANGES TO TAKE EFFECT.

import gzip
import json
import os
import sys
from collections import Counter
from os.path import join


def _top_n_getter(words_dict: dict, n: int):
    top_n = Counter(words_dict).most_common(n)
    words_dict = {}
    for key, value in top_n:
        words_dict[key] = value
    return words_dict


def _get_top_n_words_pages_dict(page_dict: dict, n: int):
    words_dict_TFIDF = {}
    words_dict_Occur = {}
    for page in page_dict:
        for word in page_dict[page]["Words"]:
            words_dict_TFIDF[word] = page_dict[page]["Words"][word]["tfidf"]
            words_dict_Occur[word] = page_dict[page]["Words"][word]["abs"]

        top_n_page = {}
        if len(words_dict_TFIDF) > n:
            top_n_page[page] = {"TopicID": page_dict[page]["TopicID"], "Tot": page_dict[page]["Tot"],
                                "Word2TFIDF": _top_n_getter(words_dict_TFIDF, n),
                                "Word2Occur": _top_n_getter(words_dict_Occur, n)}
        else:
            top_n_page[page] = {"TopicID": page_dict[page]["TopicID"], "Tot": page_dict[page]["Tot"],
                                "Word2TFIDF": words_dict_TFIDF, "Word2Occur": words_dict_Occur}

        return top_n_page


def _get_global_words(global_dict: dict):
    new_global_dict = {"@TOTAL Words": global_dict["Total Word"]}
    for word in global_dict:
        if word == "@Total Word" or word == "@Total Page":
            continue
        new_global_dict[word] = global_dict[word]["a"]

    return new_global_dict


def _words_counter(words_dic: dict):
    counter = 0
    for word in words_dic:
        counter += words_dic[word]
    return counter


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

    gloabal_tfidf = open(join(result_dir, "GlobalPagesTFIDF.json"), "r")
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
            global_top_ntfidf.write(page_json.encode())
        elif counter >= 0:
            page_json = json.dumps(page_dict)
            page_json = page_json[1:len(page_json) - 1] + ",\n"
            global_top_ntfidf.write(page_json.encode())
        global_top_ntfidf.flush()
        counter += 1

    global_top_ntfidf.write("}".encode())
    global_top_ntfidf.flush()
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

    global_words_dict = json.load(open(join(result_dir, "GlobalWords.json"), "r"))

    global_words_dict = _get_global_words(global_words_dict)
    global_word_top_n.write(json.dumps(_top_n_getter(global_words_dict, int(n) + 1)).encode())
    global_word_top_n.flush()
    global_word_top_n.close()

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

    global_topic = open(join(result_dir, "GlobalTopicsWords.json"), "r")

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
            top_n_words = _top_n_getter(topic_dict[topic], n)
            top_n_words["@TOT"] = _words_counter(topic_dict[topic])

        if counter == 0:
            page_json = json.dumps(top_n_words)
            page_json = page_json[:len(page_json) - 1] + ",\n"
            global_word_top_n.write(page_json.encode())
        elif counter >= 0:
            page_json = json.dumps(top_n_words)
            page_json = page_json[1:len(page_json) - 1] + ",\n"
            global_word_top_n.write(page_json.encode())
        global_word_top_n.flush()
        counter += 1

    global_word_top_n.write("}".encode())
    global_word_top_n.flush()
    global_word_top_n.close()
    global_topic.close()

    if delete:
        os.remove(join(result_dir, "GlobalTopicsWords.json"))


def top_n_global_badwords_extractor(result_dir: str, n):
    badw = gzip.open(join(result_dir, "BadWordsReport.json.gz"), "r")
    badw_iter = iter(badw.readline, "")

    global_badwords_dict = {}
    tot = 0
    for line in badw_iter:
        line = line.decode()
        if line == "}":
            break

        if line[:1] != "{":
            line = "{" + line

        line = line[:len(line) - 2] + "}"

        page_dict = json.loads(line)
        badw_dict = {}
        for page in page_dict:
            for word in page_dict[page]["BadW"]:
                tot += page_dict[page]["BadW"][word]
                if word in global_badwords_dict.keys():
                    global_badwords_dict[word] += page_dict[page]["BadW"][word]
                else:
                    global_badwords_dict[word] = page_dict[page]["BadW"][word]

    global_badwords_dict["@TOTAL Words"] = tot
    global_badword_top_n = gzip.GzipFile(filename=join(result_dir, "GlobalBadWords_topN.json.gz"), mode="w",
                                         compresslevel=9)
    global_badword_top_n.write(json.dumps(_top_n_getter(global_badwords_dict, int(n) + 1)).encode())
    global_badword_top_n.flush()
    global_badword_top_n.close()
    badw.close()


def top_n_topic_badwords_extractor(result_dir: str, n):
    badw = gzip.open(join(result_dir, "TopicBadWords.json.gz"), "r")
    badw_iter = iter(badw.readline, "")

    global_badword_top_n = gzip.GzipFile(filename=join(result_dir, "TopicBadWords_topN.json.gz"), mode="w",
                                         compresslevel=9)

    top_n_dict = {}
    for line in badw_iter:
        line = line.decode()
        if line == "}":
            break

        if line[:1] != "{":
            line = "{" + line

        line = line[:len(line) - 2] + "}"

        topic_dict = json.loads(line)  # map[uint32]structures.TopicBadWords
        for topic in topic_dict:
            badw_dict = topic_dict[topic]["BadW"]
            tot = topic_dict[topic]["TotBadw"]
            top_n_dict[topic] = {"TotBadw": tot, "Badwords": _top_n_getter(badw_dict, int(n))}

    global_badword_top_n.write(json.dumps(top_n_dict).encode())
    global_badword_top_n.flush()
    global_badword_top_n.close()
    badw.close()


def main():
    top_n_words_page_extractor(sys.argv[1], sys.argv[2], True)
    top_n_global_words_extractor(sys.argv[1], sys.argv[3], True)
    top_n_global_badwords_extractor(sys.argv[1], sys.argv[3])  # like globalWords
    top_n_topic_badwords_extractor(sys.argv[1], sys.argv[4])  # like topic
    top_n_topic_words_extractor(sys.argv[1], sys.argv[4], True)


if __name__ == "__main__":
    main()
