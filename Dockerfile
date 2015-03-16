FROM golang:1.4.2

RUN apt-get update
RUN apt-get install -y supervisor &&  rm -r /var/lib/apt/lists/*

ADD . /go/src/github.com/primait/go-bin

WORKDIR /go/src/github.com/primait/go-bin

EXPOSE 9001

CMD ["bin/start.sh"]
