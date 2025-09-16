"use client";
import { useEffect, useRef, useState } from "react";
import { PeerConnectionManager, IncomingStream } from "@/lib/peerConnectionManager";
import { handleJoinMessage } from "@/lib/handlers/joinHandler";
import { handleSdpMessage } from "@/lib/handlers/sdpHandler";
import { handleIceMessage } from "@/lib/handlers/iceHandler";
import { useUserStore } from "@/store/userStore";
import { useRoomStore } from "@/store/roomStore";

export default function Room() {
  const localVideoRef = useRef<HTMLVideoElement>(null);
  const [remotePeerStreams, setRemotePeerStreams] = useState<IncomingStream[]>([]);
  const remotePeerIdRef = useRef<string | null>(null);

  
  const user = useUserStore((state) => state.user);
  const room = useRoomStore((state) => state.room);

  useEffect(() => {
    if (!user || !room?.wsURL ) return;

    const peerManager = new PeerConnectionManager();
    const webSocket = new WebSocket(`${room.wsURL}`);

    webSocket.onopen = () => {
      console.log("WebSocket connected");
      console.log("Sending join message");
      
      webSocket.send(JSON.stringify({
        type: "join",
        from: user.id,
        role: room.role,
        name: user.name,
      }));
    };

    webSocket.onclose = () => {
      console.log("WebSocket disconnected");
    };

    webSocket.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    peerManager.captureLocalMedia().then((stream) => {
      if (localVideoRef.current) {
        localVideoRef.current.srcObject = stream;
      }
    }).catch((error) => {
      console.error("Failed to capture local media:", error);
      alert("Unable to access camera/microphone: " + error.message);
    });

  
    peerManager.onRemoteStreamReceived((incomingStream) => {
      console.log("Received remote stream:", incomingStream.peerId);
      
      setRemotePeerStreams((previousStreams) => {
    
        if (previousStreams.find((s) => s.peerId === incomingStream.peerId)) {
          return previousStreams;
        }
        return [...previousStreams, incomingStream];
      });
    });

   
    webSocket.onmessage = async (event) => {
      const message = JSON.parse(event.data);
      console.log("Received message:", message);

      try {
        switch (message.type) {
          case 'join':
            if (room.role === "host" && message.role === "guest") {
                remotePeerIdRef.current = message.from ?? null;

            } else if (room.role === "guest" && message.role === "host") {
                remotePeerIdRef.current = message.from ?? null;
            }
            await handleJoinMessage(message, webSocket, peerManager, room, user);
            break;
            
          case 'sdp':
            if (!remotePeerIdRef.current && message.from) {
                remotePeerIdRef.current = message.from;
            }
            await handleSdpMessage(message, webSocket, peerManager, user, room);
            break;
            
          case 'ice':
            await handleIceMessage(message, webSocket, peerManager);
            break;
            
          case 'participant_left':
            console.log("SFU message received (not implemented yet):", message.type);
            break;
            
          default:
            console.warn("Unknown message type:", message.type);
        }
      } catch (error) {
        console.error("Error handling message:", error);
      }
    };
    peerManager.getPeerConnectionInstance().onicecandidate = (event) => {
      if (event.candidate) {
        console.log("Sending ICE candidate");
        
        webSocket.send(JSON.stringify({
          type: "ice",
          from: user.id,
          role: room.role,
          to: remotePeerIdRef.current,
          ice: event.candidate,
        }));
      }
    };

 
    return () => {
      console.log("Cleaning up WebSocket and PeerConnection");
      webSocket.close();
      peerManager.cleanup();
    };
  }, [room, user]);

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
          style={{ 
            width: 300, 
            height: 200, 
            border: '2px solid #333',
            borderRadius: 8 
          }} 
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
                style={{ 
                  width: 300, 
                  height: 200,
                  border: '2px solid #666',
                  borderRadius: 8 
                }}
              />
              <p style={{ margin: '8px 0 0 0', fontSize: 14, color: '#666' }}>
                {remotePeer.peerId} ({remotePeer.type})
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