from torchvision import models, transforms
import torch
from PIL import Image
from flask import Flask, request
import io
import os

transform = transforms.Compose(
    [
        transforms.Resize(256),
        transforms.CenterCrop(224),
        transforms.ToTensor(),
        transforms.Normalize(mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225]),
    ]
)

alexnet = models.alexnet(weights=None)
with open("imagenet_classes.txt") as f:
    labels = [line.strip() for line in f.readlines()]

app = Flask(__name__)


@app.route("/recognize", methods=["POST"])
def parseImageEndpoint():
    stuff = parseImage(request.get_data())
    return stuff


@app.route("/ping", methods=["GET"])
def ping():
    return {"pong": True}


def parseImage(imgBytes: bytes) -> str:
    img = Image.open(io.BytesIO(imgBytes))
    img_t = transform(img)
    batch_t = torch.unsqueeze(img_t, 0)
    alexnet.eval()
    out = alexnet(batch_t)
    print(out.shape)

    _, indices = torch.sort(out, descending=True)
    percentage = torch.nn.functional.softmax(out, dim=1)[0] * 100
    # tags = [(labels[idx], percentage[idx].item()) for idx in indices[0][:5]]
    tags = [labels[idx].split(" ")[1] for idx in indices[0][:5]]
    return tags


def main():
    try:
        port = os.getenv("RECOG_PORT")
        if port:
            port = int(port)
        else:
            port = 8082
        app.run(debug=True, host="0.0.0.0", port=port)
    except Exception as e:
        print("Dead", e)
        exit(0)


if __name__ == "__main__":
    main()
