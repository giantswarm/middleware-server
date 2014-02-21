package test

import (
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

func Get(url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	return req
}

func Post(url string, data string, header map[string]string) *http.Request {
	ioReader := strings.NewReader(data)
	req, err := http.NewRequest("POST", url, ioReader)
	if err != nil {
		log.Fatal(err)
	}

	for headerName, headerValue := range header {
		req.Header.Add(headerName, headerValue)
	}

	return req
}

func ProcessRequest(req *http.Request) (*http.Response, string) {
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	return res, string(body)
}

func GetRequest(url string) (int, string, *http.Response) {
	req := Get(url)
	res, body := ProcessRequest(req)

	return res.StatusCode, body, res
}

func PostRequest(url string, data string, header map[string]string) (int, string, *http.Response) {
	req := Post(url, data, header)
	res, body := ProcessRequest(req)

	return res.StatusCode, body, res
}

func CreateServer(router *mux.Router) *httptest.Server {
	return httptest.NewServer(router)
}
