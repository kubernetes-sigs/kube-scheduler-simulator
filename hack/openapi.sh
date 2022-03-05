#!/usr/bin/env bash

OPENAPIFILE="vendor/k8s.io/kubernetes/pkg/generated/openapi/zz_generated.openapi.go"

# get kubernetes/kubernetes submodule
git submodule update --init

cd submodules/kubernetes
make kube-apiserver
cp pkg/generated/openapi/zz_generated.openapi.go "../../${OPENAPIFILE}"
