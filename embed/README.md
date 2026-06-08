# weblens-embed

Multimodal embedding sidecar. Loads [Jina CLIP v2](https://huggingface.co/jinaai/jina-clip-v2) (1024-dim, image + text in one space) and exposes:

| Method | Path                  | Body / params                   | Returns                                                 |
| ------ | --------------------- | ------------------------------- | ------------------------------------------------------- |
| GET    | `/encode`             | `?img-path=...`                 | `[1024 floats]` — image embedding                       |
| POST   | `/encode-text`        | `{"text": "..."}`               | `{"text_features": [1024 floats], "image_query_features": [1024 floats]}` — raw + caption-prompted query embeddings |
| POST   | `/extract-and-embed`  | `{"path": "...", "mimeHint": "...?"}` | `[{chunkIndex, page, snippet, vector}, ...]` — per-chunk text embeddings |
| GET    | `/health`             | —                               | `{"status":"ok"}`                                       |

Text extraction (`extract.py`) handles PDF, DOCX, XLSX, PPTX, plaintext / common code files, and OCR-fallback for images via tesseract. Chunking is token-aware using the model's tokenizer (500 tokens, 50 overlap).

## Running locally (dev)

```bash
cd embed
uv sync
uv run main.py
```

Listens on `:5500`. `make dev` from the repo root does this automatically via `scripts/lib/embed.bash`.

`WEBLENS_CACHE_PATH` must be set so the service can resolve `CACHES:`-prefixed paths sent from the Go backend.

## Running in Docker

```bash
docker build -f docker/embed.Dockerfile -t ghcr.io/ethanrous/weblens_embed .
docker run --rm -p 5500:5500 \
  -v $(pwd)/_build/fs/core/cache/:/images \
  -e WEBLENS_CACHE_PATH=/images \
  ghcr.io/ethanrous/weblens_embed
```

The first run downloads ~2 GB of model weights to `/root/.cache/huggingface` — mount a volume there to persist.

## Tests

```bash
cd embed && uv run pytest test_extract.py -v
```

Extraction tests run against fixtures in `test_fixtures/`. To regenerate fixtures: `uv run test_fixtures/_generate.py`. The `test_extract_and_embed_endpoint` test only runs when `EMBED_TEST_LIVE=1` is set (requires the model to be loaded).

## Swapping the model

Change `MODEL_ID` (and `EMBEDDING_DIM` if needed) in `main.py` and `preload.py`. Update `EmbeddingDim` in `models/embedding/init.go` to match, and recreate the `embeddings_vector` Atlas index. The Go-side stores `model` on every row so old rows are distinguishable.
