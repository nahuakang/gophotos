package views

import "log"

const (
	// AlertLvlError represents Bootstrap alert danger
	AlertLvlError = "danger"
	// AlertLvlWarning represents Bootstrap alert warning
	AlertLvlWarning = "warning"
	// AlertLvlInfo represents Bootstrap alert info
	AlertLvlInfo = "info"
	// AlertLvlSuccess represents Bootstrap alert success
	AlertLvlSuccess = "success"
	// AlertMsgGeneric is displayed when any random error is encountered on backend
	AlertMsgGeneric = "Something went wrong. Please try again and" +
		"contact us if the problem persists."
)

// PublicError is the interface for public error messages
type PublicError interface {
	error // Note that error type is a Go interface
	Public() string
}

// Data is the top level structure that views expect data to come in from.
type Data struct {
	Alert *Alert
	Yield interface{}
}

// SetAlert sets an error as alert
func (d *Data) SetAlert(err error) {
	var msg string

	if pErr, ok := err.(PublicError); ok {
		msg = pErr.Public()
	} else {
		log.Println(err)
		msg = AlertMsgGeneric
	}

	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}

// Alert is used to render Bootstrap Alert messages in templates.
type Alert struct {
	Level   string
	Message string
}
