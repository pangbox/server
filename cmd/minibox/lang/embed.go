package lang

import (
	"embed"
)

//go:embed *.json
var TranslationFS embed.FS

type Message struct {
	Source      string   `json:"Source"`
	Context     []string `json:"Context,omitempty"`
	Translation string   `json:"Translation"`
}

type Translation struct {
	Messages []*Message `json:"Messages"`
}
