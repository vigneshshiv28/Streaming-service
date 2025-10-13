import { z } from "zod";

export const MessageBase = z.object({
  type: z.string(),
  from: z.string().optional(),
  to: z.string().optional(),
  role: z.string().optional(),
  name: z.string().optional(),
  content: z.string().optional(),
});


export const ParticipantSchema = z.object({
  id: z.string(),
  name: z.string(),
  role: z.string(),
  status: z.string(),
});

export const TrackMetaDataSchema = z.object({
  id: z.string(),
  kind: z.enum(["audio", "video", "screen"]),
  participantId: z.string(),
  participantName: z.string(),
});

export const OutgoingTrackMetaDataSchema = z.object({
  clientTrackId: z.string(),
  trackId: z.string().optional(),
  participantId: z.string(),
  kind: z.enum(["audio", "video", "screen"]),
  participantName: z.string(),
});

export const ChatMessageSchema = MessageBase.extend({
  type: z.literal("chat"),
  content: z.string(),
});
 
export const SDPMessageSchema = MessageBase.extend({
  type: z.literal("sdp"),
  sdp: z.any(), 
  incomingTrackMetaData: z.array(TrackMetaDataSchema).optional(),
  outgoingTrackMetaData: z.array(OutgoingTrackMetaDataSchema).optional(),
});

export const ICEMessageSchema = MessageBase.extend({
  type: z.literal("ice"),
  ice: z.any(), 
});

export const JoinMessageSchema = MessageBase.extend({
  type: z.literal("join"),
});

export const ParicipantLeft = MessageBase.extend({
  type: z.literal("participant_left")
})

export const JoinAckMessageSchema = MessageBase.extend({
  type: z.literal("join_ack"),
  state: z.array(ParticipantSchema),
});

export const ErrorMessageSchema = MessageBase.extend({
  type: z.literal("error"),
  content: z.string(),
});


export const MessageSchema = z.discriminatedUnion("type", [
  ChatMessageSchema,
  SDPMessageSchema,
  ICEMessageSchema,
  JoinMessageSchema,
  JoinAckMessageSchema,
  ErrorMessageSchema,
]);

export type Message = z.infer<typeof MessageSchema>;
export type TrackMetaData = z.infer<typeof TrackMetaDataSchema>;
export type Participant = z.infer<typeof ParticipantSchema>;
export type ChatMessage = z.infer<typeof ChatMessageSchema>;
export type SDPMessage = z.infer<typeof SDPMessageSchema>;
export type ICEMessage = z.infer<typeof ICEMessageSchema>;
export type JoinMessage = z.infer<typeof JoinMessageSchema>;
export type JoinAckMessage = z.infer<typeof JoinAckMessageSchema>;
export type ErrorMessage = z.infer<typeof ErrorMessageSchema>;


