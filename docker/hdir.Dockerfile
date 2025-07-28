# Use the official PyTorch image as a base
FROM python:3.12-slim

# Set the working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
	pipx

RUN pipx install poetry
ENV PATH="/root/.local/bin:${PATH}"

# Copy the rest of the application code
COPY hdir/pyproject.toml .
COPY hdir/poetry.lock .

RUN poetry install --no-root 

COPY hdir/preload.py .
RUN poetry run python preload.py 

COPY hdir/main.py .

# Expose the server port (update if needed)
EXPOSE 5000

CMD ["poetry", "run", "python", "main.py"]

