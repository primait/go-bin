FROM golang:1.4.2

RUN apt-get update
RUN apt-get install -y python-pip
RUN pip install supervisor --pre

ADD . /go/src/github.com/primait/go-bin

RUN cd /go/src/github.com/primait/go-bin && make

EXPOSE 9001 9001

CMD ["supervisord", "-n", "-c", "src/github.com/primait/go-bin/supervisord.conf"]
