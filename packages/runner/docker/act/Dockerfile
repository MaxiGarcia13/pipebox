FROM alpine:3.21

ARG ACT_VERSION=0.2.97
ARG TARGETARCH

RUN apk add --no-cache ca-certificates curl \
  && arch="${TARGETARCH}" \
  && if [ -z "${arch}" ]; then arch="$(uname -m)"; fi \
  && case "${arch}" in \
       amd64|x86_64) archive_arch="x86_64" ;; \
       arm64|aarch64) archive_arch="arm64" ;; \
       *) echo "unsupported arch: ${arch}" >&2; exit 1 ;; \
     esac \
  && curl -fsSL "https://github.com/nektos/act/releases/download/v${ACT_VERSION}/act_Linux_${archive_arch}.tar.gz" \
       | tar -xz -C /usr/local/bin act \
  && chmod +x /usr/local/bin/act \
  && mkdir -p /root/.config/act

COPY actrc /root/.config/act/actrc

ENTRYPOINT ["act"]
