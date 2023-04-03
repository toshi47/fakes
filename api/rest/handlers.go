package rest

import (
	"github.com/labstack/echo"
	"log"
	"net/http"
)

func (s server) login(c echo.Context) error {
	credentials := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	err := c.Bind(&credentials)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	err = s.authmgr.Login(credentials.Username, credentials.Password)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	session, err := s.store.Get(c.Request(), "session-name")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	session.Values["authenticated"] = true

	err = session.Save(c.Request(), c.Response().Writer)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "logged in successfully!")
}

func (s server) register(c echo.Context) error {
	return c.String(http.StatusOK, "test!")
}

func test(c echo.Context) error {
	return c.String(http.StatusOK, "test!")
}
