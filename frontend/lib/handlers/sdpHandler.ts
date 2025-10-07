import { PeerConnectionManager } from "../peerConnectionManager";
import { User, Room, Participant } from "@/types/models";
import { Message } from "@/types/message";
export const handleSdpMessage = async (
  message: Message, 
  webSocket: WebSocket, 
  peerManager: PeerConnectionManager,
  user: User,
  participant: Participant,
  room: Room
) => {
  console.log("Processing SDP message:", message.sdp.type);
  console.log("Message MetaData", message.outgoingTrackMetaData)
  if (message.sdp.type === "offer") {

    const answer = await peerManager.createAnswer(message.sdp);
    
    webSocket.send(JSON.stringify({
      type: "sdp",
      sdp: answer,
      from: user.id,
      to: message.from,
      role: participant.role
    }));
    
    console.log("Sent SDP answer");  
  } else if (message.sdp.type === "answer") {

    await peerManager.processRemoteSessionDescription(message.sdp);
    console.log("Applied SDP answer");
  }
};
