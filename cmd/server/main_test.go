package main

import (
	"bytes"
	"email_aggregator/internal/dto"
	"email_aggregator/internal/handler"
	"email_aggregator/internal/http"
	"email_aggregator/internal/http/mocks"
	"email_aggregator/internal/service"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	netHttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmailEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		httpRecorder := httptest.NewRecorder()
		ctx, router := gin.CreateTestContext(httpRecorder)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockHttpClient := mocks.NewMockClient(ctrl)
		bodyReader := strings.NewReader("successful response")
		// convert the Reader to an io.ReadCloser
		resBody := ioutil.NopCloser(bodyReader)
		mockHttpClient.EXPECT().Do(gomock.Any()).Return(&netHttp.Response{StatusCode: netHttp.StatusCreated, Body: resBody}, nil).Times(1)
		emailSender := &service.EmailSenderImpl{HttpClient: &http.RetryableClient{HttpClient: mockHttpClient}, Provider: sendGridProvider}
		router.POST("/email", handler.SendEmailHandler(emailSender))
		emailDto := buildEmailDto()
		reqBody, _ := json.Marshal(emailDto)
		req, _ := netHttp.NewRequestWithContext(ctx, netHttp.MethodPost, "/email", bytes.NewReader(reqBody))
		router.ServeHTTP(httpRecorder, req)
		out, _ := ioutil.ReadAll(httpRecorder.Body)
		logger.Info(string(out))
		assert.Equal(t, netHttp.StatusOK, httpRecorder.Code)
	})

	t.Run("fail_http_500", func(t *testing.T) {
		httpRecorder := httptest.NewRecorder()
		ctx, router := gin.CreateTestContext(httpRecorder)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockHttpClient := mocks.NewMockClient(ctrl)
		bodyReader := strings.NewReader("failed")
		// convert the Reader to an io.ReadCloser
		resBody := ioutil.NopCloser(bodyReader)
		mockHttpClient.EXPECT().Do(gomock.Any()).
			Return(&netHttp.Response{StatusCode: netHttp.StatusInternalServerError, Body: resBody}, errors.New("error occurred")).Times(1)
		emailSender := &service.EmailSenderImpl{HttpClient: &http.RetryableClient{HttpClient: mockHttpClient}, Provider: sendGridProvider}
		router.POST("/email", handler.SendEmailHandler(emailSender))
		emailDto := buildEmailDto()
		reqBody, _ := json.Marshal(emailDto)
		req, _ := netHttp.NewRequestWithContext(ctx, netHttp.MethodPost, "/email", bytes.NewReader(reqBody))
		router.ServeHTTP(httpRecorder, req)
		assert.Equal(t, netHttp.StatusInternalServerError, httpRecorder.Code)
	})

	t.Run("fail_http_400_missing_fields", func(t *testing.T) {
		httpRecorder := httptest.NewRecorder()
		ctx, router := gin.CreateTestContext(httpRecorder)
		emailSender := &service.EmailSenderImpl{}
		router.POST("/email", handler.SendEmailHandler(emailSender))
		emailDto := buildEmailDto()
		emailDto.FromAddress = "" // clear from_address
		reqBody, _ := json.Marshal(emailDto)
		req, _ := netHttp.NewRequestWithContext(ctx, netHttp.MethodPost, "/email", bytes.NewReader(reqBody))
		router.ServeHTTP(httpRecorder, req)
		assert.Equal(t, netHttp.StatusBadRequest, httpRecorder.Code)
		out, _ := ioutil.ReadAll(httpRecorder.Body)
		assert.Contains(t, string(out), "the following fields are missing")

		emailDto.FromAddress = "test123" // invalid email address
		reqBody, _ = json.Marshal(emailDto)
		req, _ = netHttp.NewRequestWithContext(ctx, netHttp.MethodPost, "/email", bytes.NewReader(reqBody))
		router.ServeHTTP(httpRecorder, req)
		assert.Equal(t, netHttp.StatusBadRequest, httpRecorder.Code)
		out, _ = ioutil.ReadAll(httpRecorder.Body)
		assert.Contains(t, string(out), "the following email addresses are invalid")
	})

}

func buildEmailDto() dto.Email {
	return dto.Email{
		ToAddress:   "test1@gmail.com",
		ToName:      "test1",
		FromAddress: "test2@gmail.com",
		FromName:    "test2",
		Subject:     "Test Message",
		Body:        "Test Body",
	}
}

func buildInvalidRequestBody() []byte {
	emailDto := dto.Email{
		ToName:      "test1",
		FromAddress: "test2@gmail.com",
		FromName:    "test2",
		Subject:     "Test Message",
		Body:        "Test Body",
	}

	body, _ := json.Marshal(emailDto)
	return body
}
