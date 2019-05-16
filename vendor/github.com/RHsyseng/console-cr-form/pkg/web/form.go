package web

type Form struct {
	Pages []Page `json:"pages"`
}

type Page struct {
	Label    string   `json:"label"`
	Fields   []Field  `json:"fields"`
	Buttons  []Button `json:"buttons"`
	SubPages []Page   `json:"subPages"`
}

type Field struct {
	Label            string  `json:"label"`
	Default          string  `json:"default"`
	Description      string  `json:"description"`
	Type             string  `json:"type"`
	Required         bool    `json:"required"`
	JSONPath         string  `json:"jsonPath"`
	Min              int     `json:"min"`
	Max              int     `json:"max"`
	OriginalJSONPath string  `json:"originalJsonPath"`
	Fields           []Field `json:"fields"`
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
