#!/usr/bin/env bash

packages=(stdpost stdpostb stdpostc forecast coinbase hibiki inoreader pixivf)
for package in "${packages[@]}"; do
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /usr/local/bin/"$package" github.com/gong023/my-slack-process/"$package"
done
