import json
from flask import Flask, request
from flask import render_template, make_response
import tasks
import os
from PIL import Image
from datetime import datetime
import random, time

APP = Flask(__name__)

@APP.route('/',methods = ['GET','POST'])
def index(): 
    '''
    Render Home Template and Post request to Upload the image to Celery task.
    '''
    if request.method == 'GET':
        return render_template("index.html")
    if request.method == 'POST':
        result = tasks.longtime_add.delay(random.randint(0,100),5)
        return render_template("index.html")

if __name__ == '__main__':
    APP.run(host='0.0.0.0')
