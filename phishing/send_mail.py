import argparse
import random
import smtplib
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from DB_helper import DB
# Function to generate a random verification code
db=None

def generate_email_id():
    return str(random.randint(100000, 999999))

def create_message(recipient,email_id):
    # Message creation
    msg = MIMEMultipart()
    msg['From'] = 'tsukishimaaa567@mail.ru'
    msg['To'] = recipient
    msg['Subject'] = 'Вы успешно авторизировались на сайте!'
    
    # Adding HTML-code with tracking pixel
    html = """
    <html>   
    <body>
        <img src="http://127.0.0.1:5000/email_open?email_id={id}" width="0" height="0" />
        <a href="http://127.0.0.1:5000/url_click?email_id=12345&link=https://example.com/page">Ссылка</a>
        <p>Благодарим за регистрацию на нашем сайте {id}!</p>
    </body>
    </html>
    """.format(id=email_id)
    msg.attach(MIMEText(html, 'html'))
    return msg

# Function to send verification code to the user via email
def send_verification_code(context):
    recipients=context.recipients
    # Create SMTP client
    print('SMTP client created.')
    server = smtplib.SMTP_SSL('smtp.mail.ru', 465)


    # Login to sender email account
    sender_email = 'tsukishimaaa567@mail.ru'
    sender_password = 'pPHKHS7ELHJeZULJcVcn'
    server.login(sender_email, sender_password)
    print('Login to sender email account completed.')
    sql="""
            insert into emails (email_id,recipient,is_opened,is_url_clicked)
            values (%s,%s,%s,%s)
        """
    
    # Compose email message
    print('Composing email...')
    #message = f'Subject: Verify Your Email\n\nYour verification code is: {code}'
    for recipient in recipients:
        email_id=generate_email_id()
        msg=create_message('toshiro280701@gmail.com',email_id)
        print('Sendimg mail to {}:'.format(recipient))
        print(msg)
        # Send email message
        server.sendmail(sender_email, recipient, msg.as_string())
        print('Email sending completed.')
        db.cur.execute(sql,(email_id,recipient,False,False) )
        db.conn.commit()
    # Close SMTP client
    db.close()
    server.quit()

#send_verification_code(['toshiro280701@gmail.com', 'sinadotlyrchikov@gmail.com'])
if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--recipients', nargs='+', required=True)
    args = parser.parse_args()
    db=DB()
    send_verification_code(args)