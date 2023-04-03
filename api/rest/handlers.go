package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

func (s server) handleRegister(c echo.Context) error {
	req := struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}{}

	err := c.Bind(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !emailRegex.MatchString(req.Email) {
		return c.String(http.StatusBadRequest, "invalid email")
	}

	err = s.authmgr.Register(req.Username, req.Password, req.Email)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (s server) handleAuth(c echo.Context) error {
	session, err := s.store.Get(c.Request(), "session-name")
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	log.Debugf("session: %+v", session)

	if session.Values[sessionFirstFactorAuth] != true || session.Values[sessionSecondFactorAuth] != true {
		return c.NoContent(http.StatusUnauthorized)
	}

	return c.NoContent(http.StatusOK)
}

func (s server) handleLogin(c echo.Context) error {
	req := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	err := c.Bind(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	err = s.authmgr.Login(req.Username, req.Password)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	session, err := s.store.Get(c.Request(), "session-name")
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	session.Values[sessionFirstFactorAuth] = true
	session.Values[sessionUsername] = req.Username

	err = session.Save(c.Request(), c.Response().Writer)
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	err = s.authmgr.SendConfirmationCode(req.Username)
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (s server) handleConfirm(c echo.Context) error {
	req := struct {
		Code string `json:"code"`
	}{}

	err := c.Bind(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	session, err := s.store.Get(c.Request(), "session-name")
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if session.Values[sessionFirstFactorAuth] != true {
		return c.String(http.StatusUnauthorized, "no confirmation for this session")
	}

	err = s.authmgr.CheckConfirmationCode(session.Values[sessionUsername].(string), req.Code)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}

	session.Values[sessionSecondFactorAuth] = true

	err = session.Save(c.Request(), c.Response().Writer)
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "authorized successfully!")
}

type (
	predictRequest struct {
		Data string `json:"data"`
	}
	predictResponse struct {
		IsFake      bool    `json:"is_fake"`
		Probability float64 `json:"probability,omitempty"`
	}
)

func (s server) handlePredictText(c echo.Context) error {
	check := func(req predictRequest) error {
		if req.Data == "" {
			return errors.New("empty text")
		}
		return nil
	}
	url := "http://" + s.networkAddress + "/predict_text"
	return handleNetworkRequest(c, url, check)
}

func (s server) handlePredictLink(c echo.Context) error {
	check := func(req predictRequest) error {
		urlRegexp := regexp.MustCompile(`^(http|https)://[a-zA-Z0-9\-.]+\.[a-zA-Z]{2,}(?::[0-9]+)?(?:/[^\s]*)?$`)
		if !urlRegexp.MatchString(req.Data) {
			return errors.New("invalid link format")
		}
		return nil
	}
	url := "http://" + s.networkAddress + "/predict_link"
	return handleNetworkRequest(c, url, check)
}

func (s server) handlePredictImage(c echo.Context) error {
	check := func(req predictRequest) error {
		return nil
	}
	url := "http://" + s.networkAddress + "/predict_img"
	return handleNetworkRequest(c, url, check)
}

func handleNetworkRequest(c echo.Context, url string, withChecks ...func(req predictRequest) error) error {
	networkReq := predictRequest{}
	err := c.Bind(&networkReq)
	if err != nil {
		log.Error(err)
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
		log.Error(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	resp, err := http.Post(url, echo.MIMEApplicationJSON, bytes.NewReader(body))
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, string(respBytes))
	}

	predictResp := predictResponse{}
	err = json.Unmarshal(respBytes, &predictResp)
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, predictResp)
}
