package renderer

type Badge struct {
	Subject string `json:"subject"`
	Status  string `json:"status"`
	Color   Color  `json:"color"`
	Style   Style  `json:"style"`
}
