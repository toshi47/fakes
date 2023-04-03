package auth_manager

import (
	"errors"
)

type (
	AuthManager interface {
		Login(username string, password string) error
		Register(username string, pasword string, email string) error
	}

	authManager struct {
		// db mock
		users []user
	}

	user struct {
		username string
		password string
		email    string
	}
)

func New() (AuthManager, error) {
	a := authManager{
		users: []user{
			{
				"test",
				"password",
				"kek@gmail.com",
			},
		},
	}
	return a, nil
}

func (a authManager) Login(username string, password string) error {
	for _, user := range a.users {
		if user.username == username {
			if user.password == password {
				return nil
			}
			return errors.New("invalid password")
		}
	}
	return errors.New("user not found")
}
func (a authManager) Register(username string, pasword string, email string) error {
	for _, user := range a.users {
		if user.username == username {
			return errors.New("user already exists")
		}
	}
	a.users = append(a.users, user{
		username: username,
		password: pasword,
		email:    email,
	})
	return nil
}
