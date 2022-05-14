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
import time
import multiprocessing
import subprocess
import signal
import gc
import nltk
import spacy
from itertools import zip_longest

nltk.download('punkt')
from nltk.corpus import stopwords
from nltk.stem import SnowballStemmer, ISRIStemmer
from nltk.tokenize import word_tokenize

MIN_WORD_LENGTH = 1  # Min lenght for words (default: might be changed in-program wrt language)
MAX_WORD_LENGTH = 33  # Max lenght for words
ALLOWED_POS = ["ADJ", "ADV", "NOUN", "PROPN", "VERB"]  # Allowed Part Of Speech tags
STOPWORDS = []

FORBIDDEN_HTML_WORDS = ["colspan", "colspan=", "style", "style=", "https", "http"]  # not needed in new spacy flow
FORBIDDEN_WORDS = ["file", "isbn", "noeditsection", "rowspan", "colspan", "br", "en"]  # words leaked by wiki markup


def _nltk_lang_to_name(lang: str):
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


def _lang_stopwords(lang: str):
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


def _get_stemmer(lang: str):
    if lang in ["en", "da", "nl", "fr", "de", "es", "hu", "it", "simple", "no", "pt", "ro", "ru", "sv"]:
        # N.B. for portuguese (pt) is also available RSLPStemmer
        return SnowballStemmer(_nltk_lang_to_name(lang))
    elif lang == "ar":
        return ISRIStemmer()
    else:
        # if here, there not exits a satisfiable stemmer for the language, so
        # it is returned None, which means that the process of stemming must be skipped
        return None


def _stemming(revert_text, stemmer_reverse_dict, lang: str):
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


def _get_tokenizer_lang(lang: str):
    if _nltk_lang_to_name(lang) not in ['czech', 'danish', 'dutch', 'english', 'estonian',
                                  'finnish', 'french', 'german', 'greek', 'italian',
                                  'norwegian', 'polish', 'portuguese', 'slovene',
                                  'spanish', 'swedish']:
        # language not supported by nltk.tokenizers, so use by default english
        return "english"
    else:
        return _nltk_lang_to_name(lang)


def _get_nlp_processor(lang: str):  # Returns nlp processor and lemmatization capability (True/False)
    if lang == "en" or lang == "simple":
        return (spacy.load("en_core_web_sm", exclude=["parser", "ner", "textcat", "custom"]), True)
    elif lang in ["ca", "da", "de", "el", "es", "fr", "it", "lt", "mk", "nl", "pl", "pt", "ro", "ru"]:
        return (spacy.load(lang + "_core_news_sm", exclude=["parser", "ner", "textcat", "custom"]), True)
    elif lang == "ja":
        return (spacy.load("ja_core_news_sm", exclude=["parser", "ner", "textcat", "custom"]), False)
    elif lang == "zh":
        return (spacy.load("zh_core_web_sm", exclude=["parser", "ner", "textcat", "custom"]), True)
    elif lang == "no":
        return (spacy.load("nb_core_news_sm", exclude=["parser", "ner", "textcat", "custom"]), True)
    elif lang in ["eml", "fur", "lij", "la", "lmo", "nap", "pms", "sc", "scn", "roa-tara", "vec"]:
        return (spacy.blank("it"), False)
    else:  # fallback case (multilingual)
        return (spacy.blank("xx"), False)


def _get_min_word_length(lang: str):  # Returns min admitted word length for the language (sync in topwordspageextractor)
    if lang in ["gan", "ja", "ko", "vi",  "wuu", "zh", "zh-classical", "zh-yue"]:
        return 1  # Hang, Hans, Hant scripts
    elif lang == "vi":
        return 2  # Hybrid case of Chu Nom in Latn
    else:
        return 3


def _delete_dir_content(the_dir: str):
    """
    Fast-delete a directory (no error check: the show must go on)
    :param the_dir: directory to delete
    """
    empty_dir = os.path.join(os.path.dirname(the_dir), "empty_dir")
    os.makedirs(empty_dir, exist_ok=True)
    empty_dir = os.path.join(empty_dir, "")  # add a trailing slash if not present
    the_dir = os.path.join(the_dir, "")  # ditto
    subprocess.run("rsync -a --delete " + empty_dir + " " + the_dir + \
        " ; rmdir " + empty_dir + " " + the_dir, \
        shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)


def log_message(logger, *args):
    if logger != None:
        logger.write(f"{datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')} - ")
        for arg in args:
            logger.write(arg)
        logger.write("\n")
        logger.flush()


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
        shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, \
        start_new_session=True)


def memory_check():
    mem_status = os.popen('free -m')
    mem_status.readline()
    mem_status_ram = mem_status.readline().split()[1:]
    mem_status_vm = mem_status.readline().split()[1:]
    mem_status.close()
    ram_available = int(mem_status_ram[-1]) / int(mem_status_ram[0])
    if mem_status_vm[0] == "0":  # No VM
        vm_available = 1
    else:
        vm_available = int(mem_status_vm[-1]) / int(mem_status_vm[0])
    return (ram_available, vm_available, int(mem_status_vm[1]))


def memory_status(armageddon: str, apocalypse: str, logger):
    status = None
    (ram_available, vm_available, vm_used) = memory_check()
    if (vm_used > 1000) or (ram_available < 0.2) or (vm_available < 0.7):
        log_message(logger, f"Memory warning: {((1 - ram_available) * 100):.1f}% RAM used, "
                            f"{((1 - vm_available) * 100):.1f}% ({vm_used} Mb) VM used.")
        status= armageddon
        log_message(logger, "Armageddon status activated")
        if (ram_available < 0.1) or (vm_available < 0.2):
            log_message(logger, "Upgraded to APOCALYPSE status!")
            status = apocalypse
    return status


def emergency_trigger(emergency: str, executor):
    if emergency != None:
        open(emergency, "a").close()
        children = multiprocessing.active_children()
        for process in children:
            process.terminate()
        executor.join()
        os.remove(emergency)
    return emergency


def check_emergency(logger, emergency_list: list):
    for emergency in emergency_list:
        if os.path.exists(emergency):
            log_message(logger, f"Emergency {os.path.basename(emergency)} found: performing respawn")
            return True
    return False


class KillMeSoftly:
  kill_requested = False
  def __init__(self):
    signal.signal(signal.SIGINT, self.exit_softly)
    signal.signal(signal.SIGTERM, self.exit_softly)

  def exit_softly(self, *args):
    self.kill_requested = True


def _words_extractor(input_dir: str, output_dir: str, o_process: int, parallelism: int, lang: str, \
    armageddon:str, apocalypse:str, log_file: str):
    """
    _words_extractor perform tokenization, stopwords cleaning and lemmatization on a single file "filename"
    :param result_dir: path of input folder
    :param output_dir: path of output folder
    :param o_process: process ordinal (in range(parallelism))
    :param parallelism: degree of parallelism
    :param lang: wikipedia language
    :param armageddon: file for armageddon intervention
    :param apocalypse: file for apocalypse intervention
    :param log_file: file for async non-blocking logging
    """

    killer = KillMeSoftly()
    with open(log_file, "a", encoding='utf-8') as logger:  # Non-blocking async logger
        (nlp, lemmatable) = _get_nlp_processor(lang)

        bsize = 1000 // parallelism
        n_first_bucket = o_process * bsize
        n_last_bucket = 1000 if (o_process == parallelism -1) else (o_process + 1) * bsize

        for file in os.scandir(input_dir):
            if killer.kill_requested:
                if os.path.exists(armageddon) or os.path.exists(apocalypse):
                    log_message(logger, "KILL request: self-killing")
                    break
                else:
                    open(apocalypse, "a").close()
                    log_message(logger, "EXTERNAL KILL (OOM Killer?): self-killing and apocalypse signal")
                    break
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
                        log_message(logger, f"Reverts overflow: {reverts_length} chars ", \
                                            f"(with +{len(single_revert)} chars) from file {file.name}")
                        break
                    reverts_texts.append(single_revert)
                
                multidoc = nlp.pipe(reverts_texts)
                for reverts, doc in zip_longest(dump_dict["Revision"], multidoc, fillvalue=[]):
                    if reverts["Text"] is None:
                        continue
                    if lemmatable:
                        reverts["Text"] = [w.lemma_ for w in doc if (w.pos_ in ALLOWED_POS and w.is_alpha)]
                    else:
                        reverts["Text"] = [w.lower_ for w in doc if w.is_alpha]

                    reverts["Text"] = _words_cleaner(reverts["Text"])

                    if not lemmatable:
                        reverts["Text"], reverse_stemming_dict = _stemming(reverts["Text"], reverse_stemming_dict, lang)

                page_id = dump_dict["PageID"]

                with open(os.path.join(output_dir, "S" + "0" * (20 - len(str(page_id))) + str(page_id) + ".json"), "w", encoding='utf-8') as new_file:
                    json.dump(dump_dict, new_file, ensure_ascii=False)
                    new_file.flush()  # overzealous
                with open(os.path.join(output_dir, "Stem/StemRev_" + str(page_id) + ".json"), "w", encoding='utf-8') as new_file:
                    json.dump(reverse_stemming_dict, new_file, ensure_ascii=False)
                    new_file.flush()  # overzealous
                
                os.remove(file)
        
        logger.flush()  # overzealous


def _stopwords_cleaner_stemming(result_dir: str, filename: str, lang: str):
    """
    _stopwords_cleaner_stemming perform tokenization, stopwords cleaning and stemming on a single file "filename"
    :param result_dir: path of result folder
    :param filename: file to process
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
        error_log.write(f"ERROR at time {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')} - ")
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

    armageddon = "/data/armageddon"  # Armageddon makes the master program gently kill children and respawn
    apocalypse = "/data/apocalypse"  # Apocalypse mode (when external OOM Kill or inadequately low resources)
    rocket = "/data/rocket"  # Rocket mode skips all remaining data extraction and proceeds
    log_prefix = "/data/normalization"
    for log_file in glob.glob(log_prefix + "*"):
        os.remove(log_file)
    input_dir = result_dir + "_input"
    shutil.rmtree(input_dir, ignore_errors=True)
    shutil.move(result_dir, input_dir)
    os.mkdir(result_dir)
    shutil.move(os.path.join(input_dir, "Stem"), result_dir)

    parallelism = max(1, multiprocessing.cpu_count() - 1)
    sleep_time = 8
    monitor_time = 0.5
    with open(log_prefix + ".log", "a", encoding='utf-8') as logger:
        while True:
            executor = multiprocessing.Pool(parallelism)
            for i in range(parallelism):
                log_file = log_prefix + "_" + str(i) + ".log"
                executor.apply_async(_words_extractor, \
                    args=(input_dir, result_dir, i, parallelism, lang, \
                        armageddon, apocalypse, log_file), \
                        error_callback = async_error_logger)
                start_time = time.monotonic()
            executor.close()
            emergency_level = emergency_trigger(memory_status(armageddon, apocalypse, logger), executor)
            while emergency_level == None:
                monitor_time = min(monitor_time * 2, 512)
                time.sleep(monitor_time)
                children = multiprocessing.active_children()
                if os.path.exists(apocalypse):
                    log_message(logger, "External kill (OOM Killer?) detected: activating apocalypse")
                    emergency_level = emergency_trigger(apocalypse, executor)
                elif children == []:
                    break
                elif os.path.exists(rocket):
                    os.remove(rocket)
                    log_message(logger, "Rocket mode: skipping remaining clouds generation")
                    break
                elif os.path.exists(armageddon):
                    log_message(logger, "Armageddon requested: performing respawn")
                    emergency_level = emergency_trigger(armageddon, executor)
                else:
                    emergency_level = emergency_trigger(memory_status(armageddon, apocalypse, logger), executor)
            delta_time = time.monotonic() - start_time
            del executor
            gc.collect()
            if delta_time < 3600 and emergency_level == armageddon:
                emergency_level = apocalypse
            if emergency_level == armageddon:
                time.sleep(8)
                monitor_time = max(8, monitor_time // 2)
                log_message(logger, f"New settings: Monitor={monitor_time}s")
            elif emergency_level == apocalypse:
                if sleep_time > 1024:
                    log_message(logger, "END OF THE WORLD: inadequate resources, skipping processing")
                    break  # End of the world (and apocalypse file not deleted)
                else:
                    if parallelism == 1:
                        sleep_time *= 2
                    monitor_time = max(8, monitor_time // 2)
                    parallelism = max(1, parallelism - 1)
                    log_message(logger, f"New settings: Monitor={monitor_time}s, Sleep={sleep_time}s, Parallelism={parallelism}")
                    time.sleep(sleep_time)
            else:
                break
    
    _async_delete_dir_content(input_dir)
    for log_file in glob.glob(log_prefix + "*"):
        if os.path.getsize(log_file) == 0:
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

    executor = multiprocessing.Pool(multiprocessing.cpu_count())
    for filename in file_to_clean:
        executor.apply_async(_stopwords_cleaner_stemming, args=(result_dir, filename, lang, nlp, lemmatable))
    executor.close()
    executor.join()


def main():
    concurrent_stopwords_cleaner_lemmatizer(sys.argv[1], sys.argv[2])


if __name__ == "__main__":
    main()
