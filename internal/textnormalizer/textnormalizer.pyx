#!/usr/bin/env python
# cython: language_level=3
# -*- coding: utf-8 -*-

# IF YOU MODIFY THIS FILE, YOU NEED TO RUN "go generate" IN "assets" FOR CHANGES TO TAKE EFFECT.

import glob
import json
import os
import sys
from multiprocessing import Pool, cpu_count
from os.path import join
from tracemalloc import stop
import nltk
import spacy
""" @@debug
import time
import os.path
"""


nltk.download('punkt')
from nltk.corpus import stopwords
from nltk.stem import SnowballStemmer, ISRIStemmer
from nltk.tokenize import word_tokenize

MIN_WORD_LENGTH = 1  # Min lenght for words (might be changed in-program wrt language)
MAX_WORD_LENGTH = 40  # Max lenght for words
ALLOWED_POS = ["ADJ", "ADV", "NOUN", "PROPN", "VERB"]  # Allowed Part Of Speech tags

# FORBIDDEN_HTML_WORDS = ["colspan", "colspan=", "style", "style=", "https", "http"]
FORBIDDEN_HTML_WORDS = []  # @@@ blank for testing


def _nltk_lang_to_name(lang):
    lang_names = {
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
        "simple": "english"
    }
    if lang in lang_names:
        return lang_names[lang]
    else:
        return False


def _lang_stopwords(lang):
    if lang in ["eml", "fur", "lij", "lmo", "nap", "pms", "sc", "scn", "roa-tara", "vec"]:
        return stopwords.words(_nltk_lang_to_name("it"))
    elif lang in ["gd", "sco"]:
        return stopwords.words(_nltk_lang_to_name("en"))
    
    stoplang = _nltk_lang_to_name(lang)
    if stoplang:
        return stopwords.words(stoplang)
    else:
        return stopwords.words(_nltk_lang_to_name("en"))


def _stopwords_cleaner(revert_text, lang):
    stop_words = _lang_stopwords(lang)
    text = revert_text
    for word in text:
        if word.lower() in stop_words:
            revert_text = list(filter(word.__ne__, revert_text))
    return revert_text


def _words_cleaner(revert_text, lang):
    stop_words = _lang_stopwords(lang) + FORBIDDEN_HTML_WORDS
    text = revert_text
    for word in text:
        if (word.lower() in stop_words) or not (MAX_WORD_LENGTH >= len(word) >= MIN_WORD_LENGTH):
            revert_text = list(filter(word.__ne__, revert_text))
    return revert_text


def _increment_word_counter(word_dict, word):
    if word in word_dict.keys():
        word_dict[word] += 1
    else:
        word_dict[word] = 1


def _get_stemmer(lang):
    if lang in ["en", "da", "nl", "fr", "de", "es", "hu", "it", "simple", "no", "pt", "ro", "ru", "sv"]:
        # N.B. for portuguese (pt) is also available RSLPStemmer
        return SnowballStemmer(_nltk_lang_to_name(lang))
    elif lang == "ar":
        return ISRIStemmer()
    else:
        # if here, there not exits a satisfiable stemmer for the language, so
        # it is returned None, which means that the process of stemming must be skipped
        return None


def _stemming(revert_text, stemmer_reverse_dict, lang):
    stemmer = _get_stemmer(lang)
    if stemmer is None:
        return revert_text, {}
    text = []

    for word in revert_text:
        stemmed_word = stemmer.stem(word)
        print("%s --> %s" % (word, stemmed_word))
        if stemmed_word in stemmer_reverse_dict.keys() and len(word) < len(stemmer_reverse_dict[stemmed_word]):
            stemmer_reverse_dict[stemmed_word] = word
        elif stemmed_word not in stemmer_reverse_dict.keys():
            stemmer_reverse_dict[stemmed_word] = word

        text.append(stemmed_word)
    return text, stemmer_reverse_dict


def _get_tokenizer_lang(lang):
    if _nltk_lang_to_name(lang) not in ['czech', 'danish', 'dutch', 'english', 'estonian',
                                  'finnish', 'french', 'german', 'greek', 'italian',
                                  'norwegian', 'polish', 'portuguese', 'slovene',
                                  'spanish', 'swedish']:
        # language not supported by nltk.tokenizers, so use by default english
        return "english"
    else:
        return _nltk_lang_to_name(lang)


def _get_nlp_processor(lang):  # Returns nlp processor and lemmatization capability (True/False)
    if lang == "en" or lang == "simple":
        return (spacy.load("en_core_web_sm", exclude=["parser", "ner", "textcat", "custom"]), True)
    elif lang in ["ca", "da", "de", "el", "es", "fr", "it", "ja", "lt", "nl", "pl", "pt", "ro", "ru", "zh"]:
        return (spacy.load(lang + "_core_news_sm", exclude=["parser", "ner", "textcat", "custom"]), True)
    elif lang == "no":
        return (spacy.load("nb_core_news_sm", exclude=["parser", "ner", "textcat", "custom"]), True)
    elif lang in ["eml", "fur", "lij", "la", "lmo", "nap", "pms", "sc", "scn", "roa-tara", "vec"]:
        return (spacy.blank("it"), False)
    else:  # fallback case (multilingual)
        return (spacy.blank("xx"), False)


def _get_min_word_length(lang):  # Returns min admitted word length for the language
    if lang in ["gan", "ja", "ko", "vi",  "wuu", "zh", "zh-classical", "zh-yue"]:
        return 1  # Hang, Hans, Hant scripts
    elif lang == "vi":
        return 2  # Hybrid case of Chu Nom in Latn
    else:
        return 3


def _words_extractor(result_dir: str, filename: str, lang: str, nlp, lemmatable: bool):
    """
    _words_extractor perform tokenization, stopwords cleaning and lemmatization on a single file "filename"
    :param result_dir: path of result folder
    :param filename: file to preocess
    :param lang: wikipedia language
    :param nlp: the NLP processor to be used
    :lemmatable: flag indicating whether the nlp processor supports lemmatization for lang
    """
    with open(filename, "r", encoding='utf-8') as the_file:
        dump_dict = json.load(the_file)
        the_file.flush()  # overzealous

    reverse_stemming_dict = {}

    reverts_texts = []
    for reverts in dump_dict["Revision"]:
        reverts_texts.append(reverts["Text"])
    
    multidoc = nlp.pipe(reverts_texts)
    for reverts, doc in zip(dump_dict["Revision"], multidoc):
        if reverts["Text"] is None:
            continue
        if lemmatable:
            reverts["Text"] = [(w.lemma_ if w.pos_ == "PROPN" else w.norm_) for w in doc if (w.pos_ in ALLOWED_POS and w.is_alpha)]
        else:
            reverts["Text"] = [w.lower_ for w in doc if w.is_alpha]

        reverts["Text"] = _words_cleaner(reverts["Text"], lang)

        if not lemmatable:
            reverts["Text"], reverse_stemming_dict = _stemming(reverts["Text"], reverse_stemming_dict, lang)

    page_id = dump_dict["PageID"]
    #topic_id = dump_dict["TopicID"]

    os.remove(filename)  # @@@ debug
    with open(join(result_dir, "S" + "0" * (20 - len(str(page_id))) + str(page_id) + ".json"), "w", encoding='utf-8') as file:
        json.dump(dump_dict, file, ensure_ascii=False)
        file.flush()  # overzealous for debug
    with open(join(result_dir, "Stem/StemRev_" + str(page_id) + ".json"), "w", encoding='utf-8') as file:
        json.dump(reverse_stemming_dict, file, ensure_ascii=False)
        file.flush()  # overzealous for debug


def _stopwords_cleaner_stemming(result_dir: str, filename: str, lang: str):
    """
    _stopwords_cleaner_stemming perform tokenization, stopwords cleaning and stemming on a single file "filname"
    :param result_dir: path of result folder
    :param filename: file to preocess
    :param lang: wikipedia language
    """
    with open(filename, "r", encoding='utf-8') as the_file:
        dump_dict = json.load(the_file)
        the_file.flush()  # overzealous

    reverse_stemming_dict = {}

    # tokenizer = RegexpTokenizer(r'\w+')

    for reverts in dump_dict["Revision"]:
        if reverts["Text"] is None:
            continue
        reverts["Text"] = word_tokenize(reverts["Text"], language=_get_tokenizer_lang(lang))
        '''reverts["Text"] = [word for word in reverts["Text"] if
                           not (len(word) > 20 or len(word) <= 3 or word == "https" or word == "http")]  # fixing words '''
        reverts["Text"] = [word for word in reverts["Text"] if
                           not ((len(word) <= 3) or (word in FORBIDDEN_HTML_WORDS) or (len(word) > 50))]

        # length
        reverts["Text"] = _stopwords_cleaner(reverts["Text"], lang)

        reverts["Text"], reverse_stemming_dict = _stemming(reverts["Text"], reverse_stemming_dict, lang)

    page_id = dump_dict["PageID"]
    topic_id = dump_dict["TopicID"]

    os.remove(filename)
    with open(join(result_dir, "S" + "0" * (20 - len(str(page_id))) + str(page_id) + ".json"), "w", encoding='utf-8') as file:
        json.dump(dump_dict, file, ensure_ascii=False)
        file.flush()  # overzealous for debug
    with open(join(result_dir, "Stem/StemRev_" + str(page_id) + ".json"), "w", encoding='utf-8') as file:
        json.dump(reverse_stemming_dict, file, ensure_ascii=False)
        file.flush()  # overzealous for debug


def concurrent_stopwords_cleaner_lemmatizer(result_dir: str, lang: str):
    """
    The method given the result dir, perform in parallel tokenization, stopwords cleaning, lemmatization
    :param result_dir: path of result folder
    :param lang: wiki language
    """

    MIN_WORD_LENGTH = _get_min_word_length(lang)
    (nlp, lemmatable) = _get_nlp_processor(lang)

    executor = Pool(cpu_count())
    for filename in glob.iglob(join(result_dir, "W*")):
        executor.apply_async(_words_extractor, args=(result_dir, filename, lang, nlp, lemmatable))
    executor.close()
    executor.join()


def concurrent_stopwords_cleaner_stemmer(result_dir: str, lang: str):
    """
    The method given the result dir, perform in parallel tokenization, stopwords cleaning, stemming
    :param result_dir: path of result folder
    :param lang: wiki language
    """

    (nlp, lemmatable) = _get_nlp_processor(lang)

    file_to_clean = sorted(glob.glob(join(result_dir, "W[0-9]*.json")),
                           key=os.path.getsize)  # from the smallest to the biggest

    executor = Pool(cpu_count())
    for filename in file_to_clean:
        executor.apply_async(_stopwords_cleaner_stemming, args=(result_dir, filename, lang, nlp, lemmatable))
    executor.close()
    executor.join()


def main():
    concurrent_stopwords_cleaner_lemmatizer(sys.argv[1], sys.argv[2])


if __name__ == "__main__":
    """ @@@ debug
    while not os.path.exists("/data/resume"):
        time.sleep(10)
    """
    main()
