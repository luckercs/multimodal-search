# pip install openai
# pip install dashscope

import dashscope
import json
from http import HTTPStatus

def get_text_embedding_vector(sentences, api_key):
    dashscope.api_key = api_key
    input = [{'text': sentences}]
    resp = dashscope.MultiModalEmbedding.call(model="multimodal-embedding-v1",input=input)
    return str(resp.output.get('embeddings')[0].get('embedding')).replace(', ', ' ')

import dashscope
import base64
import os

def get_image_embedding_vector(imageUrl, api_key):
    dashscope.api_key = api_key
    if imageUrl.startswith("http"):
        image = imageUrl
        input = [{'image': image}]
        resp = dashscope.MultiModalEmbedding.call(model="multimodal-embedding-v1",input=input)
        return str(resp.output.get('embeddings')[0].get('embedding')).replace(', ', ' ')
    else:
        image_path = imageUrl
        image_format = os.path.splitext(image_path)[1].strip('.')
        with open(image_path, "rb") as image_file:
            base64_image = base64.b64encode(image_file.read()).decode('utf-8')
        image_data = f"data:image/{image_format};base64,{base64_image}"
        inputs = [{'image': image_data}]
        resp = dashscope.MultiModalEmbedding.call(model="multimodal-embedding-v1",input=inputs)
        return str(resp.output.get('embeddings')[0].get('embedding')).replace(', ', ' ')

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
        vecstr = get_text_embedding_vector(json_params["data"], json_params["api_key"])
        return JSONResponse(content={"embedding": vecstr})
    except KeyError as e:
        return JSONResponse(content={"error": f"Missing key in JSON parameters: {e}"}, status_code=400)
    except Exception as e:
        return JSONResponse(content={"error": f"An unexpected error occurred: {e}"}, status_code=500)

@app.post("/get_img_vec")
async def embed_img_query(request: Request):
    try:
        json_params = await request.json()
        vecstr = get_image_embedding_vector(json_params["url"], json_params["api_key"])
        return JSONResponse(content={"embedding": vecstr})
    except KeyError as e:
        return JSONResponse(content={"error": f"Missing key in JSON parameters: {e}"}, status_code=400)
    except Exception as e:
        return JSONResponse(content={"error": f"An unexpected error occurred: {e}"}, status_code=500)

if __name__ == '__main__':
    uvicorn.run(app, host="0.0.0.0", port=8010)