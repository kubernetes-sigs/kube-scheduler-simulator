#!/usr/bin/env bash
set -euo pipefail
trap 'echo "[Error] on line $LINENO"; exit 1' ERR

### 定数定義 ###
readonly KUBECONFIG_PATH=/config/kubeconfig.yaml
readonly TMP_MANIFEST=/tmp/webhook
readonly WORK_MANIFEST=/manifests/webhook
readonly CRD_DIR=/manifests/crd
readonly CAFILE_DIR=$WORK_MANIFEST/ca
readonly CERT_DIR=$WORK_MANIFEST/certs
readonly NAMESPACE=scenario-system

### マニフェスト同期（certs は残す） ###
sync_manifests() {
  mkdir -p "$WORK_MANIFEST"
  find "$WORK_MANIFEST" -mindepth 1 -maxdepth 1 \
    ! -name certs -exec rm -rf {} +
  cp -a "$TMP_MANIFEST"/. "$WORK_MANIFEST"/
}

### API サーバ起動待ち ###
wait_for_k8s() {
  echo ">>> Waiting for Kubernetes API server…"
  until kubectl get nodes &>/dev/null; do sleep 1; done
  echo ">>> API server is up!"
}

### CA 証明書生成 ###
generate_ca() {
  mkdir -p "$CAFILE_DIR"
  cat >"$CAFILE_DIR/openssl-ca.cnf" <<'EOF'
[req]
distinguished_name = req_distinguished_name
x509_extensions    = v3_ca

[req_distinguished_name]

[v3_ca]
basicConstraints = critical,CA:TRUE,pathlen:0
keyUsage         = critical,keyCertSign,cRLSign
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always
EOF

  openssl req -x509 -nodes -newkey rsa:2048 -days 3650 \
    -subj "/CN=${NAMESPACE}-webhook-ca" \
    -keyout "$CAFILE_DIR/ca.key" \
    -out    "$CAFILE_DIR/ca.crt" \
    -config "$CAFILE_DIR/openssl-ca.cnf" \
    -extensions v3_ca
}

### サーバ証明書生成 ###
generate_server_cert() {
  mkdir -p "$CERT_DIR"
  cat >"$CERT_DIR/openssl-server.cnf" <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions     = v3_req

[req_distinguished_name]

[v3_req]
keyUsage        = digitalSignature,keyEncipherment
extendedKeyUsage= serverAuth
subjectAltName  = @alt_names

[alt_names]
DNS.1 = scenario-webhook-service.${NAMESPACE}.svc
DNS.2 = scenario-webhook-service.${NAMESPACE}.svc.cluster.local
EOF

  openssl req -nodes -newkey rsa:2048 \
    -subj "/CN=scenario-webhook-service.${NAMESPACE}.svc" \
    -keyout "$CERT_DIR/tls.key" \
    -out    "$CERT_DIR/tls.csr" \
    -config "$CERT_DIR/openssl-server.cnf"

  openssl x509 -req \
    -in "$CERT_DIR/tls.csr" \
    -CA "$CAFILE_DIR/ca.crt" \
    -CAkey "$CAFILE_DIR/ca.key" \
    -CAcreateserial \
    -out "$CERT_DIR/tls.crt" \
    -days 3650 \
    -extensions v3_req \
    -extfile "$CERT_DIR/openssl-server.cnf"

  chmod 644 "$CERT_DIR/tls.crt" "$CERT_DIR/tls.key"
  chmod 755 "$CERT_DIR"
}

### Secret 作成 & Webhook 適用 ###
apply_webhook() {
  kubectl delete secret scenario-webhook-tls -n "$NAMESPACE" --ignore-not-found
  kubectl create secret tls scenario-webhook-tls \
    --cert="$CERT_DIR/tls.crt" \
    --key="$CERT_DIR/tls.key" \
    -n "$NAMESPACE"

  export CA_BUNDLE=$(base64 -w0 "$CAFILE_DIR/ca.crt")
  yq eval -i '.webhooks[].clientConfig.caBundle = strenv(CA_BUNDLE)' \
    "$WORK_MANIFEST/manifests.yaml"

  kubectl apply -k "$WORK_MANIFEST"
}

### メイン処理 ###
main() {
  sync_manifests
  export KUBECONFIG="$KUBECONFIG_PATH"

  wait_for_k8s

  kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml \
    | kubectl apply -f -

  generate_ca
  generate_server_cert
  apply_webhook

  kubectl apply -k "$CRD_DIR"

  echo ">>> Bootstrap complete"
  tail -f /dev/null
}

main "$@"
