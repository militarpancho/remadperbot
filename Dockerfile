FROM python:3.7.2


ENV PYTHONWARNINGS="ignore:Unverified HTTPS request"

COPY ./requirements.txt /usr/src/app/requirements.txt

WORKDIR /usr/src/app

RUN pip install -r requirements.txt

COPY . /usr/src/app

ENTRYPOINT [ "python" ]

CMD [ "remadperbot.py" ]
