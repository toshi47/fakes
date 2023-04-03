package rest

import (
	"api/auth_manager"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"log"
	"net/http"
	"time"
)

const SessionCookieId = "Session-Id"

type (
	Server interface {
		Start()
	}

	server struct {
		port        string
		networkPort string

		e       *echo.Echo
		store   *sessions.CookieStore
		authmgr auth_manager.AuthManager
	}
)

func NewServer(port string, networkPort string, authmgr auth_manager.AuthManager) (Server, error) {
	s := server{
		port:        port,
		networkPort: networkPort,
		e:           echo.New(),
		authmgr:     authmgr,
	}

	hashKey := []byte("my-secret-key-12345") // Replace with your own secret key
	store := sessions.NewCookieStore(hashKey)

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
	s.e.POST("/auth/login", s.login)
	s.e.POST("/auth/register", s.register)

	s.e.GET("/test", test, s.authMiddleware())
	s.e.POST("/predict_text", s.handlePredictText, s.authMiddleware())
	s.e.POST("/predict_link", s.handlePredictLink, s.authMiddleware())
	s.e.POST("/predict_image", s.handlePredictImage, s.authMiddleware())
	s.e.Logger.Fatal(s.e.Start("127.0.0.1:" + s.port))
}

func (s server) authMiddleware() func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session, err := s.store.Get(c.Request(), "session-name")
			if err != nil {
				log.Println(err)
				return c.String(http.StatusInternalServerError, err.Error())
			}

			if session.Values["authenticated"] != true {
				return c.String(http.StatusUnauthorized, "unauthorized")
			}
			return next(c)
		}
	}
}
