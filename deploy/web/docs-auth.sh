#!/bin/sh
set -eu

umask 077
printf '%s\n' "${DOCS_HTPASSWD:-}" > /etc/nginx/docs.htpasswd
chmod 644 /etc/nginx/docs.htpasswd
