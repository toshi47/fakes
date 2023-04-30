import time
import pandas as pd
from sklearn import svm
from sklearn.model_selection import train_test_split
from sklearn.metrics import accuracy_score
from sklearn.feature_extraction.text import TfidfVectorizer
from nltk.corpus import stopwords
from nltk.tokenize import word_tokenize
import pymorphy2
import regex as re
import nltk
from decouple import config

nltk.download('punkt')
nltk.download('stopwords')
stop_words = stopwords.words("russian")
morph=pymorphy2.MorphAnalyzer(lang='ru')    

df=pd.read_csv(config('TEXT_DATASET_PATH'))
df.drop(columns=['Unnamed: 0'],inplace=True)
df.dropna(inplace=True)

x = df['texts']
y = df['labels']
x_train, x_test, y_train, y_test = train_test_split(x, y, test_size=0.20)

Tfidf_vect = TfidfVectorizer(max_features=1500)
Tfidf_vect.fit(x)

x_train, x_test, y_train, y_test = train_test_split(x, y, test_size=0.20)
x_train, x_test = Tfidf_vect.transform(x_train), Tfidf_vect.transform(x_test)

ts = time.time()
SVM = svm.SVC()
SVM.fit(x_train,y_train)
timei = time.time() - ts
print('Time -> ', timei)
# predict the labels on validation dataset
y_pred = SVM.predict(x_test)
# Use accuracy_score function to get the accuracy
acci = accuracy_score(y_pred, y_test)*100
print('SVM Accuracy Score -> ', acci)

from joblib import dump, load
dump(SVM, config('TEXT_MODEL_PATH')) 
dump(Tfidf_vect, config('TEXT_PREPROCESSOR_PATH'))