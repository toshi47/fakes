package auth_manager

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"

	"github.com/jackc/pgx/v4"
	"github.com/labstack/gommon/random"
	log "github.com/sirupsen/logrus"
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
		dbUrl     string
		email     EmailInfo
		sentCodes map[string]string
	}

	user struct {
		name     string
		password string
		email    string
	}
)

func New(dbUrl string, emailInfo EmailInfo) (AuthManager, error) {
	a := authManager{
		dbUrl:     dbUrl,
		email:     emailInfo,
		sentCodes: map[string]string{},
	}
	return &a, nil
}

func (a *authManager) Register(username string, password string, email string) error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, a.dbUrl)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var u user
	err = conn.QueryRow(ctx, fmt.Sprintf("SELECT name, password, email FROM fakes_user where name='%s' or email='%s' LIMIT 1", username, email)).
		Scan(&u.name, &u.password, &u.email)
	log.Debugf("user returned %+v", u)
	if err == nil {
		if u.name == username {
			return errors.New("user already exists")
		}
		if u.email == email {
			return errors.New("email already registered")
		}
	}
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	_, err = conn.Exec(ctx, fmt.Sprintf("INSERT INTO fakes_user (name, password, email) VALUES ('%s', '%s', '%s')", username, password, email))
	if err != nil {
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (a *authManager) findUser(username string) (user, error) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, a.dbUrl)
	if err != nil {
		return user{}, err
	}
	defer conn.Close(ctx)

	var u user
	err = conn.QueryRow(ctx, fmt.Sprintf("SELECT name, password, email FROM fakes_user where name='%s'", username)).
		Scan(&u.name, &u.password, &u.email)
	log.Debugf("user returned findUser %+v", u)
	if err != nil {
		if err == pgx.ErrNoRows {
			return user{}, fmt.Errorf("user %s does not exist", username)
		}
		return user{}, err
	}
	return u, nil
}

func (a *authManager) Login(username string, password string) error {
	u, err := a.findUser(username)
	if err != nil {
		return err
	}
	if u.password != password {
		return errors.New("invalid password")
	}
	return nil
}

func (a *authManager) SendConfirmationCode(username string) error {
	u, err := a.findUser(username)
	if err != nil {
		return err
	}
	code := random.String(10)
	msg := []byte(fmt.Sprintf("Subject: verification code\nYour verification code is: %q\n", code))

	auth := smtp.PlainAuth("", a.email.Address, a.email.Password, a.email.SmtpHost)
	err = smtp.SendMail(a.email.SmtpHost+":"+a.email.SmtpPort,
		auth, a.email.Address, []string{u.email}, msg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	a.sentCodes[u.name] = code

	log.Debugf("successfully sent verification email to %s", u.email)
	return nil
}

func (a *authManager) CheckConfirmationCode(username string, code string) error {
	realCode, ok := a.sentCodes[username]
	if !ok {
		return errors.New("code was not sent or expired")
	}
	if realCode != code {
		return errors.New("wrong code")
	}
	return nil
}
