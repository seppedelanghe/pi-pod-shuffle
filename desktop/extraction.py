import librosa
import torch

def extract_embeddings(path: str, model):
    audio, _ = librosa.load(path, sr=32000, duration=120, mono=True)
    audio = audio[None, :] # add batch dim: (1, -1)
    device = 'cuda' if torch.cuda.is_available() else 'mps'

    tensor = torch.from_numpy(audio).to(device)
    out = model(tensor)
    embedding = out['embedding']
    return embedding.detach().cpu().numpy().tolist()
    

