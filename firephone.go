package firephone

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const SendVerificationCodeEndpoint = "https://www.googleapis.com/identitytoolkit/v3/relyingparty/sendVerificationCode"
const VerifyPhoneNumberEndpoint = "https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPhoneNumber"

type VerificationClient interface {
	StartVerification(phoneNumber, recaptchaToken string) (string, error)
	CompleteVerification(sessionInfo, code string) (*VerificationInfo, error)
}

type verificationClient struct {
	apiKey         string
	httpClient     *http.Client
	executeRequest func(*http.Client, *http.Request) ([]byte, int, error)
}

func NewVerificationClient(apiKey string, httpClient *http.Client) (VerificationClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("Empty apiKey received")
	}
	if httpClient == nil {
		httpClient = getDefaultHttpClient()
	}
	return &verificationClient{
		apiKey:         apiKey,
		httpClient:     httpClient,
		executeRequest: executeRequest,
	}, nil
}

func (cv *verificationClient) CompleteVerification(sessionInfo, code string) (*VerificationInfo, error) {
	logrus.Debugf("Starting verification for session id %s", sessionInfo)

	// Creating http request
	req, err := createVerificationRequest(VerifyPhoneNumberEndpoint, cv.apiKey, &completeVerificationRequestBody{
		SessionInfo: sessionInfo,
		Code:        code,
	})
	if err != nil {
		return nil, errors.Wrap(err, "CompleteVerification: Failed to create verification request")
	}

	// Execute request
	body, statusCode, err := cv.executeRequest(cv.httpClient, req)
	if err != nil {
		return nil, errors.Wrap(err, "CompleteVerification: Failed to execute request against server")
	}
	bodyStr := string(body)

	err = validateResponse(statusCode, bodyStr)
	if err != nil {
		// Not wrapping, error contains type that we want to preserve
		return nil, err
	}
	completeVerificationResp := &completeVerificationResponse{}
	err = json.Unmarshal(body, completeVerificationResp)
	if err != nil {
		return nil, errors.Wrap(err, "CompleteVerification: Failed to parse server response")
	}

	return &VerificationInfo{
		IDToken:     completeVerificationResp.IDToken,
		IsNewUser:   completeVerificationResp.IsNewUser,
		PhoneNumber: completeVerificationResp.PhoneNumber,
	}, nil
}

func (cv *verificationClient) StartVerification(phoneNumber, recaptchaToken string) (string, error) {
	logrus.Debugf("Starting verification for phone number %s", phoneNumber)

	// Creating http request
	req, err := createVerificationRequest(SendVerificationCodeEndpoint, cv.apiKey, &startVerificationRequestBody{
		PhoneNumber:    phoneNumber,
		RecaptchaToken: recaptchaToken,
	})

	if err != nil {
		return "", errors.Wrap(err, "StartVerification: Failed to create verification request")
	}

	// Execute request
	body, statusCode, err := cv.executeRequest(cv.httpClient, req)
	if err != nil {
		return "", errors.Wrap(err, "StartVerification: Failed to execute request against server")
	}
	bodyStr := string(body)

	err = validateResponse(statusCode, bodyStr)
	if err != nil {
		// Not wrapping, error contains type that we want to preserve
		return "", err
	}

	startVerificationResp := &startVerificationResponse{}
	err = json.Unmarshal(body, startVerificationResp)
	if err != nil {
		return "", errors.Wrap(err, "StartVerification: Failed to parse server response")
	}

	return startVerificationResp.SessionInfo, nil
}

func validateResponse(statusCode int, response string) error {
	// Handle http status codes
	if statusCode == http.StatusForbidden {
		return &Err{
			Message:   fmt.Sprintf("validateResponse: Forbidden (%d), likely missing key. body: %s", statusCode, response),
			ErrorCode: ErrorCodeInvalidAPIKey,
		}
	}

	// Communicate upstream if captcha is invalid
	// TODO: Parse the body to check response fields more accurately
	if statusCode == http.StatusBadRequest && strings.Contains(response, "CAPTCHA_CHECK_FAILED") {
		return &Err{
			Message:   fmt.Sprintf("validateResponse: Bad request (%d), captcha verification failed. body: %s", statusCode, response),
			ErrorCode: ErrorCodeInvalidCaptchaKey,
		}
	}

	// Communicate upstream if captcha is invalid
	// TODO: Parse the body to check response fields more accurately
	if statusCode == http.StatusBadRequest && strings.Contains(response, "INVALID_CODE") {
		return &Err{
			Message:   fmt.Sprintf("validateResponse: Bad request (%d), invalid confirmation code. body: %s", statusCode, response),
			ErrorCode: ErrorCodeInvalidConfirmationCode,
		}
	}

	if statusCode != http.StatusOK {
		return errors.Errorf("validateResponse: Unexpected error, received (%d). body: %s", statusCode, response)
	}
	return nil
}

func createVerificationRequest(apiEndpoint, apiKey string, data interface{}) (*http.Request, error) {
	// Parsing endpoint url
	endpointURL, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "createVerificationRequest: Failed to parse api endpoint url %s", apiEndpoint)
	}

	// Adding api key to url query
	q := endpointURL.Query()
	q.Add("key", apiKey)
	endpointURL.RawQuery = q.Encode()

	// Preparing request body
	jsonString, err := json.Marshal(data)

	if err != nil {
		return nil, errors.Wrap(err, "createVerificationRequest: Couldn't create the verification request json body")
	}

	// Creating request
	req, err := http.NewRequest("POST", endpointURL.String(), bytes.NewBuffer(jsonString))

	if err != nil {
		return nil, errors.Wrap(err, "createVerificationRequest: Couldn't create a send verification request")
	}

	// Adding content headers
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-type", "application/json")

	return req, nil
}
