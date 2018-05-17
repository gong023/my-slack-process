FROM golang:1.10 as builder

ADD . /go/src/github.com/gong023/my-slack-process

RUN /go/src/github.com/gong023/my-slack-process/build.sh

FROM jrottenberg/ffmpeg:3.4-centos

COPY --from=builder /usr/local/bin/stdpost /usr/local/bin/stdpost
COPY --from=builder /usr/local/bin/stdpostb /usr/local/bin/stdpostb
COPY --from=builder /usr/local/bin/stdpostc /usr/local/bin/stdpostc
COPY --from=builder /usr/local/bin/forecast /usr/local/bin/forecast
COPY --from=builder /usr/local/bin/coinbase /usr/local/bin/coinbase
COPY --from=builder /usr/local/bin/hibiki /usr/local/bin/hibiki
COPY --from=builder /usr/local/bin/inoreader /usr/local/bin/inoreader
COPY --from=builder /usr/local/bin/pixiv /usr/local/bin/pixiv
