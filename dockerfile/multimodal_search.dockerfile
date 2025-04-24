FROM bitnami/python:3.13.3

RUN pip3 config set global.index-url https://mirrors.aliyun.com/pypi/simple/
RUN pip3 install --upgrade pip -i https://mirrors.aliyun.com/pypi/simple/
RUN pip3 install openai dashscope fastapi uvicorn
COPY model/model_embed_online.py  /app/model_embed_online.py

COPY server-fe/dist /app/dist
COPY server-be/multimodal_search /app/
RUN chmod +x /app/multimodal_search

COPY dockerfile/entrypoint.sh /app/

WORKDIR /app
VOLUME /app/uploads
EXPOSE 8081
CMD [ "/bin/sh", "/app/entrypoint.sh" ]
