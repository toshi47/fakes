package auth_manager

import (
	"errors"
	"fmt"
	"net/smtp"

	"github.com/labstack/gommon/log"
	"github.com/labstack/gommon/random"
)

type (
	AuthManager interface {
		Register(username string, pasword string, email string) error
		Login(username string, password string) error
		SendConfirmationCode(username string) error
		CheckConfirmationCode(username string, code string) error
	}

	EmailInfo struct {
		Address  string
		Password string
		SmtpHost string
		SmtpPort string
	}

	authManager struct {
		email     EmailInfo
		sentCodes map[string]string
		// db mock
		users []user
	}

	user struct {
		name     string
		password string
		email    string
	}
)

func New(emailInfo EmailInfo) (AuthManager, error) {
	a := authManager{
		email:     emailInfo,
		users:     []user{},
		sentCodes: map[string]string{},
	}
	return &a, nil
}

func (a *authManager) Register(username string, pasword string, email string) error {
	for _, user := range a.users {
		if user.name == username {
			return errors.New("user already exists")
		}
		if user.email == email {
			return errors.New("email already registered")
		}
	}
	a.users = append(a.users, user{
		name:     username,
		password: pasword,
		email:    email,
	})
	return nil
}

func (a *authManager) Login(username string, password string) error {
	for _, user := range a.users {
		if user.name == username {
			if user.password == password {
				return nil
			}
			return errors.New("invalid password")
		}
	}
	return errors.New("user not found")
}

func (a *authManager) SendConfirmationCode(username string) error {
	for _, user := range a.users {
		if user.name == username {
			code := random.String(10)
			msg := []byte(fmt.Sprintf("Subject: verification code\nYour verification code is: %q\n", code))

			auth := smtp.PlainAuth("", a.email.Address, a.email.Password, a.email.SmtpHost)
			err := smtp.SendMail(a.email.SmtpHost+":"+a.email.SmtpPort,
				auth, a.email.Address, []string{user.email}, msg)
			if err != nil {
				fmt.Println(err)
				return err
			}
			a.sentCodes[user.name] = code

			log.Debugf("successfully sent verification email to %s", user.email)
			return nil
		}
	}
	return errors.New("user not found")
}

func (a *authManager) CheckConfirmationCode(username string, code string) error {
	for _, user := range a.users {
		if user.name == username {
			// send code here
			realCode, ok := a.sentCodes[user.name]
			if !ok {
				return errors.New("code was not sent or expired")
			}
			if realCode != code {
				return errors.New("wrong code")
			}
			return nil
		}
	}
	return errors.New("user not found")
}
