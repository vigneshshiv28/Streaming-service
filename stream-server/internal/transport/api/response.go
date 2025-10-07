package api

type CreateRoomResponse struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	RoomID      string `json:"roomId"`
	HostURL     string `json:"hostURL"`
	GuestURL    string `json:"guestURL"`
	AudienceURL string `json:"audienceURL"`
	CreatedAt   string `json:"createdAt"`
	CreatedBy   string `json:"createdBy"`
}

type JoinRoomResponse struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	UserID    string `json:"userId"`
	Role      string `json:"role"`
	RoomID    string `json:"roomId"`
	WSURL     string `json:"wsURL"`
	CreatedAt string `json:"createdAt"`
	CreatedBy string `json:"createdBy"`
}
