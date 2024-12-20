package sms

import (
	"github.com/HemendCo/go-core"
	"github.com/HemendCo/go-core/sms/sms_models"
	"time"
)

type SMSDriver interface {
	Name() string
	Init(app *core.App, config interface{}) error
	SendMessage(mobileNumber string, message string, sendDateTime *time.Time) (*sms_models.SMSResponse, error)
}
