FROM golang:1.9.0-stretch

RUN apt-get update -y
RUN apt-get install unzip
RUN go get "github.com/valyala/fasthttp"
RUN go get "github.com/pquerna/ffjson"
RUN go get "github.com/buger/jsonparser"
ENV GOMAXPROCS=4
ADD src ./src
ADD heater_src ./heater_src
RUN go build -o app src/*.go
RUN go build -o heater heater_src/*.go
EXPOSE 80
CMD ./app
