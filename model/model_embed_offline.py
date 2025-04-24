# pip install sentence-transformers transformers
# pip install einops
print("loading model omic-embed-text-v1.5...")
from sentence_transformers import SentenceTransformer
import numpy as np
model_text = SentenceTransformer("/root/models/nomic-ai/nomic-embed-text-v1.5", trust_remote_code=True)

# 返回字符串 "[0.1 0.2 0.3]"
def get_text_embedding_vector(sentences):
    embeddings = model_text.encode(sentences)
    embeddings_str = np.array2string(embeddings,formatter={'float_kind':lambda x: f'{x:.10f}'}).replace('\n', '')
    return embeddings_str

print("loading model omic-embed-vision-v1.5...")
import torch
import torch.nn.functional as F
from transformers import AutoTokenizer, AutoModel, AutoImageProcessor
from PIL import Image
import requests
processor = AutoImageProcessor.from_pretrained("/root/models/nomic-ai/nomic-embed-vision-v1.5")
vision_model = AutoModel.from_pretrained("/root/models/nomic-ai/nomic-embed-vision-v1.5", trust_remote_code=True)

def get_image_embedding_vector(imageUrl):
    if imageUrl.startswith("http"):
        image = Image.open(requests.get(imageUrl, stream=True).raw)
    else:
        image = Image.open(imageUrl)
    inputs = processor(image, return_tensors="pt")
    img_emb = vision_model(**inputs).last_hidden_state
    img_embeddings = F.normalize(img_emb[:, 0], p=2, dim=1)
    img_embeddings_np = img_embeddings.detach().numpy()[0]
    embeddings_str = np.array2string(img_embeddings_np,formatter={'float_kind':lambda x: f'{x:.10f}'}).replace('\n', '')
    return embeddings_str

# pip install fastapi uvicorn
from fastapi import FastAPI
import uvicorn
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

app = FastAPI()

@app.post("/get_txt_vec")
async def embed_txt_query(request: Request):
    try:
        json_params = await request.json()
        vecstr = get_text_embedding_vector(json_params["data"])
        return JSONResponse(content={"embedding": vecstr})
    except KeyError as e:
        return JSONResponse(content={"error": f"Missing key in JSON parameters: {e}"}, status_code=400)
    except Exception as e:
        return JSONResponse(content={"error": f"An unexpected error occurred: {e}"}, status_code=500)

@app.post("/get_img_vec")
async def embed_img_query(request: Request):
    try:
        json_params = await request.json()
        vecstr = get_image_embedding_vector(json_params["url"])
        return JSONResponse(content={"embedding": vecstr})
    except KeyError as e:
        return JSONResponse(content={"error": f"Missing key in JSON parameters: {e}"}, status_code=400)
    except Exception as e:
        return JSONResponse(content={"error": f"An unexpected error occurred: {e}"}, status_code=500)

if __name__ == '__main__':
    uvicorn.run(app, host="0.0.0.0", port=8010)
