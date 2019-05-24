FROM python:alpine3.7

WORKDIR /usr/src/bitfinex-crawler

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app/*.py .

CMD ["python", "-u", "crawler.py"]
