#!/usr/bin/env python
# -*- coding: utf-8 -*-

import glob
import json
import os
import sys
from multiprocessing import Pool, cpu_count

from nltk.tokenize import RegexpTokenizer
from nltk.corpus import stopwords
from nltk.stem import PorterStemmer
from nltk.stem.snowball import SnowballStemmer
''' SnowballStemmer available languages:
    danish dutch english finnish french german hungarian italian
    norwegian porter portuguese romanian russian spanish swedish
'''


def _lang_mapper(lang):
    # available language for stopwords list
    available_lang = {
        "en": "english",
        "ar": "arabic",
        "da": "danish",
        "nl": "dutch",
        "fi": "finnish",
        "fr": "french",
        "de": "german",
        "el": "greek",
        "hu": "hungarian",
        "id": "indonesian",
        "it": "italian",
        "kk": "kazakh",
        "ne": "nepali",
        "no": "norwegian",
        "pt": "portuguese",
        "ro": "romanian",
        "ru": "russian",
        "es": "spanish",
        "sv": "swedish",
        "tr": "turkish",
        "hy": "armenian",
        "az": "azerbaijani",
        "eu": "basque",
        "bn": "bengali",
        "bg": "bulgarian",
        "ca": "catalan",
        "zh": "chinese",
        "sh": "croatian",
        "cs": "czech",
        "gl": "galician",
        "he": "hebrew",
        "hi": "hindi",
        "ga": "irish",
        "ja": "japanese",
        "ko": "korean",
        "lv": "latvian",
        "lt": "lithuanian",
        "mr": "marathi",
        "fa": "persian",
        "pl": "polish",
        "sk": "slovak",
        "th": "thai",
        "uk": "ukrainian",
        "ur": "urdu",
        "simple": "english",
        "vec": "italian"  # only as test
    }
    return available_lang[lang]


def _stopwords_cleaner(revert_text, lang):
    stop_words = stopwords.words(_lang_mapper(lang))
    text = revert_text
    for word in text:
        if word.lower() in stop_words:
            revert_text = list(filter(word.__ne__, revert_text))
    return revert_text


def _increment_word_counter(word_dict, word):
    if word in word_dict.keys():
        word_dict[word] += 1
    else:
        word_dict[word] = 1


def _stemming(revert_text, stemmer_reverse_dict, lang):
    stemmer = None
    try:
        stemmer = SnowballStemmer(_lang_mapper(lang))
    except Exception:
        stemmer = PorterStemmer()

    # word_counter = {}
    text = []

    for word in revert_text:
        '''stemmed_word = stemmer.stem(word)
        if stemmed_word == word: # se sono uguali
            _increment_word_counter(word_counter, word)
            if word in stemmer_reverse_dict.keys():
                if len(stemmer_reverse_dict[word]) > len(word):
                    stemmer_reverse_dict[word] = word
        else: # se sono diverse
            if stemmed_word in word_counter.keys() and stemmed_word not in stemmer_reverse_dict.keys():
                _increment_word_counter(word_counter, stemmed_word)
            else:
                _increment_word_counter(word_counter, stemmed_word)
                if stemmed_word not in stemmer_reverse_dict.keys():
                    stemmer_reverse_dict[stemmed_word] = word
                elif len(stemmer_reverse_dict[stemmed_word]) > len(word):
                    stemmer_reverse_dict[stemmed_word] = word
        '''
        stemmed_word = stemmer.stem(word)
        if stemmed_word in stemmer_reverse_dict.keys() and len(stemmer_reverse_dict[stemmed_word]) > len(word):
            stemmer_reverse_dict[stemmed_word] = word
        elif stemmed_word not in stemmer_reverse_dict.keys():
            stemmer_reverse_dict[stemmed_word] = word

        text.append(stemmed_word)
    return text, stemmer_reverse_dict


def _stopwords_cleaner_stemming(result_dir: str, filename: str, lang: str):
    """
    _stopwords_cleaner_stemming perform tokenization, stopwords cleaning and stemming on a single file "filname"
    :param result_dir: path of result folder
    :param filename: file to preocess
    :param lang: wikipedia language
    """
    with open(filename, "r") as f:
        dump_dict = json.load(f)

    reverse_stemming_dict = {}
    tokenizer = RegexpTokenizer(r'\w+')
    for reverts in dump_dict["Revision"]:
        if reverts["Text"] is None:
            continue
        reverts["Text"] = tokenizer.tokenize(reverts["Text"])
        reverts["Text"] = [word for word in reverts["Text"] if
                           not (len(word) > 20 or len(word) <= 3 or word == "https" or word == "http")]  # fixing words
        # length
        reverts["Text"] = _stopwords_cleaner(reverts["Text"], lang)

        reverts["Text"], reverse_stemming_dict = _stemming(reverts["Text"], reverse_stemming_dict)
        if reverts["Text"] is None:
            os.remove(filename)
            return

    page_id = dump_dict["PageID"]
    topic_id = dump_dict["TopicID"]

    os.remove(filename)
    with open(result_dir + "S" + str(topic_id) + "_" + str(page_id) + ".json", "w") as f:
        json.dump(dump_dict, f, ensure_ascii=False)

    with open(result_dir + "Stem/StemRev_" + str(topic_id) + "_" + str(page_id) + ".json", "w") as f:
        json.dump(reverse_stemming_dict, f, ensure_ascii=False)


def concurrent_stopwords_cleaner_stemmer(result_dir: str, lang: str):
    """
    The method given the result dir, perform in parallel tokenization, stopwords cleaning, stemming
    :param result_dir: path of result folder
    :param lang: wiki language
    """

    file_to_clean = sorted(glob.glob(result_dir + "W[1-9]*.json"),
                           key=os.path.getsize)  # from the smallest to the biggest

    executor = Pool(cpu_count())
    for filename in file_to_clean:
        executor.apply_async(_stopwords_cleaner_stemming, args=(result_dir, filename, lang))
    executor.close()
    executor.join()


if __name__ == "__main__":
    concurrent_stopwords_cleaner_stemmer(sys.argv[1], sys.argv[2])
