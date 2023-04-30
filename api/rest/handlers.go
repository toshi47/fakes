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

type (
	predictRequest struct {
		Data string `json:"data"`
	}
	errorResponse struct {
		Error string `json:"error"`
	}
	predictResponse struct {
		IsFake      bool    `json:"is_fake"`
		Probability float64 `json:"probability,omitempty"`
	}
)

var (
	internalErr = errorResponse{Error: "internal server error"}
)

func (s server) handleRegister(c echo.Context) error {
	req := struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}{}

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	}

	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !emailRegex.MatchString(req.Email) {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid email"})
	}

	err = s.authmgr.Register(req.Username, req.Password, req.Email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	}

	return c.NoContent(http.StatusOK)
}

func (s server) handleAuth(c echo.Context) error {
	session, err := s.store.Get(c.Request(), "session-name")
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusInternalServerError, internalErr)
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
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	}

	err = s.authmgr.Login(req.Username, req.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	}

	session, err := s.store.Get(c.Request(), "session-name")
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusInternalServerError, internalErr)
	}

	session.Values[sessionFirstFactorAuth] = true
	session.Values[sessionUsername] = req.Username

	err = session.Save(c.Request(), c.Response().Writer)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusInternalServerError, internalErr)
	}

	err = s.authmgr.SendConfirmationCode(req.Username)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusInternalServerError, internalErr)
	}

	return c.NoContent(http.StatusOK)
}

func (s server) handleConfirm(c echo.Context) error {
	req := struct {
		Code string `json:"code"`
	}{}

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	}

	session, err := s.store.Get(c.Request(), "session-name")
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusInternalServerError, internalErr)
	}

	if session.Values[sessionFirstFactorAuth] != true {
		return c.NoContent(http.StatusUnauthorized)
	}

	err = s.authmgr.CheckConfirmationCode(session.Values[sessionUsername].(string), req.Code)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: err.Error()})
	}

	session.Values[sessionSecondFactorAuth] = true

	err = session.Save(c.Request(), c.Response().Writer)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusInternalServerError, internalErr)
	}

	return c.NoContent(http.StatusOK)
}

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
		return c.JSON(http.StatusInternalServerError, internalErr)
	}

	for _, check := range withChecks {
		err := check(networkReq)
		if err != nil {
			return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		}
	}

	body, err := json.Marshal(networkReq)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusInternalServerError, internalErr)
	}

	resp, err := http.Post(url, echo.MIMEApplicationJSON, bytes.NewReader(body))
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusInternalServerError, internalErr)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusInternalServerError, internalErr)
	}

	if resp.StatusCode == http.StatusInternalServerError {
		return c.JSON(http.StatusInternalServerError, internalErr)
	}

	var response interface{}
	if resp.StatusCode != http.StatusOK {
		response = errorResponse{}
	} else {
		response = predictResponse{}
	}
	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusInternalServerError, internalErr)
	}
	return c.JSON(resp.StatusCode, response)
}
