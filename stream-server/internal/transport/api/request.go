package api

type CreateRoomRequest struct {
	UserId   string `json:"userId"`
	UserName string `json:"userName"`
	Name     string `json:"name"`
}

type JoinRoomRequest struct {
	UserID string `json:"userId"`
	RoomID string `json:"roomId"`
	Role   string `json:"role"`
}
