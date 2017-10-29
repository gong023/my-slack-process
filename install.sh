#!/bin/bash

/usr/local/go/bin/go get github.com/gong023/my-slack-process...
packages=(stdpost forecast)
for package in "${packages[@]}"; do
    /usr/local/go/bin/go install github.com/gong023/my-slack-process/"$package"
done
