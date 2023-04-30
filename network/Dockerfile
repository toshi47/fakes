FROM python:3.8

RUN apt-get update

WORKDIR /network

# deployment server does not support AVR instructions. to use tensorflow > 1.5.0 must use custom build
# got it from here https://www.dropbox.com/s/rnmn205f1skj139/tensorflow-2.3.0-cp38-cp38-linux_x86_64.whl
COPY ./tensorflow-2.3.0-cp38-cp38-linux_x86_64.whl ./
RUN pip3 install tensorflow-2.3.0-cp38-cp38-linux_x86_64.whl

COPY ./requirements.txt ./requirements.txt
RUN pip3 install -r ./requirements.txt

ENV PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION=python

COPY ./models ./models
ARG TEXT_DATASET_PATH
ARG TEXT_MODEL_PATH
ARG TEXT_PREPROCESSOR_PATH
RUN python3 models/text_svm.py

COPY ./ ./

CMD ["python3", "server/server.py"]