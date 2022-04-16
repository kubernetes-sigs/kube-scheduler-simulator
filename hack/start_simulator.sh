#!/usr/bin/env bash

source "./hack/etcd.sh"

check_if_etcd_exists() {
  echo "Checking etcd is on \$PATH"
  which etcd && return
  echo "Cannot find etcd on \$PATH."
  echo "Please see https://git.k8s.io/community/contributors/devel/sig-testing/integration-tests.md#install-etcd-dependency for instructions."
  echo "You can use 'hack/install-etcd.sh'on kubernetes/kubernetes repo to install a copy."
  exit 1
}

CLEANUP_REQUIRED=
start_etcd() {
  echo "Starting etcd instance"
  CLEANUP_REQUIRED=1
  kube::etcd::start
  echo "etcd started"
}

cleanup_etcd() {
  if [[ -z "${CLEANUP_REQUIRED}" ]]; then
    return
  fi
  echo "Cleaning up etcd"
  kube::etcd::cleanup
  CLEANUP_REQUIRED=
  echo "Clean up finished"
}

check_if_etcd_exists

start_etcd
trap cleanup_etcd EXIT

PORT=1212 FRONTEND_URL=http://localhost:3000 ./bin/simulator

