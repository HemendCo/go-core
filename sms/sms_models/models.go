package sms_models

type StatusCode int

const (
	_ StatusCode = iota
	OkStatusCode
	MobileInvalidStatusCode
	MessageInvalidStatusCode
	DateTimeInvalidStatusCode
	ResponseParseErrorStatusCode
	TokenExpiredStatusCode
	ServiceUnavailableStatusCode
	InternalServerErrorStatusCode
	UnknownStatusCode
)

var ErrorMessages = map[StatusCode]string{
	OkStatusCode:                  "Request was successful.",
	MobileInvalidStatusCode:       "The provided mobile number is invalid.",
	MessageInvalidStatusCode:      "The message content is invalid.",
	DateTimeInvalidStatusCode:     "The send date/time is invalid.",
	ResponseParseErrorStatusCode:  "Failed to parse response from SMS service.",
	TokenExpiredStatusCode:        "The token has expired.",
	ServiceUnavailableStatusCode:  "The service is currently unavailable.",
	InternalServerErrorStatusCode: "An internal server error has occurred.",
	UnknownStatusCode:             "An unknown error occurred.",
}

type SMSResponse struct {
	StatusCode StatusCode
}

type HemendSMSConfig struct {
	ApiKey    string
	SecretKey string
	Version   string
	IsTest    bool
	Timezone  string
}
