import { PeerConnectionManager } from "../peerConnectionManager";
import { User, Participant } from "@/types/models";
import { TrackMetaData, useVideoCallStore } from "@/store/videoCallStore";
import { SDPMessage } from "@/schemas";


export const handleSdpMessage = async (
  message: SDPMessage, 
  webSocket: WebSocket, 
  peerManager: PeerConnectionManager,
  user: User,
  participant: Participant,
) => {
  console.log("Processing SDP message:", message.sdp.type);
  console.log("Message MetaData", message.outgoingTrackMetaData) 
  
  const metaData: TrackMetaData[] = (message.outgoingTrackMetaData ?? []).map((t) => ({
    clientId: t.clientTrackId,
    id: t.clientTrackId, // Use clientTrackId instead of trackId for mapping
    kind: t.kind,
    participantId: t.participantId,
    participantName: t.participantName,
  }));
      
     
  if (metaData?.length) {
    useVideoCallStore.getState().updateTracks(metaData);
  }
  
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
