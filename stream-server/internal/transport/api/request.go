package api

type CreateRoomRequest struct {
	UserId string `json:"user_id"`
	Role   string `json:"role"`
	Name   string `json:"name"`
}

type JoinRoomRequest struct {
	UserId string `json:"user_id"`
	Role   string `json:"role"`
	Name   string `json:"name"`
	URL    string `json:"url"`
}
