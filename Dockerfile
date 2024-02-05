FROM debian:bookworm
LABEL maintainer="otiai10 <otiai10@gmail.com>"

ARG LOAD_LANG=jpn

RUN apt update \
    && apt install -y \
      ca-certificates \
      libtesseract5 \
      libtesseract-dev \
      tesseract-ocr \
      golang

ENV GO111MODULE=on
ENV GOPATH=${HOME}/go
ENV PATH=${PATH}:${GOPATH}/bin

ADD . $GOPATH/src/github.com/otiai10/ocrserver
WORKDIR $GOPATH/src/github.com/otiai10/ocrserver
RUN go get -v ./... && go install .

# Load languages
RUN if [ -n "${LOAD_LANG}" ]; then apt-get install -y tesseract-ocr-${LOAD_LANG}; fi

COPY tessdata/ticketplusV5.traineddata /usr/share/tesseract-ocr/5/tessdata/ticketplusV5.traineddata

ENV PORT=8080
CMD ["ocrserver"]
