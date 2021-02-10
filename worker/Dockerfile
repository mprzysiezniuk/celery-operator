FROM python:latest

ADD requirements.txt /app/requirements.txt

ADD app/* /app/

WORKDIR /app/

RUN pip install -r requirements.txt
