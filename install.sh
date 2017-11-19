#!/bin/bash

/usr/local/go/bin/go get -u github.com/gong023/my-slack-process...
packages=(stdpost stdpostb forecast inoreader coinbase pixiv)
for package in "${packages[@]}"; do
    /usr/local/go/bin/go install github.com/gong023/my-slack-process/"$package"
done
