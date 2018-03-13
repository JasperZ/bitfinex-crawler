FROM python:alpine3.7

WORKDIR /usr/src/bitfinex-crawler

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY apis .
COPY crawler.py .

CMD ["python", "./crawler.py"]
