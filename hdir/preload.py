# import open_clip
import clip
import torch

# model, _, preprocess = open_clip.create_model_and_transforms('ViT-bigG-14-quickgelu', pretrained='metaclip_fullcc')

device = "cuda" if torch.cuda.is_available() else "cpu"
model, preprocess = clip.load("ViT-B/32", device=device)

