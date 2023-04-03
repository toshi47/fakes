package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/labstack/echo"
	"io"
	"log"
	"net/http"
	"regexp"
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

type (
	networkRequest struct {
		Data string `json:"data"`
	}
	networkResponse struct {
		Answer string `json:"answer"`
	}
)

func (s server) handlePredictText(c echo.Context) error {
	check := func(req networkRequest) error {
		if req.Data == "" {
			return errors.New("empty text!")
		}
		return nil
	}
	url := "http://127.0.0.1:" + s.networkPort + "/predict_text"
	return handleNetworkRequest(c, url, check)
}

func (s server) handlePredictLink(c echo.Context) error {
	check := func(req networkRequest) error {
		urlRegexp := regexp.MustCompile(`^(http|https)://[a-zA-Z0-9\-.]+\.[a-zA-Z]{2,}(?::[0-9]+)?(?:/[^\s]*)?$`)
		if !urlRegexp.MatchString(req.Data) {
			return errors.New("invalid link format!")
		}
		return nil
	}
	url := "http://127.0.0.1:" + s.networkPort + "/predict_link"
	return handleNetworkRequest(c, url, check)
}

func (s server) handlePredictImage(c echo.Context) error {
	check := func(req networkRequest) error {
		return nil
	}
	url := "http://127.0.0.1:" + s.networkPort + "/predict_img"
	return handleNetworkRequest(c, url, check)
}

func handleNetworkRequest(c echo.Context, url string, withChecks ...func(req networkRequest) error) error {
	networkReq := networkRequest{}
	err := c.Bind(&networkReq)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	for _, check := range withChecks {
		err := check(networkReq)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
	}

	body, err := json.Marshal(networkReq)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	resp, err := http.Post(url, echo.MIMEApplicationJSON, bytes.NewReader(body))
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, string(respBytes))
	}

	networkResp := networkResponse{}
	err = json.Unmarshal(respBytes, &networkResp)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, networkResp)
}
