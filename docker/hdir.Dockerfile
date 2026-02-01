# Use the official PyTorch image as a base
FROM ghcr.io/astral-sh/uv:bookworm-slim

# Set the working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates && rm -rf /var/lib/apt/lists/*;

# Copy the rest of the application code
COPY hdir/pyproject.toml .

RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --python 3.13;

COPY hdir/open.main.py main.py

RUN uv run --python 3.13 main.py --preload

# Expose the server port (update if needed)
EXPOSE 5000

CMD ["uv", "run", "--python", "3.13", "main.py"]

