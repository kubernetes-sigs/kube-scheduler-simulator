#!/usr/bin/env bash

OPENAPIFILE="k8sapiserver/openapi/zz_generated.openapi.go"

# get kubernetes/kubernetes submodule
git submodule update --init

cd submodules/kubernetes
make kube-apiserver
cp pkg/generated/openapi/zz_generated.openapi.go "../../${OPENAPIFILE}"
