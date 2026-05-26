FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y curl locales && rm -rf /var/lib/apt/lists/*
RUN locale-gen en_US.UTF-8
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8
ENV PYTHONIOENCODING=utf-8

COPY ai/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY ai/ .

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]