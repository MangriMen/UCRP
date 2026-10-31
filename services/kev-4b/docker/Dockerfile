# Kev-4B decision model — llama.cpp server with the TypeSafe /v1/systemone API.
#
#   docker build -t kev-4b:latest .
#   docker run --rm -p 8080:8080 kev-4b:latest
#
# Self-contained: clones llama.cpp at a pinned commit and pulls the GGUF from
# Hugging Face at build time, so the repo carries no vendored sources and no
# multi-GB weights. Build with BuildKit for the model download cache:
#
#   DOCKER_BUILDKIT=1 docker build -t kev-4b:latest .
#
# Pin / swap via build args (see ARGs below). Runtime config is env vars, see
# docker/entrypoint.sh.
#
# NOTE: this model is a decision model (pointer head, no LM head). vLLM cannot
# serve it; /v1/systemone exists only in llama.cpp. See README.md.

ARG BASE_IMAGE=ubuntu:24.04

# ---------------------------------------------------------------- builder ---
FROM ${BASE_IMAGE} AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential \
        cmake \
        ninja-build \
        git \
        ca-certificates \
        curl \
        libssl-dev \
        zlib1g-dev \
        libzstd-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

# llama.cpp is cloned at a pinned commit (works for non-HEAD refs too).
ARG LLAMA_CPP_REPO=https://github.com/ggml-org/llama.cpp.git
ARG LLAMA_CPP_REF=7bd18083082126c08d375cefaa6a5aa1c170ecde

RUN git init llama.cpp \
    && cd llama.cpp \
    && git remote add origin "${LLAMA_CPP_REPO}" \
    && git fetch --depth 1 origin "${LLAMA_CPP_REF}" \
    && git checkout --detach FETCH_HEAD \
    && echo "llama.cpp @ $(git rev-parse HEAD)"

# GGML_NATIVE=OFF keeps the image portable across x86-64 hosts (e.g. Zen 2
# without AVX-512). Rebuild with -DGGML_NATIVE=ON on the deploy host for
# maximum throughput.
ARG GGML_NATIVE=OFF
ARG CMAKE_BUILD_TYPE=Release

RUN cmake -S llama.cpp -B llama.cpp/build -G Ninja \
        -DCMAKE_BUILD_TYPE=${CMAKE_BUILD_TYPE} \
        -DGGML_NATIVE=${GGML_NATIVE} \
        -DLLAMA_CURL=OFF \
        -DLLAMA_BUILD_TESTS=OFF \
        -DLLAMA_BUILD_EXAMPLES=OFF \
        -DLLAMA_BUILD_SERVER=ON \
    && cmake --build llama.cpp/build --target llama-server -j "$(nproc)"

# Model weights pulled from HF. Mount a local file at runtime instead if you
# prefer not to bake them in (see README.md).
ARG MODEL_REPO=ggml-org/Kev-4B-GGUF
ARG MODEL_FILE=Kev-4B-Q4_K_M.gguf
RUN --mount=type=cache,target=/root/.cache/kev-model \
    set -eu; \
    dst="/root/.cache/kev-model/${MODEL_FILE}"; \
    if [ ! -s "${dst}" ]; then \
        echo "downloading ${MODEL_REPO}/${MODEL_FILE}"; \
        curl -fL --retry 3 --retry-delay 2 \
             -o "${dst}.part" \
             "https://huggingface.co/${MODEL_REPO}/resolve/main/${MODEL_FILE}"; \
        mv "${dst}.part" "${dst}"; \
    fi; \
    mkdir -p /src/models; \
    cp "${dst}" "/src/models/${MODEL_FILE}"; \
    ls -lh /src/models/

# ---------------------------------------------------------------- runtime ---
FROM ${BASE_IMAGE} AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
        libgomp1 \
        libssl3 \
        zlib1g \
        libzstd1 \
        curl \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --create-home --home-dir /app --shell /usr/sbin/nologin kev

WORKDIR /app

# Shared libs + server binary built in the builder stage.
COPY --from=builder --chown=kev:kev /src/llama.cpp/build/bin/ /app/bin/
ENV LD_LIBRARY_PATH=/app/bin

ARG MODEL_FILE=Kev-4B-Q4_K_M.gguf
COPY --from=builder --chown=kev:kev /src/models/ /app/models/
ENV MODEL_PATH=/app/models/${MODEL_FILE}

COPY --chown=kev:kev docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

USER kev

ENV PORT=8080 \
    HOST=0.0.0.0 \
    THREADS=0 \
    CTX_SIZE=8192 \
    PARALLEL_SLOTS=1

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
    CMD curl -sf "http://127.0.0.1:${PORT}/health" || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
