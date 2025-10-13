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
import { MessageSchema, TrackMetaDataSchema } from "@/schemas"; 
import { Participant, Role, ConnectionStatus } from "@/types/models";
import {z} from "zod";


const connectionState = {
  isConnecting: false,
  activeConnection: null as WebSocket | null,
};

export default function Room() {
  const user = useUserStore((state) => state.user);
  const room = useRoomStore((state) => state.currentRoom);
  const participant = useParticipantStore((state) => state.currentParticipant);
  const localVideoRef = useRef<HTMLVideoElement>(null);
  const [remotePeerStreams, setRemotePeerStreams] = useState<IncomingStream[]>([]);
  const peerManagerRef = useRef<PeerConnectionManager>(new PeerConnectionManager(user!));
  const webSocketRef = useRef<WebSocket>(null);
  

  const initializeVideoCall = useVideoCallStore((state) => state.initializeVideoCall);
  const setAudioTrack = useVideoCallStore((state) => state.setAudioTrack);
  const setVideoTrack = useVideoCallStore((state) => state.setVideoTrack);
  const setScreenShareTrack = useVideoCallStore((state) => state.setScreenShareTrack);
  const videoCall = useVideoCallStore((state) => state.videoCall)
  const addParticipant = useVideoCallStore((state) => state.addParticipant)
  const addParticipants = useVideoCallStore((state) => state.addParticipants)

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

        const joinMessage = {
          type: "join",
          from: user.id,
          role: participant.role,
          name: user.name,
        };

        const joinMessageResult = MessageSchema.safeParse(joinMessage);

        
        if (!joinMessageResult.success) {
          console.error("Invalid join message schema:");
          return;
        }

        ws.send(JSON.stringify(joinMessageResult.data));
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

        let parseData
        try{
          parseData = JSON.parse(event.data)
        } catch(error){
          console.log("err",error)
        }
        console.log(parseData)
        const messageResult = MessageSchema.safeParse(parseData)

        if (!messageResult.success){
          console.log("Received invalid message schema",messageResult.error)
          return
        }

        const message = messageResult.data
        try {
          switch (message.type) {
            case 'join_ack':
              console.log("Join acknowledged by server");


              peerManager.onRemoteStreamReceived((incomingStream) => {
                setRemotePeerStreams(prev => {
                  const newArr = prev.find(s => s.stream.id === incomingStream.stream.id) ? prev : [...prev, incomingStream];
                  return newArr;
                });
              });
            
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

                const participants = message.state
                .filter(p => p.id !== participant.id)
                .map(p => ({
                  id: p.id,
                  name: p.name,
                  role: p.role as Role,
                  status: p.status as ConnectionStatus,
                }));
                console.log("Participants:", participants)
                addParticipants(participants);

                const offer = await peerManager.createOffer();

                const trackMetaData = [
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
                ].filter((t)=> t !== null);
                const trackMetaDataResult = z.array(TrackMetaDataSchema).safeParse(trackMetaData);

                if (!trackMetaDataResult.success){
                  console.log("Fail to parse MetaData")
                  return 
                }

                console.log("trackMetaData", trackMetaDataResult.data)  ;
                const sdpOfferMessage = {
                  type: "sdp",
                  sdp: offer,
                  from: user.id,
                  role: participant.role,
                  incomingTrackMetaData: trackMetaDataResult.data,
                };

                const sdpOfferMessageRes = MessageSchema.safeParse(sdpOfferMessage)
                if(!sdpOfferMessageRes.success){
                  console.log("Invalid sdp message ")
                }

                ws.send(JSON.stringify(sdpOfferMessageRes.data));
              }
              break;

            case 'join':
              if (message.from && message.role && message.name){
                const newParticipant:Participant = {
                  id: message.from,
                  name: message.name,
                  role: message.role as Role,
                  status: "connected"
                }
                addParticipant(newParticipant)
              } else{
                console.log("empty join message")
              }
              
              break;

            case 'sdp':
              await handleSdpMessage(message, ws, peerManager, user, participant);
              break;

            case 'ice':
              await handleIceMessage(message, ws, peerManager);
              break;
            /*
            case 'participant_left':
              if (message.content == null ){
                console.log("message missing content");
                break;
              }
              const leftPeerId = JSON.parse(message .content).participant_id;
              setRemotePeerStreams((prev) => prev.filter((p) => p.peerId !== leftPeerId));
              break;
            */
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
 
        {videoCall?.participants.map((p) => {
 
          const stream = new MediaStream(
            remotePeerStreams
              .filter((s) => s.peerId === p.id)
              .flatMap((s) => s.stream.getTracks())
          );

          return (
            <div key={p.id} style={{ display: "flex", flexDirection: "column", alignItems: "center" }}>
              <video
                ref={(el) => { if (el) el.srcObject = stream; }}
                autoPlay
                playsInline
                style={{ width: 200, height: 150, border: "2px solid #333", borderRadius: 8 }}
              />
              <span>{p.name}</span>
            </div>
          );
        })}

      </div>
  );
}
