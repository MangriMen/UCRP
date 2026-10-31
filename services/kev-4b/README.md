# Kev-4B — классификатор над математикой

Стенд поднимает модель [`ggml-org/Kev-4B-GGUF`](https://huggingface.co/ggml-org/Kev-4B-GGUF)
и проверяет её как классификатор математических текстов.

```bash
docker compose up -d --build            # собрать и поднять (порт 8080)
docker compose --profile test run --rm test   # smoke-тесты
python3 test_math_classifier.py --url http://127.0.0.1:8080
```

**Результат: 8/8 тестов проходят.**

## Почему не vLLM

Запрос был «поднять vLLM». Попытка сделана, вот что выяснилось.

| Шаг | Что произошло                                                                                                                                                                                           |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Установлен `vllm-cpu==0.32.0` + `vllm-gguf-plugin` (GPU нет, только AMD iGPU без ROCm)                                                                                                                  |
| 2   | `RuntimeError: Unknown gguf model_type: qwen3_5` — плагин ищет архитектуру по HF `model_type`, а gguf-py называет её `qwen35`. Локально починено патчем в `vllm_gguf_plugin/weights_adapter/default.py` |
| 3   | `AttributeError: 'Qwen3_5VisionConfig' object has no attribute 'num_hidden_layers'` — мультимодальная обёртка `Qwen3_5ForConditionalGeneration`                                                         |
| 4   | **Корневая причина:** в GGUF **нет LM-head**. Нет тензора `output.weight` (проекции на словарь в 248 320 токенов). Вместо него `cls.output.weight` размера `[2560, 512]` + `qwen35.decision.type = kev` |

Kev-4B — это **decision model** (TypeSafe «System One»): он читает состояние и
набор типизированных вопросов, прогоняет их за один prefill и возвращает
калиброванное распределение вероятностей по вариантам ответа. Он не умеет
генерировать текст — карточка модели прямо называет генерацию out-of-scope.
Единственный реализованный интерфейс — `POST /v1/systemone`, и он есть в
**llama.cpp** (`tools/server/server-decision.cpp`), а в vLLM его нет в принципе:
ни endpoint, ни pointer-head, ни чтения `qwen35.decision.*`.

Поэтому vLLM может загрузить бэкбон, но не может «подавать» эту модель как
классификатор. Рабочий путь — `llama-server` из llama.cpp (собран здесь из
`7bd1808`, где этот model type поддержан).

## Что делает тест

`test_math_classifier.py` шлёт на `/v1/systemone` восемь случаев:

- **6 × `choice`** — «к какой ветви математики это относится?» с критериями
  `algebra / geometry / calculus / statistics / number_theory / not_math`;
- **2 × `noul`** — «это вообще про математику?» (да/нет).

Каждый кейс содержит ожидаемую метку и допустимый набор (чтобы смежные ветви
не роняли тест). Тест валидирует, что сервер отвечает и что метрополитен
совпадает с ожиданием. Код — только стандартная библиотека Python.

Пример ответа (сокращён):

```
PASS  derivative -> calculus: choice=calculus conf=0.965 probs={'calculus': 0.971, ...}
PASS  shopping list -> not math: choice=not_math conf=0.974 probs={'not_math': 0.978, ...}
PASS  is this math at all? (noul) -> yes: p(yes)=0.944
result: 8 passed, 0 failed, 8 total
```

## Окружение

- CPU: AMD Ryzen 7 9700X (8C/16T), 29 ГБ RAM. GPU нет (AMD iGPU без ROCm).
- Модель: `Kev-4B-Q4_K_M.gguf`, 3.0 ГБ, архитектура `qwen35` (Qwen3.5-4B-Base
  - LoRA rank 16 + pointer head), 32 слоя (24 Gated DeltaNet + 8 full attention).
- llama.cpp клонируется и собирается внутри образа на пиновом коммите `7bd1808`.
- Переменные: `PORT` (по умолчанию 8080), `MODEL_PATH` (путь к GGUF).

## Сборка и деплой

Сборка самодостаточная: Dockerfile сам клонирует llama.cpp и скачивает веса с
Hugging Face. В репозитории нет ни исходников llama.cpp, ни многогигабайтных
весов — всё это появляется в момент сборки.

```bash
DOCKER_BUILDKIT=1 docker build -f docker/Dockerfile -t kev-4b:latest .
docker run -d --name kev4b -p 8080:8080 kev-4b:latest
python3 test_math_classifier.py --url http://127.0.0.1:8080
```

Или через compose:

```bash
docker compose up -d --build                 # сборка + запуск
docker compose --profile test run --rm test  # smoke-тесты
```

Для деплоя уже собранного образа (без локальной сборки):

```bash
IMAGE=registry.example.com/kev-4b:latest \
  docker compose -f docker-compose.deploy.yml up -d
```

На выходе — `llama-server` с `/v1/systemone` и healthcheck на `/health`.
Образ с моделью Q4_K_M весит ~6.1 ГБ.

Build-арги Dockerfile:

| Арг              | По умолчанию                                | Смысл                                     |
| ---------------- | ------------------------------------------- | ----------------------------------------- |
| `LLAMA_CPP_REPO` | `https://github.com/ggml-org/llama.cpp.git` | репозиторий llama.cpp                     |
| `LLAMA_CPP_REF`  | `7bd1808…` (проверенный)                    | тег или коммит                            |
| `GGML_NATIVE`    | `OFF`                                       | `ON` — оптимизация под CPU сборки         |
| `MODEL_REPO`     | `ggml-org/Kev-4B-GGUF`                      | HF-репозиторий                            |
| `MODEL_FILE`     | `Kev-4B-Q4_K_M.gguf`                        | файл весов (`Q8_0` / `BF16` для точности) |

Переменные окружения контейнера:

| Переменная       | По умолчанию                     | Смысл                                        |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `PORT`           | `8080`                           | порт сервера                                 |
| `HOST`           | `0.0.0.0`                        | адрес прослушивания                          |
| `MODEL_PATH`     | `/app/models/Kev-4B-Q4_K_M.gguf` | путь к GGUF внутри контейнера                |
| `THREADS`        | `0` (auto = nproc)               | потоки CPU                                   |
| `CTX_SIZE`       | `8192`                           | контекст                                     |
| `PARALLEL_SLOTS` | `1`                              | слоты (decision model, увеличивать не нужно) |

Чтобы подменить модель на более точную (Q8_0 / BF16) — смонтируйте файл:

```bash
docker run -d --name kev4b -p 8080:8080 \
  -v /path/to/Kev-4B-Q8_0.gguf:/app/models/model.gguf \
  -e MODEL_PATH=/app/models/model.gguf \
  kev-4b:latest
```

**Про портативность.** Сборка ведётся с `GGML_NATIVE=OFF` — образ работает на
любом x86-64, включая Zen 2 без AVX-512 (Ryzen 5 3600). Если пересобираете уже
на целевом хосте, ставьте `--build-arg GGML_NATIVE=ON` ради оптимизаций под его
CPU — но тогда образ заведётся только там.

**Ограничение.** `n_cpu_moe`/GPU-варианты не задействованы: модель крутится на
CPU, ~2.2 ГБ RAM, 8 тестов прогоняются примерно за 3 с. Для прод-нагрузки
стоит смотреть в сторону ускорения — но в этой машине дискретной GPU нет, и
Kev-4B на GPU требует отдельного ядра (`kev.serve`), а не llama-server.

## Файлы

- `docker/Dockerfile` — multi-stage сборка (llama.cpp + модель, всё тянется в момент сборки)
- `docker/entrypoint.sh` — старт сервера, настройка через env
- `docker-compose.yml` — сборка + запуск + one-shot тест
- `docker-compose.deploy.yml` — деплой готового образа из registry
- `test_math_classifier.py` — smoke-тест классификатора (stdlib)
- `.dockerignore`, `.gitignore` — отсекают веса, артефакты и логи
