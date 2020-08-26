FROM golang:1.14

ENV GOPATH /go
#ENV GOOS linux
#ENV GOARCH  amd64

WORKDIR /go
COPY ./reversi ./src/reversi
COPY ./reversiSimulation ./src/reversiSimulation

RUN go build reversi
RUN go build reversiSimulation

