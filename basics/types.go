package driver_tools

// accept msg
type Message struct {
	DB    string `json:"db"`
	Key   string `json:"key"`
	Value []byte `json:"value"`
	Act   int    `json:"act"`
}

// status replay
type ReplayStatus struct {
	Content    []byte `json:"content"`
	StatusCode int    `json:"code"`
	Type       string `json:"type"`
}

type Connector interface {
	Connect() error
	Ping() error
	Close() error
}
