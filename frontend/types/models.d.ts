export interface User {
  id: string;
  name: string;
}

export interface Room {
  roomId: string;
  name: string;
  createdAt: string;
  createdBy: string;
  hostURL: string | null;
  guestURL: string | null;
  audienceURL: string | null;
}



export type Role = "host" | "guest" | "audience";
export type ConnectionStatus = "connecting" | "connected" | "disconnected";

export interface Participant extends User {
  role: Role;
  status: ConnectionStatus;
  wsURL?: string;
}