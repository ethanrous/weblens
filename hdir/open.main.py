import json
import sys
import torch
import open_clip
from PIL import Image
from flask import Flask, request, jsonify

print("Loading model...", flush=True)

device = "cuda" if torch.cuda.is_available() else "cpu"
model, _, preprocess = open_clip.create_model_and_transforms(
    "hf-hub:laion/CLIP-ViT-H-14-laion2B-s32B-b79K", device=device
)
model.eval()
tokenizer = open_clip.get_tokenizer("ViT-bigG-14")
print("Model loaded on device:", device)

if len(sys.argv) > 1 and sys.argv[1] == "--preload":
    print("Preloaded model, exiting...", flush=True)
    exit(0)

app = Flask(__name__)


@app.route("/encode", methods=["GET"])
def encode():
    img_path = request.args.get("img-path")
    if not img_path:
        return "Image path not provided", 400

    img_path = img_path.replace("CACHES:", "/images/")

    try:
        image = preprocess(Image.open(img_path)).unsqueeze(0).to(device)
    except Exception as e:
        return f"Error processing image: {str(e)}", 404

    with torch.no_grad():
        image_features = model.encode_image(image)
        image_features /= image_features.norm(dim=-1, keepdim=True)

        image_features_lst = image_features.cpu().numpy().tolist()[0]
        features_str = json.dumps(image_features_lst)

        return features_str


@app.errorhandler(500)
def page_not_found(e):
    return jsonify(error=500, text=str(e)), 500


@app.route("/encode-text", methods=["POST"])
def encodeText():
    request_data = request.get_json()
    search_text: str = request_data["text"]
    text = tokenizer([search_text]).to(device)

    with torch.no_grad():
        text_features = model.encode_text(text).to(device, dtype=torch.float32)
        text_features /= text_features.norm(dim=-1, keepdim=True)

    return jsonify({"text_features": text_features.cpu().numpy().tolist()[0]})


if __name__ == "__main__":
    app.run(debug=False, host="0.0.0.0", port=5000)
