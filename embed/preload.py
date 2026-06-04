from transformers import AutoModel
import torch

MODEL_ID = "jinaai/jina-clip-v2"

device = "cuda" if torch.cuda.is_available() else "cpu"
model = AutoModel.from_pretrained(MODEL_ID, trust_remote_code=True).to(device)
