package provider

import (
	"email_aggregator/internal/dto"
	"email_aggregator/internal/http"
)

type Provider interface {
	GetRequestSetterOptions(dto.Email) []http.RequestOption
	GetUrl() string
	GetName() string
}

type Base struct {
	URL        string
	HttpClient *http.RetryableClient
}

func (base *Base) GetUrl() string {
	return base.URL
}
