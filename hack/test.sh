#!/bin/bash

GOOS=linux go build ./...
git stash
docker run -ti -v $(pwd):/root/.dotfiles debian /root/.dotfiles/dotfiles
git stash apply
git stash clear