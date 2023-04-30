from flask import Flask, jsonify, request, make_response
from joblib import load
import keras
import re
from bs4 import BeautifulSoup
from urllib.parse import urlparse
import requests
import base64
from PIL import Image
from io import BytesIO
from flask_restful import Resource, Api
from fake_detection import TextFakeDetector,ImageFakeDetector
from decouple import config




TXT_DET = None
IMG_DET = None

PORT = config('NN_PORT')
SECRET = config('SECRET_KEY')
TEXT_PREPROCESSOR_PATH=config('TEXT_PREPROCESSOR_PATH')
TEXT_MODEL_PATH=config('TEXT_MODEL_PATH')
IMG_MODEL_PATH=config('IMG_MODEL_PATH')

app = Flask(__name__)
api = Api(app)
app.secret_key = SECRET

def init():
    print(PORT,SECRET)
    global TXT_DET
    global IMG_DET
    global STATE
    #load models
    text_clf = load(TEXT_MODEL_PATH)
    img_clf=keras.models.load_model(IMG_MODEL_PATH)
    TXT_DET=TextFakeDetector(text_clf,TEXT_PREPROCESSOR_PATH)
    IMG_DET=ImageFakeDetector(img_clf)

class TextHandler(Resource):
    def post(self):
        print("Entering text handler")
        json=request.get_json()
        text=json['data']
        print(json['data'])
        try:
            is_fake=TXT_DET.predict(text)
            print(f"Text detector answer: fake {is_fake}")
            return jsonify({"is_fake": bool(is_fake)})
        except Exception as e:
            return make_response(jsonify({"answer":e}),500)

class ImageHandler(Resource):
    def post(self):
        print("Entering image handler")
        json=request.get_json()
        try:
            image_data = re.sub('^data:image/.+;base64,', '', json['data'])
            img = Image.open(BytesIO(base64.b64decode(image_data)))
            is_fake, probability =IMG_DET.predict(img)
            print(f"Image detector answer: fake {is_fake} probability {probability}")
            return jsonify({"is_fake":bool(is_fake), "probability": probability})
        except Exception as e:
            return make_response(jsonify({"answer":e}),500)
            
class LinkHandler(Resource):
    def post(self):
        print("Entering link handler")
        json=request.get_json()
        data=json['data']
        all_sourses = {'panorama.pub':'entry-contents', 'lenta.ru':'topic-body__content'}
        if re.match("^https?:\\/\\/(?:www\\.)?[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}\\b(?:[-a-zA-Z0-9()@:%_\\+.~#?&\\/=]*)$",data) is None:
            return make_response(jsonify({"answer":"invalid link"}), 500)
        else:
            try:
                all_text = ''
                source=data
                u = urlparse(source)
                html_text = requests.get(source).text
                soup = BeautifulSoup(html_text, 'html.parser')
                all_text += soup.body.find('h1').text.replace('\n', "")
                if u.netloc in all_sourses.keys():
                    all_text += soup.body.find('div', attrs={'class': all_sourses[u.netloc]}).text.replace('\n', "")
                print('Link text:', all_text)
                is_fake=TXT_DET.predict(all_text)
                print(f"Text detector answer: fake {is_fake}")
                return jsonify({"is_fake": bool(is_fake)})
            except Exception as e:
                return make_response(jsonify({"answer":e}), 500)


api.add_resource(TextHandler,'/predict_text')
api.add_resource(ImageHandler,'/predict_img')
api.add_resource(LinkHandler,'/predict_link')


if __name__ == '__main__':
    init()
    app.run(debug=True,port=PORT)