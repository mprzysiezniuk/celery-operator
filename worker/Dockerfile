FROM python:latest

ADD requirements.txt /app/requirements.txt

ADD app/* /app/

ADD app/templates /app/templates

WORKDIR /app/

RUN pip install -r requirements.txt
