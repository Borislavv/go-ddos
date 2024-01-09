package floodermodel

type Data struct {
	Expected interface{} `json:"expected,omitempty"`
	Gotten   interface{} `json:"gotten,omitempty"`
}

type Log struct {
	URL        string              `json:"URL,omitempty"`
	Error      string              `json:"error,omitempty"`
	StatusCode int                 `json:"statusCode,omitempty"`
	Headers    map[string][]string `json:"headers,omitempty"`
	Data       Data                `json:"data,omitempty"`
}
