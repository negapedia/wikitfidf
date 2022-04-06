#!/usr/bin/env python
# cython: language_level=3
# -*- coding: utf-8 -*-

# IF YOU MODIFY THIS FILE, YOU NEED TO RUN "go generate" IN "assets" FOR CHANGES TO TAKE EFFECT.

import glob
import json
import os
import shutil
import sys
import datetime
from multiprocessing import Pool, cpu_count
import subprocess
import nltk
import spacy
from itertools import zip_longest


nltk.download('punkt')
from nltk.corpus import stopwords
from nltk.stem import SnowballStemmer, ISRIStemmer
from nltk.tokenize import word_tokenize

MIN_WORD_LENGTH = 1  # Min lenght for words (might be changed in-program wrt language)
MAX_WORD_LENGTH = 33  # Max lenght for words
ALLOWED_POS = ["ADJ", "ADV", "NOUN", "PROPN", "VERB"]  # Allowed Part Of Speech tags
STOPWORDS = []

FORBIDDEN_HTML_WORDS = ["colspan", "colspan=", "style", "style=", "https", "http"]  # not needed in new spacy flow
FORBIDDEN_WORDS = ["file", "isbn", "noeditsection", "rowspan", "colspan", "br", "en"]  # words leaked by wiki markup


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
    if lang in ["co", "eml", "fur", "lij", "lmo", "nap", "pms", "sc", "scn", "roa-tara", "vec"]:
        return set(stopwords.words(_nltk_lang_to_name("it")) + \
            stopwords.words(_nltk_lang_to_name("en")) + FORBIDDEN_WORDS)
    
    stoplang = _nltk_lang_to_name(lang)
    if stoplang:
        return set(stopwords.words(stoplang) + \
            stopwords.words(_nltk_lang_to_name("en")) + FORBIDDEN_WORDS)
    else:
        return set(stopwords.words(_nltk_lang_to_name("en")) + FORBIDDEN_WORDS)


def _stopwords_cleaner(revert_text):
    return [word for word in revert_text if not (word.lower() in STOPWORDS)]


def _words_cleaner(revert_text):
    return [word for word in revert_text \
        if not (word.lower() in STOPWORDS) and (MAX_WORD_LENGTH >= len(word) >= MIN_WORD_LENGTH)]


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


def _get_min_word_length(lang):  # Returns min admitted word length for the language (sync in topwordspageextractor)
    if lang in ["gan", "ja", "ko", "vi",  "wuu", "zh", "zh-classical", "zh-yue"]:
        return 1  # Hang, Hans, Hant scripts
    elif lang == "vi":
        return 2  # Hybrid case of Chu Nom in Latn
    else:
        return 3


def _async_delete_dir_content(the_dir: str):
    """
    Fast-delete in parallel a directory
    (NOTE: by design choice, no guarantee this ends before the calling program, might stay orphaned until completion)
    :param the_dir: directory to delete
    """
    empty_dir = os.path.join(os.path.dirname(the_dir), "empty_dir")
    os.makedirs(empty_dir, exist_ok=True)
    empty_dir = os.path.join(empty_dir, "")  # add a trailing slash if not present
    the_dir = os.path.join(the_dir, "")  # ditto
    subprocess.Popen("rsync -a --delete " + empty_dir + " " + the_dir + \
        " ; rmdir " + empty_dir + " " + the_dir, \
        shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)


def _words_extractor(input_dir: str, output_dir: str, o_process: int, parallelism: int, lang: str, log_file: str):
    """
    _words_extractor perform tokenization, stopwords cleaning and lemmatization on a single file "filename"
    :param result_dir: path of input folder
    :param output_dir: path of output folder
    :param o_process: process ordinal (in range(parallelism))
    :param parallelism: degree of parallelism
    :param lang: wikipedia language
    :log_file: file for async non-blocking logging
    """

    with open(log_file, "w", encoding='utf-8') as logger:  # Non-blocking async logger
        (nlp, lemmatable) = _get_nlp_processor(lang)

        bsize = 1000 // parallelism
        n_first_bucket = o_process * bsize
        n_last_bucket = 1000 if (o_process == parallelism -1) else (o_process + 1) * bsize

        for file in os.scandir(input_dir):
            if n_first_bucket <= int(file.name[-8:-5]) < n_last_bucket:
                with open(file.path, "r", encoding='utf-8') as the_file:
                    dump_dict = json.load(the_file)
                    the_file.flush()  # overzealous

                reverse_stemming_dict = {}

                reverts_texts = []
                reverts_length = 0
                for reverts in dump_dict["Revision"]:
                    single_revert = reverts["Text"]
                    reverts_length += len(single_revert)
                    if reverts_length > 1000000:  # spacy limit (cf. max_length)
                        logger.write(f"{datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')} Reverts overflow (reaching {reverts_length} chars) with {len(single_revert)} chars of file: {file.name}\n")
                        logger.flush()
                        break
                    reverts_texts.append(single_revert)
                
                multidoc = nlp.pipe(reverts_texts)
                for reverts, doc in zip_longest(dump_dict["Revision"], multidoc, fillvalue=[]):
                    if reverts["Text"] is None:
                        continue
                    if lemmatable:
                        reverts["Text"] = [(w.lemma_ if w.pos_ == "PROPN" else w.norm_) \
                                            for w in doc if (w.pos_ in ALLOWED_POS and w.is_alpha)]
                    else:
                        reverts["Text"] = [w.lower_ for w in doc if w.is_alpha]

                    reverts["Text"] = _words_cleaner(reverts["Text"])

                    if not lemmatable:
                        reverts["Text"], reverse_stemming_dict = _stemming(reverts["Text"], reverse_stemming_dict, lang)

                page_id = dump_dict["PageID"]

                with open(os.path.join(output_dir, "S" + "0" * (20 - len(str(page_id))) + str(page_id) + ".json"), "w", encoding='utf-8') as file:
                    json.dump(dump_dict, file, ensure_ascii=False)
                    file.flush()  # overzealous
                with open(os.path.join(output_dir, "Stem/StemRev_" + str(page_id) + ".json"), "w", encoding='utf-8') as file:
                    json.dump(reverse_stemming_dict, file, ensure_ascii=False)
                    file.flush()  # overzealous
        logger.flush()  # overzealous


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
        reverts["Text"] = _stopwords_cleaner(reverts["Text"])

        reverts["Text"], reverse_stemming_dict = _stemming(reverts["Text"], reverse_stemming_dict, lang)

    page_id = dump_dict["PageID"]

    os.remove(filename)
    with open(os.path.join(result_dir, "S" + "0" * (20 - len(str(page_id))) + str(page_id) + ".json"), "w", encoding='utf-8') as file:
        json.dump(dump_dict, file, ensure_ascii=False)
        file.flush()  # overzealous
    with open(os.path.join(result_dir, "Stem/StemRev_" + str(page_id) + ".json"), "w", encoding='utf-8') as file:
        json.dump(reverse_stemming_dict, file, ensure_ascii=False)
        file.flush()  # overzealous


def async_error_logger(e):  # The show must go on.
    with open(str(os.getpid()) + "-error.log", "a", encoding='utf-8') as error_log:
        error_log.write(f"ERROR at time {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')} :")
        error_log.write("[[[{}]]]".format(e.__cause__))
        error_log.flush()  # overzealous


def concurrent_stopwords_cleaner_lemmatizer(result_dir: str, lang: str):
    """
    The method given the result dir, perform in parallel tokenization, stopwords cleaning, lemmatization
    :param result_dir: path of result folder
    :param lang: wiki language
    """
    global MIN_WORD_LENGTH
    global STOPWORDS

    MIN_WORD_LENGTH = _get_min_word_length(lang)
    STOPWORDS = _lang_stopwords(lang)

    log_prefix = "/data/normalization_"
    input_dir = result_dir + "_input"
    shutil.rmtree(input_dir, ignore_errors=True)
    shutil.move(result_dir, input_dir)
    os.mkdir(result_dir)
    shutil.move(os.path.join(input_dir, "Stem"), result_dir)

    parallelism = max(1, cpu_count() - 1)
    executor = Pool(parallelism)
    for i in range(parallelism):
        log_file = log_prefix + str(i) + ".log"
        executor.apply_async(_words_extractor, \
            args=(input_dir, result_dir, i, parallelism, lang, log_file), \
                error_callback = async_error_logger)
    executor.close()
    executor.join()
    _async_delete_dir_content(input_dir)
    for i in range(parallelism):
        log_file = log_prefix + str(i) + ".log"
        if os.path.exists(log_file) and os.path.getsize(log_file) == 0:
            os.remove(log_file)


def concurrent_stopwords_cleaner_stemmer(result_dir: str, lang: str):
    """
    The method given the result dir, perform in parallel tokenization, stopwords cleaning, stemming
    :param result_dir: path of result folder
    :param lang: wiki language
    """

    (nlp, lemmatable) = _get_nlp_processor(lang)

    file_to_clean = sorted(glob.glob(os.path.join(result_dir, "W[0-9]*.json")),
                           key=os.path.getsize)  # from the smallest to the biggest

    executor = Pool(cpu_count())
    for filename in file_to_clean:
        executor.apply_async(_stopwords_cleaner_stemming, args=(result_dir, filename, lang, nlp, lemmatable))
    executor.close()
    executor.join()


def main():
    concurrent_stopwords_cleaner_lemmatizer(sys.argv[1], sys.argv[2])


if __name__ == "__main__":
    main()
