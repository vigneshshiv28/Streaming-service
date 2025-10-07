import apiClient from "./client";
import {
  CreateRoomRequestSchema,
  CreateRoomResponseSchema,
  JoinRoomRequestSchema,
  JoinRoomResponseSchema,
  CreateRoomRequestDto,
  CreateRoomResponseDto,
  JoinRoomRequestDto,
  JoinRoomResponseDto,
} from "./schemas";


export async function createRoom(payload: CreateRoomRequestDto): Promise<CreateRoomResponseDto> {

  const validatedPayload = CreateRoomRequestSchema.parse(payload);
  const { data } = await apiClient.post("/rooms", validatedPayload);
  const validatedResponse = CreateRoomResponseSchema.parse(data);

  return validatedResponse;
}


export async function joinRoom(payload: JoinRoomRequestDto): Promise<JoinRoomResponseDto> {

  const validatedPayload = JoinRoomRequestSchema.parse(payload);
  const { data } = await apiClient.post(`/rooms/${payload.roomId}/join`, validatedPayload);
  const validatedResponse = JoinRoomResponseSchema.parse(data);

  return validatedResponse;
}