package websocket

type Message struct {
	Type    string      `json:"type"`
	From    string      `json:"from,omitempty"`
	Content string      `json:"content,omitempty"`
	SDP     interface{} `json:"sdp,omitempty"`
	ICE     interface{} `json:"ice,omitempty"`
	Action  string      `json:"action,omitempty"`
}
