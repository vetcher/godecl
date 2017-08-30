package types

type Import struct {
	Docs    []string `json:"docs,omitempty"`
	Package string   `json:"package,omitempty"`
	Alias   string   `json:"alias,omitempty"`
}
