#!/usr/bin/env bash

ETCD_VERSION=${ETCD_VERSION:-3.4.13}
ETCD_HOST=${ETCD_HOST:-127.0.0.1}
ETCD_PORT=${ETCD_PORT:-2379}
export KUBE_SCHEDULER_SIMULATOR_ETCD_URL="http://${ETCD_HOST}:${ETCD_PORT}"

kube::etcd::cleanup() {
  kube::etcd::stop
  kube::etcd::clean_etcd_dir
}

kube::etcd::stop() {
  if [[ -n "${ETCD_PID-}" ]]; then
    kill "${ETCD_PID}" &>/dev/null || :
    wait "${ETCD_PID}" &>/dev/null || :
  fi
}

kube::etcd::clean_etcd_dir() {
  if [[ -n "${ETCD_DIR-}" ]]; then
    rm -rf "${ETCD_DIR}"
  fi
}

kube::etcd::start() {
  # Start etcd
  ETCD_DIR=${ETCD_DIR:-$(mktemp -d 2>/dev/null || mktemp -d -t test-etcd.XXXXXX)}
  if [[ -d "${ARTIFACTS:-}" ]]; then
    ETCD_LOGFILE="${ARTIFACTS}/etcd.$(uname -n).$(id -un).log.DEBUG.$(date +%Y%m%d-%H%M%S).$$"
  else
    ETCD_LOGFILE=${ETCD_LOGFILE:-"/dev/null"}
  fi
  echo "etcd --advertise-client-urls ${KUBE_SCHEDULER_SIMULATOR_ETCD_URL} --data-dir ${ETCD_DIR} --listen-client-urls http://${ETCD_HOST}:${ETCD_PORT} --log-level=debug > \"${ETCD_LOGFILE}\" 2>/dev/null"
  etcd --advertise-client-urls "${KUBE_SCHEDULER_SIMULATOR_ETCD_URL}" --data-dir "${ETCD_DIR}" --listen-client-urls "${KUBE_SCHEDULER_SIMULATOR_ETCD_URL}" --log-level=debug 2> "${ETCD_LOGFILE}" >/dev/null &
  ETCD_PID=$!

  echo "Waiting for etcd to come up."
  kube::util::wait_for_url "${KUBE_SCHEDULER_SIMULATOR_ETCD_URL}/health" "etcd: " 0.25 80
  curl -fs -X POST "${KUBE_SCHEDULER_SIMULATOR_ETCD_URL}/v3/kv/put" -d '{"key": "X3Rlc3Q=", "value": ""}'
}

kube::util::wait_for_url() {
  local url=$1
  local prefix=${2:-}
  local wait=${3:-1}
  local times=${4:-30}
  local maxtime=${5:-1}

  command -v curl >/dev/null || {
    echo "curl must be installed"
    exit 1
  }

  local i
  for i in $(seq 1 "${times}"); do
    local out
    if out=$(curl --max-time "${maxtime}" -gkfs "${@:6}" "${url}" 2>/dev/null); then
      echo "On try ${i}, ${prefix}: ${out}"
      return 0
    fi
    sleep "${wait}"
  done
  echo "Timed out waiting for ${prefix} to answer at ${url}; tried ${times} waiting ${wait} between each"
  exit 1
}
