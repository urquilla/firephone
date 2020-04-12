package firephone

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// doHTTPReq created as var to allow easier testing
var doHTTPReq = func(client *http.Client, req *http.Request) (*http.Response, error) {
	return client.Do(req)
}

func executeRequest(client *http.Client, req *http.Request) ([]byte, int, error) {
	logrus.Debugf("Executing http request: %v", req)
	resp, err := doHTTPReq(client, req)

	if err != nil {
		return nil, -1, errors.Wrap(err, "executeRequest: Could't communicate with verification endpoint")
	}

	// Reading request body
	body, statusCode, err := readResponse(resp)
	if err != nil {
		return nil, -1, errors.Wrap(err, "executeRequest: Failed to read server response")
	}
	return body, statusCode, nil
}

func readResponse(resp *http.Response) ([]byte, int, error) {
	logrus.Debugf("Received http response with status %s", resp.Status)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	logrus.Debugf("Received http response %s", string(body))
	if err != nil {
		return []byte{}, -1, errors.Wrap(err, "readResponse: Failed to read the server response")
	}

	return body, resp.StatusCode, nil
}

func getDefaultHttpClient() *http.Client {
	return &http.Client{
		Timeout: 20 * time.Second,
	}
}
