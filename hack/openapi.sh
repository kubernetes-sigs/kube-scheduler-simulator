#!/usr/bin/env bash
SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

go install k8s.io/kube-openapi/cmd/openapi-gen@latest

KUBE_INPUT_DIRS=(
  $(
    grep --color=never -rl '+k8s:openapi-gen=' vendor/k8s.io | \
    xargs -n1 dirname | \
    sed "s,^vendor/,," | \
    sort -u | \
    sed '/^k8s\.io\/kubernetes\/build\/root$/d' | \
    sed '/^k8s\.io\/kubernetes$/d' | \
    sed '/^k8s\.io\/kubernetes\/staging$/d' | \
    sed 's,k8s\.io/kubernetes/staging/src/,,' | \
    grep -v 'k8s.io/code-generator' | \
    grep -v 'k8s.io/sample-apiserver'
  )
)

KUBE_INPUT_DIRS=$(IFS=,; echo "${KUBE_INPUT_DIRS[*]}")

function join { local IFS="$1"; shift; echo "$*"; }

echo "Generating Kubernetes OpenAPI"

$GOPATH/bin/openapi-gen \
  --output-file-base zz_generated.openapi \
  --output-base="${GOPATH}/src" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate/boilerplate.generatego.txt \
  --output-base="./" \
  --input-dirs $(join , "${KUBE_INPUT_DIRS[@]}") \
  --output-package "vendor/k8s.io/kubernetes/pkg/generated/openapi" \
  --report-filename "${SCRIPT_ROOT}/hack/openapi-violation.list" \
  "$@"
