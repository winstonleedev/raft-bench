FROM golang as build-env

RUN mkdir /gostuff
WORKDIR /gostuff
COPY go.mod go.sum ./
RUN go mod download

WORKDIR /go/src/app
COPY . .
RUN go build .

CMD ["../raftbench"]
