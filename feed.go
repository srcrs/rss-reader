package main

type feed struct {
	Title  string            `json:"title,omitempty"`
	Link   string            `json:"link"`
	Custom map[string]string `json:"custom,omitempty"`
	Items  []item            `json:"items,omitempty"`
}

type item struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
}
