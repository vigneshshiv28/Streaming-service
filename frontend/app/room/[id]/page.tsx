"use client";
import { useEffect, useRef, useState } from "react";
import { PeerConnectionManager, IncomingStream } from "@/lib/peerConnectionManager";
import { handleSdpMessage } from "@/lib/handlers/sdpHandler";
import { handleIceMessage } from "@/lib/handlers/iceHandler";
import { useUserStore } from "@/store/userStore";
import { useRoomStore } from "@/store/roomStore";
import { useParticipantStore } from "@/store/participantStore";
import { useVideoCallStore } from "@/store/videoCallStore";
import { generateTrackID } from "@/lib/generateIDs";
import { Message, TrackMetaData } from "@/types/message";


const connectionState = {
  isConnecting: false,
  activeConnection: null as WebSocket | null,
};

export default function Room() {
  const localVideoRef = useRef<HTMLVideoElement>(null);
  const [remotePeerStreams, setRemotePeerStreams] = useState<IncomingStream[]>([]);
  const peerManagerRef = useRef<PeerConnectionManager>(new PeerConnectionManager());
  const webSocketRef = useRef<WebSocket>(null);
  
  const user = useUserStore((state) => state.user);
  const room = useRoomStore((state) => state.currentRoom);
  const participant = useParticipantStore((state) => state.currentParticipant);
  const initializeVideoCall = useVideoCallStore((state) => state.initializeVideoCall);
  const setAudioTrack = useVideoCallStore((state) => state.setAudioTrack);
  const setVideoTrack = useVideoCallStore((state) => state.setVideoTrack);
  const setScreenShareTrack = useVideoCallStore((state) => state.setScreenShareTrack);

  useEffect(() => {
    if (!user || !participant || !participant.wsURL || !room) {
      console.log(user, participant);
      console.log("returning");
      return;
    }

    if (connectionState.isConnecting || 
        (connectionState.activeConnection?.readyState === WebSocket.OPEN ||
         connectionState.activeConnection?.readyState === WebSocket.CONNECTING)) {
      console.log("Connection already exists, reusing...");
      webSocketRef.current = connectionState.activeConnection;
      return;
    }

    connectionState.isConnecting = true;
    let isConnected = false;
    const peerManager = peerManagerRef.current;

    const setupAndConnect = async () => {
      const ws = new WebSocket(`${participant.wsURL}`);
      webSocketRef.current = ws;
      connectionState.activeConnection = ws;

      if (!ws) {
        alert("fail to create ws object");
        connectionState.isConnecting = false;
        return;
      }

      ws.onopen = async () => {
        isConnected = true;
        connectionState.isConnecting = false;
        console.log("webSocketRef connected");
        console.log("room", room);

        const joinMessage: Message = {
          type: "join",
          from: user.id,
          role: participant.role,
          name: user.name,
        };

        ws.send(JSON.stringify(joinMessage));
      };

      ws.onclose = () => {
        isConnected = false;
        connectionState.isConnecting = false;
        connectionState.activeConnection = null;
        console.log("webSocketRef disconnected");
      };

      ws.onerror = (error) => {
        connectionState.isConnecting = false;
        console.error("webSocketRef error:", error);
      };

      ws.onmessage = async (event) => {
        const message = JSON.parse(event.data);
        try {
          switch (message.type) {
            case 'join_ack':
              console.log("Join acknowledged by server");
              if (participant.role === "host" || participant.role === "guest") {
                try {
                  const stream = await peerManager.captureLocalMedia();
                  console.log("Generated track:", stream.getTracks());
                  if (localVideoRef.current) {
                    localVideoRef.current.srcObject = stream;
                  }

                  let currentState = useVideoCallStore.getState().videoCall;

                  if (!currentState) {
                    initializeVideoCall();
                    currentState = useVideoCallStore.getState().videoCall;
                  }

                  if (!currentState) {
                    console.error("Failed to initialize video call state");
                    return;
                  }
                } catch (error) {
                  console.error("Failed to capture local media:", error);
                  alert("Unable to access camera/microphone: " + error.message);
                  return;
                }

                const state = useVideoCallStore.getState().videoCall!;

                let audioId = state.audioTrackId;
                let videoId = state.videoTrackId;
                let screenId = state.screenShareId;

                if (!state.muted && !audioId) {
                  audioId = generateTrackID("audio");
                  setAudioTrack(audioId);
                }

                if (state.cameraOn && !videoId) {
                  videoId = generateTrackID("video");
                  setVideoTrack(videoId);
                }

                if (state.screenSharing && !screenId) {
                  screenId = generateTrackID("screen");
                  setScreenShareTrack(screenId);
                }

                console.log("video call state", state);
                console.log("creating offer");

                const offer = await peerManager.createOffer();

                const trackMetaData: TrackMetaData[] = [
                  state.muted === false && audioId
                    ? {
                        id: audioId,
                        kind: "audio",
                        participantId: participant.id,
                        participantName: participant.name,
                      }
                    : null,
                  state.cameraOn === true && videoId
                    ? {
                        id: videoId,
                        kind: "video",
                        participantId: participant.id,
                        participantName: participant.name,
                      }
                    : null,
                  state.screenSharing === true && screenId
                    ? {
                        id: screenId,
                        kind: "screen",
                        participantId: participant.id,
                        participantName: participant.name,
                      }
                    : null,
                ].filter((t): t is TrackMetaData => t !== null);

                console.log("trackMetaData", trackMetaData);
                const sdpOfferMessage: Message = {
                  type: "sdp",
                  sdp: offer,
                  from: user.id,
                  role: participant.role,
                  incomingTrackMetaData: trackMetaData,
                };
                ws.send(JSON.stringify(sdpOfferMessage));
              }

              peerManager.onRemoteStreamReceived((incomingStream) => {
                setRemotePeerStreams((prev) =>
                  prev.find((s) => s.stream.id === incomingStream.stream.id)
                    ? prev
                    : [...prev, incomingStream]
                );
              });
              break;

            case 'join':
              break;

            case 'sdp':
              await handleSdpMessage(message, ws, peerManager, user, participant, room);
              break;

            case 'ice':
              await handleIceMessage(message, ws, peerManager);
              break;

            case 'participant_left':
              const leftPeerId = JSON.parse(message.content).participant_id;
              setRemotePeerStreams((prev) => prev.filter((p) => p.peerId !== leftPeerId));
              break;
            default:
              console.warn("Unknown message type:", message.type);
          }
        } catch (error) {
          console.error("Error handling message:", error);
        }
      };
    };

    setupAndConnect();

    return () => {
      if (isConnected && webSocketRef.current) {
        webSocketRef.current.close();
        connectionState.activeConnection = null;
      }
      peerManager.cleanup();
    };
  }, [participant?.wsURL, room?.roomId, user?.id]);

  if (!room) {
    return (
      <div>
        <h2>Loading room...</h2>
      </div>
    );
  }

  return (
    <div style={{ padding: 20 }}>
      <h1>Video Room: {room.name || room.roomId}</h1>

      <div style={{ marginBottom: 20 }}>
        <h2>My Video</h2>
        <video
          ref={localVideoRef}
          autoPlay
          playsInline
          muted
          style={{ width: 300, height: 200, border: '2px solid #333', borderRadius: 8 }}
        />
      </div>

      <div>
        <h2>Remote Participants ({remotePeerStreams.length})</h2>
        <div style={{ display: "flex", gap: 15, flexWrap: "wrap" }}>
          {remotePeerStreams.map((remotePeer) => (
            <div key={remotePeer.peerId} style={{ textAlign: 'center' }}>
              <video
                autoPlay
                playsInline
                ref={(video) => {
                  if (video) video.srcObject = remotePeer.stream;
                }}
                style={{ width: 300, height: 200, border: '2px solid #666', borderRadius: 8 }}
              />
              <p style={{ margin: '8px 0 0 0', fontSize: 14, color: '#666' }}>
                {remotePeer.peerId}
              </p>
            </div>
          ))}
          {remotePeerStreams.length === 0 && (
            <p style={{ color: '#888', fontStyle: 'italic' }}>
              Waiting for other participants to join...
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
