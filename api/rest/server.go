package rest

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"

	"api/auth_manager"
)

const (
	sessionUsername         = "username"
	sessionFirstFactorAuth  = "first-factor-authenticated"
	sessionSecondFactorAuth = "second-factor-authenticated"
)

type (
	Server interface {
		Start()
	}

	server struct {
		address        string
		networkAddress string

		e       *echo.Echo
		store   *sessions.CookieStore
		authmgr auth_manager.AuthManager
	}
)

func NewServer(address string, networkAddress string, storeHashKey []byte, authmgr auth_manager.AuthManager) (Server, error) {
	s := server{
		address:        address,
		networkAddress: networkAddress,
		e:              echo.New(),
		authmgr:        authmgr,
	}

	store := sessions.NewCookieStore(storeHashKey)

	store.Options = &sessions.Options{
		Path:     "/",
		Domain:   "",
		MaxAge:   int(time.Minute.Seconds() * 15),
		Secure:   false,
		HttpOnly: true,
		SameSite: 0,
	}
	s.store = store

	return s, nil
}

func (s server) Start() {
	s.e.Static("/", "./static")

	s.e.POST("/auth", s.handleAuth)
	s.e.POST("/auth/register", s.handleRegister)
	s.e.POST("/auth/login", s.handleLogin)
	s.e.POST("/auth/confirm", s.handleConfirm)

	s.e.POST("/predict_text", s.handlePredictText, s.authMiddleware())
	s.e.POST("/predict_link", s.handlePredictLink, s.authMiddleware())
	s.e.POST("/predict_image", s.handlePredictImage, s.authMiddleware())

	s.e.Logger.Fatal(s.e.Start(s.address))
}

func (s server) authMiddleware() func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session, err := s.store.Get(c.Request(), "session-name")
			if err != nil {
				log.Error(err)
				return c.String(http.StatusInternalServerError, err.Error())
			}

			log.Debugf("session: %+v", session)

			if session.Values[sessionFirstFactorAuth] != true || session.Values[sessionSecondFactorAuth] != true {
				return c.String(http.StatusUnauthorized, "unauthorized")
			}
			return next(c)
		}
	}
}
