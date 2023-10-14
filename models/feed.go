package models

type Feed struct {
	Title  string            `json:"title,omitempty"`
	Link   string            `json:"link"`
	Custom map[string]string `json:"custom,omitempty"`
	Items  []Item            `json:"items,omitempty"`
}

type Item struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
}
