from __future__ import absolute_import
import time
from celery import Celery

CELERY= Celery('tasks',
             broker='amqp://guest:guest@celery-rabbitmq:5672/',
             backend='rpc://')


@CELERY.task
def longtime_add(x, y):
    print('long time task begins')
    # sleep 5 seconds
    time.sleep(5)
    print('long time task finished')
    return x + y