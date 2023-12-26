package model

type Data struct {
	Expected string `json:"expected,omitempty"`
	Gotten   string `json:"gotten,omitempty"`
}

type Log struct {
	URL        string              `json:"URL,omitempty"`
	Error      string              `json:"error,omitempty"`
	StatusCode int                 `json:"statusCode,omitempty"`
	Headers    map[string][]string `json:"headers,omitempty"`
	Data       Data                `json:"data,omitempty"`
}
