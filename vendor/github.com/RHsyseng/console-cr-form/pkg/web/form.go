package web

type Form struct {
	Pages []Page `json:"pages,omitempty"`
}

type Page struct {
	Label    string   `json:"label,omitempty"`
	Fields   []Field  `json:"fields,omitempty"`
	Buttons  []Button `json:"buttons,omitempty"`
	SubPages []Page   `json:"subPages,omitempty"`
}

type Field struct {
	Label            string  `json:"label,omitempty"`
	Default          string  `json:"default,omitempty"`
	Description      string  `json:"description,omitempty"`
	Type             string  `json:"type,omitempty"`
	Required         bool    `json:"required,omitempty"`
	JSONPath         string  `json:"jsonPath,omitempty"`
	Min              int     `json:"min,omitempty"`
	Max              int     `json:"max,omitempty"`
	OriginalJSONPath string  `json:"originalJsonPath,omitempty"`
	Visible          bool    `json:"visible,omitempty"`
	DisplayWhen      string  `json:"displayWhen,omitempty"`
	Fields           []Field `json:"fields,omitempty"`
}

type Button struct {
	Label  string     `json:"label,omitempty"`
	Action ActionType `json:"action,omitempty"`
}

type ActionType string

const (
	Next   ActionType = "next"
	Back   ActionType = "back"
	Cancel ActionType = "cancel"
	Submit ActionType = "submit"
)
