package api

type CreateRoomResponse struct {
	Name   string `json:"name"`
	RoomID string `json:"roomID"`
}

type JoinRoomResponse struct {
	Role   string `json:"role"`
	Status string `json:"status"`
	UserID string `json:"userID"`
}
