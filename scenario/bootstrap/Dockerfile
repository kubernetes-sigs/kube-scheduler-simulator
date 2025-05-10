# Dockerfile
FROM bitnami/kubectl:latest

USER root
RUN install_packages openssl curl ca-certificates \
    && curl -Lo /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 \
    && chmod +x /usr/local/bin/yq

COPY bootstrap.sh /usr/local/bin/bootstrap.sh
RUN chmod +x /usr/local/bin/bootstrap.sh

ENTRYPOINT ["bootstrap.sh"]