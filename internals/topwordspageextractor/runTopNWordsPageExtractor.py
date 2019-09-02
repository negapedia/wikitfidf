import sys

import TopNWordsPageExtractor

TopNWordsPageExtractor.top_n_words_page_extractor(sys.argv[1], sys.argv[2], True)
TopNWordsPageExtractor.top_n_global_words_extractor(sys.argv[1], sys.argv[3], True)
TopNWordsPageExtractor.top_n_topic_words_extractor(sys.argv[1], sys.argv[4], True)
