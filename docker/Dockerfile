FROM golang:1.4.2

RUN apt-get update
RUN apt-get install -y supervisor &&  rm -r /var/lib/apt/lists/*

WORKDIR /code

EXPOSE 9001

CMD ["bin/start-local.sh"]
