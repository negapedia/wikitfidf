FROM ebonetti/golang-petsc

RUN set -eux; \
	apt-get update && apt-get install -y --no-install-recommends \
		p7zip-full \
        default-jdk \
        python3-dev \
		python3-pip \
        python3-setuptools; \
    pip3 install --no-cache-dir \
        nltk \
        spacy[lookups,ja,ko,th] \
        cython; \
    python3 -m spacy download ca_core_news_sm; \
    python3 -m spacy download da_core_news_sm; \
    python3 -m spacy download de_core_news_sm; \
    python3 -m spacy download el_core_news_sm; \
    python3 -m spacy download en_core_web_sm; \
    python3 -m spacy download es_core_news_sm; \
    python3 -m spacy download fr_core_news_sm; \
    python3 -m spacy download it_core_news_sm; \
    python3 -m spacy download ja_core_news_sm; \
    python3 -m spacy download lt_core_news_sm; \
    python3 -m spacy download nl_core_news_sm; \
    python3 -m spacy download pl_core_news_sm; \
    python3 -m spacy download pt_core_news_sm; \
    python3 -m spacy download ro_core_news_sm; \
    python3 -m spacy download ru_core_news_sm; \
    python3 -m spacy download zh_core_web_sm; \
    python3 -m spacy download xx_ent_wiki_sm; \
	apt-get clean; \
	rm -rf /var/lib/apt/lists/*;

ENV PROJECT github.com/negapedia/wikitfidf
ADD . $GOPATH/src/$PROJECT
RUN go get $PROJECT/...;
WORKDIR /data