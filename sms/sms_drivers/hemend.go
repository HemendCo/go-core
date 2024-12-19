package sms_drivers

import (
	"HemendCo/go-core"
	"HemendCo/go-core/cache"
	"HemendCo/go-core/sms/sms_models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HemendSMSResponse struct {
	StatusCode    string `json:"status_code"`
	StatusMessage string `json:"status_message"`
}

// HemendSMSDriver for logging to a file
type HemendSMSDriver struct {
	app   *core.App
	cfg   *sms_models.HemendSMSConfig
	cache cache.CacheDriver
}

// Name implements the name for the driver
func (hs *HemendSMSDriver) Name() string {
	return "hemend"
}

// Init creates a new HemendSMSDriver and configures it with settings
func (hs *HemendSMSDriver) Init(app *core.App, config interface{}) error {
	if hs.app != nil {
		return nil
	}

	cfg, ok := config.(sms_models.HemendSMSConfig)
	if !ok {
		return errors.New("invalid file logger configuration: expected a logger_models.FileLoggerConfig type")
	}

	// Attempt to get the service from the app
	service, err := app.Get(core.CacheKeyword)
	if err != nil {
		return fmt.Errorf("failed to get cache service: %w", err)
	}

	// Attempt to assert the serviceValue to type T
	cache, ok := service.(cache.CacheDriver)
	if !ok {
		return fmt.Errorf("unsupported cache service: expected type .CacheDriver but got %T", service)
	}

	hs.app = app
	hs.cfg = &cfg
	hs.cache = cache

	return nil
}

// SendMessage sends an SMS message
func (hs *HemendSMSDriver) SendMessage(mobileNumber string, message string, sendDateTime *time.Time) (*sms_models.SMSResponse, error) {
	res := sms_models.SMSResponse{}

	token, err := hs.getToken()
	if err != nil {
		res.StatusCode = sms_models.TokenExpiredStatusCode
		return &res, nil
	}

	postData := map[string]interface{}{
		"message":        message,
		"mobile_number":  mobileNumber,
		"send_date_time": sendDateTime,
	}

	url := hs.getAPIMessageSendUrl() + "/message.send"
	hemendResponse, err := hs.execute(postData, url, &token)
	if err != nil {
		res.StatusCode = sms_models.InternalServerErrorStatusCode
		return &res, nil
	}

	smsResponse, ok := hemendResponse.(*HemendSMSResponse)
	if !ok {
		res.StatusCode = sms_models.ResponseParseErrorStatusCode
		return &res, nil
	}

	switch smsResponse.StatusCode {
	case "OK":
		res.StatusCode = sms_models.OkStatusCode
	case "MOBILE_NUMBER_INVALID":
		res.StatusCode = sms_models.MobileInvalidStatusCode
	case "MESSAGE_INVALID":
		res.StatusCode = sms_models.MessageInvalidStatusCode
	case "SEND_DATE_TIME_INVALID":
		res.StatusCode = sms_models.DateTimeInvalidStatusCode
	default:
		res.StatusCode = sms_models.UnknownStatusCode
	}

	return &res, nil
}

// getAPIMessageSendUrl gets the API URL for sending messages
func (hs *HemendSMSDriver) getAPIMessageSendUrl() string {
	baseURL := "https://sms.hemend.com/api/"
	if hs.cfg.IsTest {
		return baseURL + "test/" + hs.cfg.Version
	}
	return baseURL + "main/" + hs.cfg.Version
}

// getToken retrieves the access token from cache or requests a new one
func (hs *HemendSMSDriver) getToken() (string, error) {
	cacheKey := hs.getCacheTokenKey()
	token, err := hs.cache.Get(cacheKey) // Retrieve token from cache

	if err != nil {
		return "", errors.New("invalid token type in cache")
	}

	if token == nil {
		// Token not in cache, request a new one
		postData := map[string]string{
			"api_key":    hs.cfg.ApiKey,
			"secret_key": hs.cfg.SecretKey,
		}

		url := hs.getAPIMessageSendUrl() + "/auth.getToken"
		tokenRes, err := hs.execute(postData, url, nil)
		if err != nil {
			return "", err
		}

		// Validate the response
		if tokenRes.(map[string]interface{})["status_code"] == "OK" {
			token = tokenRes.(map[string]interface{})["token"].(map[string]interface{})["access_token"].(string)
			expiresIn := tokenRes.(map[string]interface{})["token"].(map[string]interface{})["expires_in"].(string)

			location, err := time.LoadLocation(hs.cfg.Timezone)
			if err != nil {
				return "", err
			}

			// Parse the expiration time string to time.Time
			parsedTime, err := time.ParseInLocation(time.RFC3339, expiresIn, location)
			if err != nil {
				return "", err
			}

			// Calculate expiration time for cache, subtracting 60 seconds
			expireTime := parsedTime.Add(-60 * time.Second)

			// Use time.Until to set the cache expiration duration
			if err := hs.cache.Set(cacheKey, token, time.Until(expireTime)); err != nil {
				return "", err
			}
		}
	}

	strToken, ok := token.(string)
	if !ok {
		return "", errors.New("token is not valid")
	}

	return strToken, nil
}

// execute sends an HTTP request to the specified URL with the given postData
func (hs *HemendSMSDriver) execute(postData interface{}, url string, token *string) (interface{}, error) {
	postString, err := json.Marshal(postData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(postString))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != nil {
		req.Header.Set("Authorization", "Bearer "+*token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(body))
	}

	var result interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// getCacheTokenKey returns the cache key for the token
func (hs *HemendSMSDriver) getCacheTokenKey() string {
	if hs.cfg.IsTest {
		return "hemend_sms_token_test"
	}
	return "hemend_sms_token"
}
