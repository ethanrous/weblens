import sys
import json
import numpy
import torch
import clip
from PIL import Image
from flask import Flask, request, jsonify

device = "cuda" if torch.cuda.is_available() else "cpu"
model, preprocess = clip.load("ViT-B/32", device=device)
print("Model loaded on device:", device)

# If the script is run with the argument "--preload", it will exit immediately after loading the model.
if len(sys.argv) > 1 and sys.argv[1] == "--preload":
    exit(0)

app = Flask(__name__)

@app.route("/encode", methods=["GET"])
def encode():
    img_path = request.args.get('img-path')
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

@app.route("/match", methods=["POST"])
def match():
    request_data = request.get_json()
    search_text: str = request_data['text']
    text = clip.tokenize([search_text]).to(device)

    with torch.no_grad():
        text_features = model.encode_text(text)
        text_features /= text_features.norm(dim=-1, keepdim=True)

    image_features = torch.from_numpy(numpy.array(request_data['image_features'])).to(device, dtype=torch.float32)
    similarity = (image_features @ text_features.T).squeeze(1).cpu().numpy()
    print(f"Similarity: {similarity}", flush=True)

    return jsonify({"similarity": similarity.tolist()})

if __name__ == '__main__':
    app.run(debug=False, host='0.0.0.0', port=5000)
