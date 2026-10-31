#!/usr/bin/env bash
# Entrypoint for the Kev-4B decision-model container.
# Configured entirely via env vars (see Dockerfile) so it deploys without edits.
set -euo pipefail

PORT="${PORT:-8080}"
HOST="${HOST:-0.0.0.0}"
MODEL_PATH="${MODEL_PATH:-/app/models/Kev-4B-Q4_K_M.gguf}"
CTX_SIZE="${CTX_SIZE:-8192}"
PARALLEL_SLOTS="${PARALLEL_SLOTS:-1}"

# THREADS=0 -> let llama.cpp pick nproc; otherwise pin explicitly.
if [[ "${THREADS:-0}" == "0" ]]; then
  THREAD_ARGS=()
else
  THREAD_ARGS=(-t "${THREADS}")
fi

if [[ ! -f "${MODEL_PATH}" ]]; then
  echo "ERROR: model file not found at ${MODEL_PATH}" >&2
  echo "Mount a GGUF with -v /path/to.gguf:${MODEL_PATH} or set MODEL_PATH." >&2
  exit 1
fi

echo "starting llama-server:"
echo "  model   : ${MODEL_PATH}"
echo "  listen  : http://${HOST}:${PORT}"
echo "  ctx     : ${CTX_SIZE}  slots: ${PARALLEL_SLOTS}  threads: ${THREADS:-auto}"

exec /app/bin/llama-server \
  -m "${MODEL_PATH}" \
  --host "${HOST}" \
  --port "${PORT}" \
  -c "${CTX_SIZE}" \
  -np "${PARALLEL_SLOTS}" \
  "${THREAD_ARGS[@]}" \
  "$@"
