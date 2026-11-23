#!/bin/sh
# jenkins.sh keeps the start user. Start as root, then continue as the host
# user with the socket's group so durable-task files stay deletable and docker.sock is usable.
set -eu

home="${JENKINS_HOME:-/var/jenkins_home}"
sock=/var/run/docker.sock
uid="${PIPEBOX_UID:-1000}"
gid="${PIPEBOX_GID:-1000}"

if [ "$(id -u)" = "0" ]; then
  if [ -d "$home" ]; then
    chown -R "$uid:$gid" "$home"
  fi
  export HOME="$home"
  groups=""
  if [ -S "$sock" ]; then
    groups=$(stat -c '%g' "$sock")
  fi
  if [ -n "$groups" ]; then
    exec setpriv --reuid="$uid" --regid="$gid" --groups "$groups" --inh-caps=-all /usr/local/bin/jenkins.sh "$@"
  fi
  exec setpriv --reuid="$uid" --regid="$gid" --inh-caps=-all /usr/local/bin/jenkins.sh "$@"
fi

exec /usr/local/bin/jenkins.sh "$@"
