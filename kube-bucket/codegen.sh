#!/bin/bash

ROOT_PACKAGE="github.com/yuankunzhang/devops-challenge/kube-bucket"
CUSTOM_RESOURCE_NAME="bucket"
CUSTOM_RESOURCE_VERSION="v1"

go get -u k8s.io/code-generator/...
cd $GOPATH/src/k8s.io/code-generator
./generate-groups.sh all "$ROOT_PACKAGE/pkg/client" "$ROOT_PACKAGE/pkg/apis" "$CUSTOM_RESOURCE_NAME:$CUSTOM_RESOURCE_VERSION"
