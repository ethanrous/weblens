import open_clip
import torch

device = "cuda" if torch.cuda.is_available() else "cpu"
model, _, preprocess = open_clip.create_model_and_transforms('hf-hub:laion/CLIP-ViT-H-14-laion2B-s32B-b79K', device=device)

# model, preprocess = clip.load("ViT-B/32", device=device)

