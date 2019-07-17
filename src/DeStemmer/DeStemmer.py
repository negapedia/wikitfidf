#!/usr/bin/env python
# -*- coding: utf-8 -*-

import json, os, copy, codecs


def write_json(dict_to_write, output_stream):
    output_stream.write(json.dumps(dict_to_write, ensure_ascii=False)[1:-1])
    output_stream.write(",\n")


def _create_result_dir_if_not_exist(result_dir):
    if not os.path.exists(result_dir):
        os.makedirs(result_dir)


def _check_file_name(file_name):
    if file_name.find(".json") != -1:
        file_name = file_name[:len(file_name)-5]
    return file_name


def dict_writer(dict_to_write, file_name, output_dir):
    """
    The method write in .json a dictionary
    :param dict_to_write: a dictionary
    :param file_name: the file name
    :param output_dir: where to write
    """

    _create_result_dir_if_not_exist(output_dir)

    print("Writing ", file_name, " start...")
    _output_file_path = output_dir + _check_file_name(file_name) + ".json"
    with codecs.open(_output_file_path, "w", encoding='utf-8') as output:
        json.dump(dict_to_write, output, ensure_ascii=False)
    print("Writing output end.")


def GlobalPageDeStem(result_dir):
    with open(result_dir+"GlobalStem.json", "r") as rev_stem_file:
        reverse_stemming_dict = json.load(rev_stem_file)

    if len(reverse_stemming_dict) > 0:

        global_dict_file = open(result_dir+"GlobalPageTFIDF.json", "r")
        global_dict_file_iter = iter(global_dict_file.readline, "")

        destemmed_global_dict_file = open(result_dir + "DESTEM_GlobalPageTFIDF.json", "w")
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
                global_dict_new = {page: {"Title": page_dict[page]["Title"], "Tot": page_dict[page]["Tot"], "Words": {}}}
                for word in page_dict[page]["Words"]:
                    if word in reverse_stemming_dict.keys():
                        global_dict_new[page]["Words"][reverse_stemming_dict[word]] = page_dict[page]["Words"][word]
                    else:
                        global_dict_new[page]["Words"][word] = page_dict[page]["Words"][word]
                write_json(global_dict_new, destemmed_global_dict_file)

        global_dict_file.close()
        destemmed_global_dict_file.write("}")
        destemmed_global_dict_file.close()
        os.remove(result_dir+"GlobalPageTFIDF.json")
        os.rename(result_dir+"DESTEM_GlobalPageTFIDF.json", result_dir+"GlobalPageTFIDF.json")


def GlobalWordDeStem(result_dir):
    with open(result_dir+"GlobalStem.json", "r") as rev_stem_file:
        reverse_stemming_dict = json.load(rev_stem_file)

    with open(result_dir+"GlobalWord.json", "r") as global_dict_file:
        global_dict = json.load(global_dict_file)

    global_dict_new = copy.deepcopy(global_dict)
    for word in global_dict:
        if word in ("@Total page", "@Grand total"):
            continue
        if word in reverse_stemming_dict.keys():
            global_dict_new[reverse_stemming_dict[word]] = global_dict[word]
            del global_dict_new[word]

    dict_writer(global_dict_new, "GlobalWord", result_dir)