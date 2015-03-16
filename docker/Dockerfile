FROM golang:1.4.2

RUN apt-get update
RUN apt-get install -y curl supervisor build-essential &&  rm -r /var/lib/apt/lists/*

ADD . /go/src/github.com/primait/go-bin

RUN cd /go/src/github.com/primait/go-bin && make

WORKDIR /go/src/github.com/primait/go-bin

EXPOSE 9001

CMD ["bin/start-local.sh"]
