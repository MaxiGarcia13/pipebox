# Ephemeral PipeBox Jenkins controller (official LTS, not jenkinsfile-runner).
# Plugins, JCasC, built-in toolchain, and a Docker CLI for DooD agents.
ARG JENKINS_IMAGE=jenkins/jenkins:lts-jdk21
FROM ${JENKINS_IMAGE}

ARG TARGETARCH
ARG GO_VERSION=1.22.12
ARG NODE_VERSION=26.11.0
ARG PNPM_VERSION=12.18.0
ARG DOCKER_VERSION=27.5.1
ARG PIPEBOX_JENKINS_REVISION

LABEL pipebox.jenkins.revision="${PIPEBOX_JENKINS_REVISION}"

USER root

# xvfb and the GTK stack are for Jenkinsfiles that launch a desktop app
# on the built-in node (this repo's e2e stage uses xvfb-run).
RUN apt-get update \
  && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    openssh-client \
    build-essential \
    xz-utils \
    xvfb \
    xauth \
    libnss3 \
    libxss1 \
    libxtst6 \
    libgbm1 \
    libdrm2 \
    libnotify4 \
    libxcomposite1 \
    libxdamage1 \
    libxrandr2 \
    libxkbcommon0 \
    libpango-1.0-0 \
    libcairo2 \
    libcups2 \
    libxshmfence1 \
    fonts-liberation \
  && (DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
       libgtk-3-0t64 libatk-bridge2.0-0t64 libasound2t64 \
     || DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
       libgtk-3-0 libatk-bridge2.0-0 libasound2) \
  && rm -rf /var/lib/apt/lists/*

RUN set -eux; \
  arch="${TARGETARCH:-}"; \
  if [ -z "${arch}" ]; then arch="$(uname -m)"; fi; \
  case "${arch}" in \
    amd64|x86_64) goarch=amd64; nodearch=x64; dockerarch=x86_64 ;; \
    arm64|aarch64) goarch=arm64; nodearch=arm64; dockerarch=aarch64 ;; \
    *) echo "unsupported arch: ${arch}" >&2; exit 1 ;; \
  esac; \
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${goarch}.tar.gz" \
    | tar -C /usr/local -xz; \
  curl -fsSL "https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-${nodearch}.tar.xz" \
    | tar -xJ -C /usr/local --strip-components=1; \
  curl -fsSL "https://download.docker.com/linux/static/stable/${dockerarch}/docker-${DOCKER_VERSION}.tgz" \
    | tar -xz -C /tmp; \
  mv /tmp/docker/docker /usr/local/bin/docker; \
  rm -rf /tmp/docker; \
  export PATH="/usr/local/bin:/usr/local/go/bin:${PATH}"; \
  npm install -g --allow-scripts=pnpm "pnpm@${PNPM_VERSION}"; \
  chmod 755 /usr/local/bin/docker

ENV PATH="/usr/local/bin:/usr/local/go/bin:/var/jenkins_home/go/bin:${PATH}"
ENV GOPATH=/var/jenkins_home/go
ENV DOCKER_HOST=unix:///var/run/docker.sock
ENV CASC_JENKINS_CONFIG=/var/jenkins_home/casc_configs
ENV JAVA_OPTS="-Djenkins.install.runSetupWizard=false -Dhudson.security.csrf.GlobalCrumbIssuerConfiguration.DISABLE_CSRF_PROTECTION=true"

COPY --chown=jenkins:jenkins plugins.txt /usr/share/jenkins/ref/plugins.txt
RUN jenkins-plugin-cli --plugin-file /usr/share/jenkins/ref/plugins.txt

COPY --chown=jenkins:jenkins casc.yaml /usr/share/jenkins/ref/casc_configs/jenkins.yaml

COPY entrypoint.sh /usr/local/bin/pipebox-entrypoint.sh
RUN chmod 755 /usr/local/bin/pipebox-entrypoint.sh

ENTRYPOINT ["/usr/bin/tini", "--", "/usr/local/bin/pipebox-entrypoint.sh"]
USER jenkins
