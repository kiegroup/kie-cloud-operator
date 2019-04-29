package web

type Form struct {
	Pages []Page `json:"pages"`
}

type Page struct {
	Fields  []Field  `json:"fields"`
	Buttons []Button `json:"buttons"`
}

type Field struct {
	Label    string `json:"label"`
	Default  string `json:"default"`
	Required bool   `json:"required"`
	JSONPath string `json:"jsonPath"`
}

type Button struct {
	Label  string     `json:"label"`
	Action ActionType `json:"action"`
}

type ActionType string

const (
	Next   ActionType = "next"
	Back   ActionType = "back"
	Cancel ActionType = "cancel"
	Submit ActionType = "submit"
)
