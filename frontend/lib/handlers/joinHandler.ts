import { PeerConnectionManager } from "../peerConnectionManager";
import { User, Room } from "@/types/models";

export const handleJoinMessage = async (
  message: any, 
  webSocket: WebSocket, 
  peerManager: PeerConnectionManager,
  room: Room,
  user: User
) => {
  console.log("Participant joined:", message.from);
  

  if (room.role === "host" && message.role === "guest") {
    console.log("Host creating offer for guest");
    const offer = await peerManager.createOffer();
    
    webSocket.send(JSON.stringify({
      type: "sdp",
      sdp: offer,
      from: user.id,
      to: message.from,
      role: room.role
    }));
  } else if (room.role === "guest" && message.role === "host") {
    console.log("Guest acknowledged host presence");
  }
};