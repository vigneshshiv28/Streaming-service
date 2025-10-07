import { PeerConnectionManager } from "../peerConnectionManager";

export const handleIceMessage = async (
  message: any, 
  webSocket: WebSocket, 
  peerManager: PeerConnectionManager
) => {
  console.log("Adding ICE candidate");
  console.log("ICE message",message)
  await peerManager.addRemoteIceCandidate(message.ice);
};