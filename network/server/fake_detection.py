from nltk.corpus import stopwords
from nltk.tokenize import word_tokenize
import pymorphy2
import regex as re
import nltk

import numpy as np
from PIL import Image, ImageChops, ImageEnhance
from numpy import array

import os
from joblib import load
np.random.seed(2)

class FakeDetection:
    def __init__(self,model):
        self.model = model
    def predict(self,sample):
        answer = self.model.predict(sample)
        return answer
    def preprocessing(self,data):
        pass


class TextFakeDetector(FakeDetection):
    def __init__(self,model,vectorizer_path):
        super().__init__(model)
        self.vectorizer=load(vectorizer_path)
        nltk.download('stopwords')
        nltk.download('punkt')
        self.stop_words = stopwords.words("russian")
        self.morph=pymorphy2.MorphAnalyzer(lang='ru')    

    def predict(self, sample):
        answer = super().predict(self.preprocessing(sample))
        return answer[0]==1
        
    def preprocessing(self, data):
        clean_txt = []
        tokenized_sent = word_tokenize(data.lower().strip(),language="russian")
        tokenized_sent=[self.morph.parse(x)[0].normal_form for x in tokenized_sent if x not in self.stop_words and (re.sub(r'[^\w\s]', '', x)!='') and re.search(r'[0-9]+',x) is None and re.search(r'[_a]+',x) is None]
        clean_txt.append(' '.join(tokenized_sent))
        vectorized_txt = self.vectorizer.transform(clean_txt)
        return vectorized_txt
        
class ImageFakeDetector(FakeDetection):
    def __init__(self,model):
        super().__init__(model)
    def predict(self, sample):
        answer = super().predict(self.preprocessing(sample))
        i=np.argmax(answer[0])
        return i==1, answer[0][i] * 100
    def preprocessing(self, img):
        filename = './tmp/photo'
        resaved_filename = filename + '.resaved.jpg'
        im =img.convert('RGB')
        im.save(resaved_filename, 'JPEG', quality=90)
        resaved_im = Image.open(resaved_filename)
        ela_im = ImageChops.difference(im, resaved_im)
        extrema = ela_im.getextrema()
        max_diff = max([ex[1] for ex in extrema])
        if max_diff == 0:
            max_diff = 1
        scale = 255.0 / max_diff
        ela_im = ImageEnhance.Brightness(ela_im).enhance(scale)
        x = array(ela_im.resize((128, 128))).flatten() / 255.0
        X = np.array(x)
        X=X.reshape(-1, 128, 128, 3)
        os.remove(resaved_filename)
        return X