#!/usr/bin/env python
#cython: language_level=3
# -*- coding: utf-8 -*-

#IF YOU MODIFY THIS FILE, YOU NEED TO RUN "go generate" IN "assets" FOR CHANGES TO TAKE EFFECT.

import copy
import json
import os
import sys
from os.path import join


def _write_json(dict_to_write, output_stream):
    output_stream.write(json.dumps(dict_to_write, ensure_ascii=False)[1:-1])
    output_stream.write(",\n")


def _create_result_dir_if_not_exist(result_dir):
    if not os.path.exists(result_dir):
        os.makedirs(result_dir)


def _check_file_name(file_name):
    if file_name.find(".json") != -1:
        file_name = file_name[:len(file_name) - 5]
    return file_name


def _dict_writer(dict_to_write, file_name, output_dir):
    """
    dict_writer write in .json a dictionary
    :param dict_to_write: a dictionary
    :param file_name: the file name
    :param output_dir: where to write
    """

    _create_result_dir_if_not_exist(output_dir)

    with open(join(output_dir, _check_file_name(file_name) + ".json"), "w", encoding='utf-8') as _output_file_path:
        json.dump(dict_to_write, _output_file_path, ensure_ascii=False)
        _output_file_path.flush() #overzealous


def global_page_destem(result_dir):
    """
    GlobalPageDeStem given the result dir perform the de-stemming process on GlobalPageTFIDF
    :param result_dir: path of result folder
    """
    reverse_stemming_dict = json.load(open(join(result_dir, "GlobalStem.json"), "r", encoding='utf-8'))

    if len(reverse_stemming_dict) > 0:

        global_dict_file = open(join(result_dir, "GlobalPagesTFIDF.json"), "r", encoding='utf-8')
        global_dict_file_iter = iter(global_dict_file.readline, "")

        destemmed_global_dict_file = open(join(result_dir, "DESTEM_GlobalPagesTFIDF.json"), "w", encoding='utf-8')
        destemmed_global_dict_file.write("{")

        for line in global_dict_file_iter:
            if len(line) > 1:
                line = line[:-2] + "}"
            if line[0] != "{":
                line = "{" + line
            if line == "}":
                break
            page_dict = json.loads(line)
            for page in page_dict:
                global_dict_new = {
                    page: {"TopicID": page_dict[page]["TopicID"], "Tot": page_dict[page]["Tot"], "Words": {}}}
                for word in page_dict[page]["Words"]:
                    if word in reverse_stemming_dict.keys():
                        global_dict_new[page]["Words"][reverse_stemming_dict[word]] = page_dict[page]["Words"][word]
                    else:
                        global_dict_new[page]["Words"][word] = page_dict[page]["Words"][word]
                _write_json(global_dict_new, destemmed_global_dict_file)

        global_dict_file.close()
        destemmed_global_dict_file.write("}")
        destemmed_global_dict_file.close()
        os.remove(join(result_dir, "GlobalPagesTFIDF.json"))
        os.rename(join(result_dir, "DESTEM_GlobalPagesTFIDF.json"), join(result_dir, "GlobalPagesTFIDF.json"))


def global_word_destem(result_dir):
    """
    global_word_destem given the result dir perform the de-stemming process on GlobalWord
    :param result_dir: path of result folder
    """
    reverse_stemming_dict = json.load(open(join(result_dir, "GlobalStem.json"), "r", encoding='utf-8'))
    global_dict = json.load(open(join(result_dir, "GlobalWords.json"), "r", encoding='utf-8'))

    global_dict_new = copy.deepcopy(global_dict)
    for word in global_dict:
        if word in ("@Total page", "@Grand total"):
            continue
        if word in reverse_stemming_dict.keys():
            global_dict_new[reverse_stemming_dict[word]] = global_dict[word]
            del global_dict_new[word]

    _dict_writer(global_dict_new, "GlobalWords", result_dir)


def remove_destem_file(result_dir):
    os.remove(join(result_dir, "GlobalStem.json"))
    os.removedirs(join(result_dir, "Stem/"))

def main():
    global_page_destem(sys.argv[1])
    global_word_destem(sys.argv[1])
    remove_destem_file(sys.argv[1])

if __name__ == "__main__":
    main()