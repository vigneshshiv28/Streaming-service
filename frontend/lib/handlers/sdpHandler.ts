import { PeerConnectionManager } from "../peerConnectionManager";
import { User, Room } from "@/types/models";

export const handleSdpMessage = async (
  message: any, 
  webSocket: WebSocket, 
  peerManager: PeerConnectionManager,
  user: User,
  room: Room
) => {
  console.log("Processing SDP message:", message.sdp.type);
  
  if (message.sdp.type === "offer") {

    const answer = await peerManager.createAnswer(message.sdp);
    
    webSocket.send(JSON.stringify({
      type: "sdp",
      sdp: answer,
      from: user.id,
      to: message.from,
      role: room.role
    }));
    
    console.log("Sent SDP answer");
  } else if (message.sdp.type === "answer") {

    await peerManager.processRemoteSessionDescription(message.sdp);
    console.log("Applied SDP answer");
  }
};
