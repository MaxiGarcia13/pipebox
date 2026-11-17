# Owned local image for PipeBox Jenkins runs (same pattern as pipebox-act).
# Base is the public GHCR jenkinsfile-runner vanilla distribution — not Docker Hub.
ARG JFR_IMAGE=ghcr.io/jenkinsci/jenkinsfile-runner:jre-21
FROM ${JFR_IMAGE}

# Workspace defaults match jenkinsfile-runner vanilla (/workspace Jenkinsfile, /build runWorkspace).
# Entrypoint remains jenkinsfile-runner from the base image.
