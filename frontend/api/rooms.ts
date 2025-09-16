import apiClient from "./client";

interface CreateRoomDTo{
    userId: string;
    name: string
}

interface JoinRoomDTo{
    userId: string;
    roomId: string;
    role: string;
}

interface Room{
    userId:string;
    name:string;
    role:string;
    roomId:string;
    hostURL:string;
    guestURL:string;
    audienceURL:string;
    createdAt:string;
}

interface JoinRoomResponse {
  status: string;
  userId: string;
  role: string;
  roomId: string;
  wsUrl: string;
}


export async function createRoom(payload: CreateRoomDTo):Promise<Room>{
    console.log("calling server")
    const {data} = await apiClient.post("/rooms",payload)
    return data
}

export async function joinRoom(payload:JoinRoomDTo):Promise<JoinRoomResponse>{
    const {data} = await apiClient.post(`/rooms/${payload.roomId}/join`,payload)
    return data
}