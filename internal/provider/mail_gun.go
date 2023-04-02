package provider

import (
	"bytes"
	"email_aggregator/internal/constants"
	"email_aggregator/internal/dto"
	"email_aggregator/internal/http"
	"io/ioutil"
	"mime/multipart"
	netHttp "net/http"
)

const MailGun = "mail_gun"

type MailGunProvider struct {
	Base
	UserName string
	Password string
}

// GetRequestSetterOptions - returns a list of []http.RequestOption. this will be invoked for mutating the req object of the http client
func (provider *MailGunProvider) GetRequestSetterOptions(emailDto dto.Email) []http.RequestOption {
	authSetterOption := func(req *netHttp.Request) error {
		req.SetBasicAuth(provider.UserName, provider.Password)
		return nil
	}

	reqBodySetterOption := func(req *netHttp.Request) error {
		// Create a new multipart form and add the fields to it
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add the form fields
		writer.WriteField("from", emailDto.FromAddress)
		writer.WriteField("to", emailDto.ToAddress)
		writer.WriteField("subject", emailDto.Subject)
		writer.WriteField("text", emailDto.Body)
		// Close the multipart form and set the request body
		err := writer.Close()
		if err != nil {
			return err
		}

		contentType := writer.FormDataContentType()
		req.Header.Set(constants.ContentTypeHeaderKey, contentType)
		req.Body = ioutil.NopCloser(body)
		return nil
	}

	return []http.RequestOption{authSetterOption, reqBodySetterOption}
}

func (provider *MailGunProvider) GetName() string {
	return MailGun
}
