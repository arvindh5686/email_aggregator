package service

import (
	"email_aggregator/internal/dal"
	"email_aggregator/internal/dto"
	"email_aggregator/internal/http"
	"email_aggregator/internal/provider"
	"fmt"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewProduction()
}

type EmailSender interface {
	Send(emailDto dto.Email) error
}

type EmailSenderImpl struct {
	Provider   provider.Provider
	HttpClient *http.RetryableClient
}

// Send - invokes a POST request by using the retryable http client.
func (emailSender *EmailSenderImpl) Send(emailDto dto.Email) error {
	// Note: the below dal call to save is not needed for this excercise but addresses the point about sql injection
	emailDal := dal.EmailDalImpl{}
	emailDal.Save(emailDto)
	options := emailSender.Provider.GetRequestSetterOptions(emailDto)
	err := emailSender.HttpClient.Post(emailSender.Provider.GetUrl(), options...)
	if err != nil {
		logger.Error(fmt.Sprintf("provider - %v, error - %v", emailSender.Provider.GetName(), err))
		return err
	}

	logger.Info("email successfully sent")
	return nil
}
