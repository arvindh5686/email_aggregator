package handler

import (
	"email_aggregator/internal/dto"
	"email_aggregator/internal/service"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jaytaylor/html2text"
	netHttp "net/http"
	"regexp"
	"time"
)

const EmailRegexMatchPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

var emailRegex *regexp.Regexp

func init() {
	emailRegex, _ = regexp.Compile(EmailRegexMatchPattern)
}

// SendEmailHandler handles incoming http requests for POST /email endpoint
func SendEmailHandler(service service.EmailSender) gin.HandlerFunc {
	return func(context *gin.Context) {
		startTime := time.Now()
		emailDto, err := bindAndValidateInputPayload(context)
		if err != nil {
			logger.Error(err.Error())
			return
		}

		// converts html to text and handles XSS
		body, err := htmlToText(emailDto.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("error occurred parsing email body: %v", err))
			context.AbortWithStatusJSON(netHttp.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		emailDto.Body = *body
		err = service.Send(*emailDto)
		if err != nil {
			context.AbortWithStatusJSON(netHttp.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Info(fmt.Sprintf("email delivered successfully. overall latency: %v", time.Now().Sub(startTime)))
		context.JSON(netHttp.StatusOK, gin.H{"message": "email sent successfully"})
	}
}

// validates for required fields and email validation
func bindAndValidateInputPayload(context *gin.Context) (*dto.Email, error) {
	emailDto := dto.Email{}
	err := context.BindJSON(&emailDto)
	if err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			missingFields := checkMissingFields(errs)
			errorMsg := fmt.Sprintf("the following fields are missing %v", missingFields)
			context.AbortWithStatusJSON(netHttp.StatusBadRequest, gin.H{"error": errorMsg})
			return nil, err
		}

		context.AbortWithStatusJSON(netHttp.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, err
	}

	var invalidEmailAddresses []string
	isValid := validateEmail(emailDto.FromAddress)
	if !isValid {
		invalidEmailAddresses = append(invalidEmailAddresses, emailDto.FromAddress)
	}

	isValid = validateEmail(emailDto.ToAddress)
	if !isValid {
		invalidEmailAddresses = append(invalidEmailAddresses, emailDto.ToAddress)
	}

	if len(invalidEmailAddresses) > 0 {
		msg := fmt.Sprintf("the following email addresses are invalid: %v", invalidEmailAddresses)
		context.AbortWithStatusJSON(netHttp.StatusBadRequest, gin.H{"error": msg})
		return nil, errors.New(msg)
	}

	return &emailDto, nil
}

func checkMissingFields(errors validator.ValidationErrors) []string {
	var missingFields []string
	for _, err := range errors {
		if err.Tag() == "required" {
			missingFields = append(missingFields, err.Field())
		}
	}

	return missingFields
}

func validateEmail(emailAddr string) bool {
	if len(emailAddr) == 0 {
		return false
	}

	return emailRegex.MatchString(emailAddr)
}

func htmlToText(htmlBody string) (*string, error) {
	plainText, err := html2text.FromString(htmlBody)
	if err != nil {
		return nil, err
	}

	return &plainText, nil
}
