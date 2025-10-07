import { create } from "domain";
import { z } from "zod";

export const CreateRoomRequestSchema = z.object({
  userId: z.string(),
  username: z.string(),
  name: z.string(), 
});

export const CreateRoomResponseSchema = z.object({
  name: z.string(),
  role: z.enum(["host", "guest", "audience"]),
  roomId: z.string(),
  hostURL: z.string(),
  guestURL: z.string(),
  audienceURL: z.string(),
  createdAt: z.string(),
  createdBy: z.string(),
});


export const JoinRoomRequestSchema = z.object({
  userId: z.string(),
  roomId: z.string(),
  role: z.enum(["host", "guest", "audience"]),
});

export const JoinRoomResponseSchema = z.object({
  userId: z.string(),
  name: z.string(),
  role: z.enum(["host", "guest", "audience"]),
  roomId: z.string(),
  wsURL: z.string(),
  status: z.enum(["connecting","connected","disconnected"]),
  createdAt: z.string(),
  createdBy: z.string(),
});


export type CreateRoomRequestDto = z.infer<typeof CreateRoomRequestSchema>;
export type CreateRoomResponseDto = z.infer<typeof CreateRoomResponseSchema>;
export type JoinRoomRequestDto = z.infer<typeof JoinRoomRequestSchema>;
export type JoinRoomResponseDto = z.infer<typeof JoinRoomResponseSchema>;