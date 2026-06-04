import json
import sys
import torch
import os
from transformers import AutoModel
from PIL import Image
from flask import Flask, request, jsonify

from extract import extract_text, ExtractionError

MODEL_ID = "jinaai/jina-clip-v2"
EMBEDDING_DIM = 1024
CHUNK_TOKENS = 500
CHUNK_OVERLAP = 50
SNIPPET_CHARS = 400

print("Loading model...", flush=True)
cache_path = (
    os.environ["WEBLENS_CACHE_PATH"]
    if "WEBLENS_CACHE_PATH" in os.environ
    else "/images/"
)

if not cache_path.endswith("/"):
    cache_path += "/"

if not os.path.exists(cache_path):
    print(f"Cache path '{cache_path}' does not exist. Exiting...", flush=True)
    exit(1)


print(f"Using cache path: {cache_path}")

device = "cpu"
if torch.backends.mps.is_available():
    device = torch.device("mps")
    print("MPS device found. Using MPS (Apple Silicon).", flush=True)
elif torch.cuda.is_available():
    device = "cuda"
    print("CUDA device found. Using GPU.", flush=True)

print(f"Loading model: {MODEL_ID}", flush=True)
model = AutoModel.from_pretrained(MODEL_ID, trust_remote_code=True).to(device)
model.eval()
print("Model loaded on device:", device)

if len(sys.argv) > 1 and "--preload" in sys.argv:
    print("Preloaded model, exiting...", flush=True)
    exit(0)

app = Flask(__name__)


@app.route("/encode", methods=["GET"])
def encode():
    img_path = request.args.get("img-path")
    if not img_path:
        return "Image path not provided", 400

    img_path = img_path.replace("CACHES:", cache_path)

    try:
        image = Image.open(img_path).convert("RGB")
    except Exception as e:
        return f"Error processing image: {str(e)}", 404

    with torch.no_grad():
        vec = model.encode_image([image], truncate_dim=EMBEDDING_DIM)[0]
        norm = (vec ** 2).sum() ** 0.5
        if norm > 0:
            vec = vec / norm

    return json.dumps(vec.tolist())


@app.errorhandler(500)
def page_not_found(e):
    return jsonify(error=500, text=str(e)), 500


@app.route("/encode-text", methods=["POST"])
def encodeText():
    request_data = request.get_json()
    search_text: str = request_data["text"]

    # Two encodings of the same query:
    #  - text_features: the raw query, symmetric with how document chunks are
    #    embedded in /extract-and-embed. Used for file_chunk matching.
    #  - image_query_features: wrapped in a caption template. CLIP-family text
    #    encoders discriminate images better when the query reads like a
    #    caption ("a photo of ..."); this vector is used only for image
    #    matching, where that bias is desired.
    prompted = f"a photo of {search_text}"

    with torch.no_grad():
        plain_vec = model.encode_text([search_text], truncate_dim=EMBEDDING_DIM)[0]
        prompted_vec = model.encode_text([prompted], truncate_dim=EMBEDDING_DIM)[0]

        plain_norm = (plain_vec ** 2).sum() ** 0.5
        if plain_norm > 0:
            plain_vec = plain_vec / plain_norm

        prompted_norm = (prompted_vec ** 2).sum() ** 0.5
        if prompted_norm > 0:
            prompted_vec = prompted_vec / prompted_norm

    return jsonify({
        "text_features": plain_vec.tolist(),
        "image_query_features": prompted_vec.tolist(),
    })


def _chunk_text(text: str) -> list[str]:
    """Token-aware chunking using the model's tokenizer (with word-based fallback)."""
    tokenizer = getattr(model, "tokenizer", None)
    if tokenizer is None:
        words = text.split()
        step = CHUNK_TOKENS - CHUNK_OVERLAP
        if not words:
            return [""]
        return [" ".join(words[i:i + CHUNK_TOKENS]) for i in range(0, len(words), step)]

    ids = tokenizer.encode(text, add_special_tokens=False)
    if not ids:
        return [""]
    chunks: list[str] = []
    step = CHUNK_TOKENS - CHUNK_OVERLAP
    for start in range(0, len(ids), step):
        piece_ids = ids[start:start + CHUNK_TOKENS]
        chunks.append(tokenizer.decode(piece_ids, skip_special_tokens=True))
        if start + CHUNK_TOKENS >= len(ids):
            break
    return chunks


@app.route("/extract-and-embed", methods=["POST"])
def extract_and_embed():
    data = request.get_json(force=True) or {}
    path = data.get("path")
    mime = data.get("mimeHint")
    if not path:
        return jsonify({"error": "path required"}), 400

    path = path.replace("CACHES:", cache_path)
    try:
        pages = extract_text(path, mime)
    except ExtractionError as e:
        return jsonify({"error": str(e), "code": "extraction_failed"}), 422
    except FileNotFoundError:
        return jsonify({"error": "file not found"}), 404

    out: list[dict] = []
    chunk_index = 0
    with torch.no_grad():
        for page_num, page_text in pages:
            page_text = page_text.strip()
            if not page_text:
                continue
            for chunk in _chunk_text(page_text):
                vec = model.encode_text([chunk], truncate_dim=EMBEDDING_DIM)[0]
                norm = (vec ** 2).sum() ** 0.5
                if norm > 0:
                    vec = vec / norm
                out.append({
                    "chunkIndex": chunk_index,
                    "page": page_num,
                    "snippet": chunk[:SNIPPET_CHARS],
                    "vector": vec.tolist(),
                })
                chunk_index += 1
    return jsonify(out)


@app.route("/health")
def health():
    return jsonify({"status": "ok"})


if __name__ == "__main__":
    port = 5500
    if "--port" in sys.argv:
        port_index = sys.argv.index("--port") + 1
        port = int(sys.argv[port_index])
        print(f"Using custom port: {port}", flush=True)

    app.run(debug=False, host="0.0.0.0", port=port)
