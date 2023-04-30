from flask import Flask, request, redirect, render_template, make_response
from flask import send_from_directory
app = Flask(__name__, static_url_path='')
from DB_helper import DB

db=None
@app.route('/login',methods=['POST'])
def save_login_info():  
    if request.method == 'POST':
        content_type = request.headers.get('Content-Type')
        print(content_type)
        if (content_type == 'application/json'):
            json = request.json
            print(json)
            sql="""
                insert into logins (login, password)
                values (%s,%s)
                """
            db.cur.execute(sql,(json['username'],json['password']) )
            db.conn.commit()
            return "", 200
        else:
            return "", 401

@app.route('/url_click')
def url_click_tracking():
    email_id = request.args.get('email_id') 
    link = request.args.get('link') 
    sql="""
            update emails set is_url_clicked=%s where email_id=%s
         """
    db.cur.execute(sql,(True,email_id) )
    db.conn.commit()
    print("url {} clicked in email {}".format(link,email_id))
    return redirect(link)

@app.route('/email_open', methods=['GET'])
def email_open_tracking():
    if request.method == 'GET':
        email_id = request.args.get('email_id')
        print("{} email opened".format(email_id))
        sql="""
            update emails set is_opened=%s where email_id=%s
         """
        db.cur.execute(sql,(True,email_id) )
        db.conn.commit()
        return send_from_directory('pixel_folder', "cat.png")
    
if __name__ == '__main__':
    db=DB()
    app.run()