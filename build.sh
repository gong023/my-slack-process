#!/usr/bin/env bash
BUILD_DIR=$(cd $(dirname $0); pwd)
packages=(stdpost stdpostb stdpostc forecast coinbase hibiki inoreader pixivf pixivr twitterf)
for package in "${packages[@]}"; do
    cd "$BUILD_DIR"/"$package"
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /usr/local/bin/"$package" github.com/gong023/my-slack-process/"$package"
done
