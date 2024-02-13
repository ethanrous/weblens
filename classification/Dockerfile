FROM --platform=linux/amd64 python:3.11-slim as build

WORKDIR /app
COPY requirements.txt torch_requirements.txt imageRecog.py imagenet_classes.txt ./
RUN pip install -r requirements.txt
RUN pip install -r torch_requirements.txt

EXPOSE 8080

CMD ["python", "/app/imageRecog.py"]