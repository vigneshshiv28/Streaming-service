import { OutgoingMessage } from "http";
import { z } from "zod";

export const TrackMetaDataSchema = z.object({
  id: z.string(), 
  kind: z.enum(["audio", "video", "screen"]), 
  participantId: z.string(),
  participantName: z.string(),
});

export interface OutgoingTrackMetaData {
  clientTrackId: string
  trackId: string;
  participantId: string;
  kind: "audio" | "video" | "screen";
  participantName: string;
}

export const MessageSchema = z.object({
  type: z.string(),
  from: z.string().optional(),
  to: z.string().optional(),
  role: z.string().optional(),
  name: z.string().optional(),
  content: z.string().optional(),
  sdp: z.any().optional(),
  ice: z.any().optional(),
  action: z.string().optional(),
  incomingTrackMetaData: z.array(TrackMetaDataSchema).optional(),
  outgoingTrackMetaData: z.array(OutgoingMessage).optional() 
});

export type TrackMetaData = z.infer<typeof TrackMetaDataSchema>;
export type Message = z.infer<typeof MessageSchema>;