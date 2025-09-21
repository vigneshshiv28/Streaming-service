package core

type Message struct {
	Type    string      `json:"type"`
	From    string      `json:"from,omitempty"`
	To      string      `json:"to,omitempty"`
	Role    string      `json:"role,omitempty"`
	Name    string      `json:"name,omitempty"`
	Content string      `json:"content,omitempty"`
	SDP     interface{} `json:"sdp,omitempty"`
	ICE     interface{} `json:"ice,omitempty"`
	Action  string      `json:"action,omitempty"`
}
