package api

type CreateRoomResponse struct {
	UserID      string `json:"userID"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	RoomID      string `json:"roomID"`
	HostURL     string `json:"hostURL"`
	GuestURL    string `json:"joiningURL"`
	AudienceURL string `json:"audienceURL"`
	CreatedAt   string `json:"createdAt"`
}

type JoinRoomResponse struct {
	Status string `json:"status"`
	UserID string `json:"userID"`
	Role   string `json:"role"`
	RoomID string `json:"roomID"`
	WSURL  string `json:"wsUrl"`
}
