package api

type CreateRoomResponse struct {
	UserID      string `json:"userId"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	RoomID      string `json:"roomId"`
	HostURL     string `json:"hostURL"`
	GuestURL    string `json:"guestURL"`
	AudienceURL string `json:"audienceURL"`
	CreatedAt   string `json:"createdAt"`
}

type JoinRoomResponse struct {
	Status string `json:"status"`
	UserID string `json:"userId"`
	Role   string `json:"role"`
	RoomID string `json:"roomId"`
	WSURL  string `json:"wsUrl"`
}
