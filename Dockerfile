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
        cython; \
	apt-get clean; \
	rm -rf /var/lib/apt/lists/*;

ENV PROJECT github.com/negapedia/wikitfidf
ADD . $GOPATH/src/$PROJECT
RUN go get $PROJECT/...;
RUN mkdir -p /data /data/internal /data/internal/textnormalizer /data/internal/destemmer /data/internal/topwordspageextractor
RUN ls
RUN cp $GOPATH/src/$PROJECT/internal/textnormalizer/* /data/internal/textnormalizer
RUN cp $GOPATH/src/$PROJECT/internal/destemmer/* /data/internal/destemmer
RUN cp $GOPATH/src/$PROJECT/internal/topwordspageextractor/* /data/internal/topwordspageextractor
WORKDIR /data