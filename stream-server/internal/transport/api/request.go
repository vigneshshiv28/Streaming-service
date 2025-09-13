package api

type CreateRoomRequest struct {
	UserId string `json:"user_id"`
	Name   string `json:"name"`
}

type JoinRoomRequest struct {
	UserID string `json:"user_id"`
	RoomID string `json:"room_id"`
	Role   string `json:"role"`
}
