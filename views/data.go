package views

const (
	// AlertLvlError represents Bootstrap alert danger
	AlertLvlError = "danger"
	// AlertLvlWarning represents Bootstrap alert warning
	AlertLvlWarning = "warning"
	// AlertLvlInfo represents Bootstrap alert info
	AlertLvlInfo = "info"
	// AlertLvlSuccess represents Bootstrap alert success
	AlertLvlSuccess = "success"
)

// Data is the top level structure that views expect data to come in from.
type Data struct {
	Alert *Alert
	Yield interface{}
}

// Alert is used to render Bootstrap Alert messages in templates.
type Alert struct {
	Level   string
	Message string
}