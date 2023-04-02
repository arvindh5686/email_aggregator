package main

import (
	"email_aggregator/internal/handler"
	"email_aggregator/internal/http"
	"email_aggregator/internal/provider"
	"email_aggregator/internal/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	netHttp "net/http"
	"time"
)

var logger *zap.Logger
var mailGunProvider provider.Provider
var sendGridProvider provider.Provider

// TODO: add logic for circuit breaking
func init() {
	viper.AutomaticEnv()
	viper.SetDefault("HTTP_PORT", 8080)
	viper.SetDefault("HTTP_TIMEOUT_MS", 2_000)
	logger, _ = zap.NewProduction()
	viper.SetDefault("MAILGUN_BASE_URL", "https://api.mailgun.net")
	viper.SetDefault("MAILGUN_VERSION", "v3")
	mailGunProvider = buildMailGunProvider()
	viper.SetDefault("SENDGRID_BASE_URL", "https://api.sendgrid.com")
	viper.SetDefault("SENDGRID_VERSION", "v3")
	sendGridProvider = buildSendGridProvider()
	viper.SetDefault("IS_SENDGRID_DEFAULT", true)
}

func buildMailGunProvider() provider.Provider {
	userName := viper.GetString("MAILGUN_USERNAME")
	password := viper.GetString("MAILGUN_PASSWORD")
	baseUrl := viper.GetString("MAILGUN_BASE_URL")
	domain := viper.GetString("MAILGUN_DOMAIN")
	version := viper.GetString("MAILGUN_VERSION")
	url := baseUrl + "/" + version + "/" + domain + "/" + "messages"
	return &provider.MailGunProvider{UserName: userName, Password: password, Base: provider.Base{URL: url}}
}

func buildSendGridProvider() provider.Provider {
	baseUrl := viper.GetString("SENDGRID_BASE_URL")
	apiKey := viper.GetString("SENDGRID_API_KEY")
	version := viper.GetString("SENDGRID_VERSION")
	url := baseUrl + "/" + version + "/" + "mail/send"
	return &provider.SendGridProvider{APIKey: apiKey, Base: provider.Base{URL: url}}
}

func main() {
	router := createWebServer()
	shutdownChan := make(chan struct{})
	port := viper.GetString("HTTP_PORT")
	logger.Info(fmt.Sprintf("starting web server at port: %v", port))
	router.Run(fmt.Sprintf(":%v", port))
	<-shutdownChan
	logger.Info("shutting down server..")
	defer logger.Sync()
}

func createWebServer() *gin.Engine {
	router := gin.Default()
	router.Use(gin.CustomRecovery(panicRecoveryHandler))
	registerEndpoints(router)
	return router
}

func registerEndpoints(router *gin.Engine) {
	// health check endpoint for liveness/readiness checks
	httpTimeoutMs := viper.GetInt64("HTTP_TIMEOUT_MS")
	duration := time.Duration(httpTimeoutMs) * time.Millisecond
	httpClient := http.RetryableClient{HttpClient: &netHttp.Client{Timeout: duration}}
	isSendGridDefault := viper.GetBool("IS_SENDGRID_DEFAULT")
	selectedProvider := sendGridProvider
	if !isSendGridDefault {
		selectedProvider = mailGunProvider
	}

	service := &service.EmailSenderImpl{HttpClient: &httpClient, Provider: selectedProvider}
	router.GET("/health", handler.HealthCheckHandler())
	// Note: AuthHandler is no-op now
	router.POST("/email", handler.AuthHandler(), handler.SendEmailHandler(service))
}

func panicRecoveryHandler(ginCtx *gin.Context, err any) {
	logger.Info(fmt.Sprintf("error occurred processing request - %v", err))
	ginCtx.AbortWithStatus(netHttp.StatusInternalServerError)
}
