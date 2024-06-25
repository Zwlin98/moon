FROM golang:1.22

COPY . /moon

WORKDIR /moon

RUN go build -o /usr/bin/moon

expose 3345

CMD ["/usr/bin/moon"]
