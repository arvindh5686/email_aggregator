package provider

import (
	"bytes"
	"email_aggregator/internal/constants"
	"email_aggregator/internal/dto"
	"email_aggregator/internal/http"
	"encoding/json"
	"io/ioutil"
	netHttp "net/http"
)

const SendGrid = "send_grid"

type SendGridProvider struct {
	Base
	APIKey string
}

type RequestBody struct {
	Personalizations []Personalization `json:"personalizations"`
	From             Email             `json:"from"`
	Content          []Content         `json:"content"`
}

type Personalization struct {
	To      []Email `json:"to"`
	Subject string  `json:"subject"`
}

type Email struct {
	Address string `json:"email"`
	Name    string `json:"name"`
}

type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (provider *SendGridProvider) buildRequest(emailDto dto.Email) ([]byte, error) {
	requestBody := RequestBody{}
	toEmail := Email{Address: emailDto.ToAddress, Name: emailDto.ToName}
	personalization := Personalization{
		To:      []Email{toEmail}, // current /email API only supports one TO email address
		Subject: emailDto.Subject,
	}

	requestBody.Personalizations = []Personalization{personalization}
	requestBody.From = Email{Address: emailDto.FromAddress, Name: emailDto.FromName}
	requestBody.Content = []Content{{Type: constants.ContentTypePlainText, Value: emailDto.Body}}
	return json.Marshal(requestBody)
}

// GetRequestSetterOptions - returns a list of []http.RequestOption. this will be invoked for mutating the req object of the http client
func (provider *SendGridProvider) GetRequestSetterOptions(emailDto dto.Email) []http.RequestOption {
	authSetterOption := func(req *netHttp.Request) error {
		req.Header.Set(constants.AuthorizationHeaderKey, constants.Bearer+" "+provider.APIKey)
		req.Header.Set(constants.ContentTypeHeaderKey, constants.ContentTypeApplicationJson)
		return nil
	}

	reqBodySetterOption := func(req *netHttp.Request) error {
		reqBody, err := provider.buildRequest(emailDto)
		if err != nil {
			return err
		}
		// convert the Reader to an io.ReadCloser
		req.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
		return nil
	}

	return []http.RequestOption{authSetterOption, reqBodySetterOption}
}

func (provider *SendGridProvider) GetName() string {
	return SendGrid
}
