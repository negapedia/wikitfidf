#!/usr/bin/env python
# -*- coding: utf-8 -*-

import string, json, sys, glob, os
import mwparserfromhell as mwparser  # clone and install from: `https://github.com/earwig/mwparserfromhell/` or `pip3 install mwparserfromhell`
from multiprocessing import Pool, cpu_count


def _fix_text(dict_text):
    """
    Il metodo prende in input del media wiki text e lo pulisce, ritorna un array
    :param text:
    :return: list of pure text
    """
    text = mwparser.parse(dict_text)
    wikilinks = text.ifilter_wikilinks(recursive=False)
    templates = text.ifilter_templates(recursive=False)
    headings = text.ifilter_headings(recursive=False)
    external_links = text.ifilter_external_links(recursive=False)
    if wikilinks:
        for w in wikilinks: text.remove(w, recursive=False)
    if templates:
        for t in templates: text.remove(t, recursive=False)
    if headings:
        for h in headings: text.remove(h, recursive=False)
    if external_links:
        for e in external_links: text.remove(e, recursive=False)

    fixed_text = text.strip_code()
    fixed_text = fixed_text.replace("[", "")
    fixed_text = fixed_text.replace("]", "")
    fixed_text = fixed_text.replace("REDIRECT", "")
    fixed_text = fixed_text.replace("redirect", "")
    fixed_text = fixed_text.replace(".", "")
    fixed_text = fixed_text.replace(",", "")
    fixed_text = fixed_text.replace(";", "")
    fixed_text = fixed_text.replace("'", " ")

    symbols_to_remove = string.punctuation
    symbols_to_remove += "’"
    symbols_to_remove += "–"
    symbols_to_remove += "°"
    table = str.maketrans({key: None for key in symbols_to_remove})
    fixed_text = fixed_text.translate(table)

    if fixed_text == " ":
        return None

    fixed_text = fixed_text.lower()
    return fixed_text


def _dict_text_correction(filename: str):
    print(filename)
    with open(filename, "r") as f:
        dump_dict = json.load(f)

    if dump_dict["Revision"] is None:
        os.remove(filename)
        return

    for revision in dump_dict["Revision"]:
        if revision["Text"] is not None:
            revision["Text"] = _fix_text(revision["Text"])

    page_id = dump_dict["PageID"]
    result_dir = filename[:-(len(page_id)+5)]

    os.remove(filename)
    with open(result_dir+"W"+page_id+".json", "w") as f:
        json.dump(dump_dict, f, ensure_ascii=False)


def concurrent_wiki_markup_cleaner(result_dir: str):
    """
    The method given the reduced dump, clean the dump from wikipedia markup calling _fix_text(...)
    :param dump_dict: reduced dict of the dump
    :return: the cleaned up dump
    """

    file_to_clean = sorted(glob.glob(result_dir+"[1-9]*.json"), key=os.path.getsize)  # from the smallest to the biggest

    with Pool(cpu_count()) as executor:
        executor.map(_dict_text_correction, file_to_clean)



if __name__ == "__main__":
    #dict_text_correction("../../Result/it_20190601/")
    concurrent_wiki_markup_cleaner(sys.argv[1])

