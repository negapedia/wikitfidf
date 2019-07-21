FROM ebonetti/golang-petsc

RUN apt-get update

RUN apt-get install -y software-properties-common
RUN apt install default-jdk -y
RUN apt install python3-pip -y
RUN pip3 install nltk
RUN pip3 install cython

RUN apt-get install -y --no-install-recommends p7zip-full

ADD src/ $GOPATH/src/
RUN 7z x $GOPATH/src/nltk_data.7z -o/root/nltk_data
RUN 7z x $GOPATH/src/badwords_data.7z -o/root/badwords_data

RUN go get github.com/negapedia/wikibrief
RUN go get github.com/ebonetti/ctxutils

RUN cd $GOPATH/src/textnormalizer/ && python3 compile.py build_ext --inplace
RUN cd $GOPATH/src/destemmer/ && python3 compile.py build_ext --inplace

WORKDIR $GOPATH/src

RUN go build DumpProcessor.go
ENTRYPOINT ["./DumpProcessor"]
