FROM golang:1.9

WORKDIR /go/src/forecast-process-slack
COPY . .

RUN go build -o /usr/local/bin/forecast-process-slack

CMD ["forecast-process-slack"]
