package httprequester

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var ErrNot2xx = errors.New("http: status code is not 2xx")

type HTTPRequester interface {
	Get(path string, result any) error
	Post(path string, reqBody any, options ...func(*post)) error
	Patch(path string, reqBody any, options ...func(*patch)) error
}

type httpRequester struct {
	client  *http.Client
	baseURL string
}

func New(options ...func(*httpRequester)) HTTPRequester {
	hr := &httpRequester{}
	for _, o := range options {
		o(hr)
	}
	hr.client = http.DefaultClient
	return hr
}

func WithBaseURL(baseURL string) func(*httpRequester) {
	return func(httpreq *httpRequester) {
		httpreq.baseURL = baseURL
	}
}

func (hr *httpRequester) Get(path string, result any) error {
	res, err := http.Get(hr.baseURL + path)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if result != nil {
		if res.StatusCode == http.StatusOK {
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}
			return json.Unmarshal(resBody, &result)
		} else {
			return ErrNot2xx
		}
	}
	return nil
}

type post struct {
	result      any
	contentType string
}

func PostWithContentType(contentType string) func(*post) {
	return func(p *post) {
		p.contentType = contentType
	}
}

func PostWithResult(result any) func(*post) {
	return func(p *post) {
		p.result = result
	}
}

func (hr *httpRequester) Post(path string, reqBody any, options ...func(*post)) error {
	p := &post{contentType: "application/json", result: nil}
	for _, o := range options {
		o(p)
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	res, err := http.Post(hr.baseURL+path, p.contentType, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if p.result != nil {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return json.Unmarshal(resBody, &p.result)
	}
	return nil
}

type patch struct {
	result      any
	contentType string
}

func PatchWithContentType(contentType string) func(*patch) {
	return func(p *patch) {
		p.contentType = contentType
	}
}

func PatchWithResult(result any) func(*patch) {
	return func(p *patch) {
		p.result = result
	}
}

func (hr *httpRequester) Patch(path string, reqBody any, options ...func(*patch)) error {
	p := &patch{contentType: "application/json", result: nil}
	for _, o := range options {
		o(p)
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPatch, hr.baseURL+path, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", p.contentType)
	req.Close = true
	defer req.Body.Close()

	res, err := hr.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if p.result != nil {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return json.Unmarshal(resBody, &p.result)
	}
	return nil
}
