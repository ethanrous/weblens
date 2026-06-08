FROM ghcr.io/astral-sh/uv:bookworm-slim

# Set the working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git ca-certificates tesseract-ocr tesseract-ocr-eng \
    && rm -rf /var/lib/apt/lists/*;

COPY embed/pyproject.toml embed/uv.lock ./

RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --python 3.13 --no-install-project;

COPY embed/main.py embed/extract.py ./

# DON'T PRELOAD THE MODEL. This makes the image massive, and this will happen automatically on the first run anyway.
# The user should mount a volume at `/root/.cache/huggingface` to persist the model across runs instead.
# RUN uv run --python 3.13 main.py --preload

# Expose the server port (update if needed)
EXPOSE 5500

CMD ["uv", "run", "--python", "3.13", "main.py"]

