import sys

import DeStemmer

DeStemmer.global_page_destem(sys.argv[1])
DeStemmer.global_word_destem(sys.argv[1])
DeStemmer.remove_destem_file(sys.argv[1])
