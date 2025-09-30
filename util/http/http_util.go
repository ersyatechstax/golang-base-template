package http

import (
	"bytes"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const (
	defaultTimeout = 3000
)

type (
	HTTPUtilI interface {
		DoRequest() (respBody []byte, err error)
	}
	HTTPConfig struct {
		URL           string
		Method        string
		RequestBody   []byte
		RequestHeader map[string]string
		Timeout       int
		HTTPClient    *http.Client
	}
	httpUtil struct {
		HTTPConfig HTTPConfig
	}
)

func NewHTTPUtil(httpConfig HTTPConfig) HTTPUtilI {
	return &httpUtil{
		HTTPConfig: httpConfig,
	}
}

func (h *httpUtil) DoRequest() (respBody []byte, err error) {
	if h.HTTPConfig.Timeout == 0 {
		h.HTTPConfig.Timeout = defaultTimeout
	}

	if h.HTTPConfig.Method == "" {
		return nil, errors.New("http method can't be empty")
	}

	req, err := http.NewRequest(h.HTTPConfig.Method, h.HTTPConfig.URL, bytes.NewBuffer(h.HTTPConfig.RequestBody))
	if err != nil {
		return nil, errors.Errorf("failed when create request: %v", err)
	}

	for key, value := range h.HTTPConfig.RequestHeader {
		req.Header.Set(key, value)
	}

	var resp *http.Response
	resp, err = h.HTTPConfig.HTTPClient.Do(req)
	if err != nil {
		return respBody, errors.Errorf("failed when do request: %v", err)
	}

	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	if resp != nil {
		var errRead error
		respBody, errRead = io.ReadAll(resp.Body)
		if errRead != nil {
			return nil, errors.Errorf("failed when read response body: %s", errRead.Error())
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("got unexpected status code, status_code: %v", resp.StatusCode)
	}

	return respBody, nil
}
