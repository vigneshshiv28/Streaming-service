import { User } from "@/types/models";


export interface IncomingStream {
  peerId: string;
  stream: MediaStream;
  type: "camera" | "mic" | "screen";
}


export class PeerConnectionManager {
  private peerConnection?: RTCPeerConnection;
  private localStream?: MediaStream;
  private onIncomingStream?: (stream: IncomingStream) => void;
  private queuedIceCandidates: RTCIceCandidate[] = [];
  private user: User

  constructor(user: User) {
    this.user = user
    this.createPeerConnection();
  }

  private createPeerConnection() {
    if (this.peerConnection) {
      console.log("Closing old PeerConnection");
      this.peerConnection.close();
    }

    this.peerConnection = new RTCPeerConnection({
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    });

    this.peerConnection.ontrack = (event) => {
      const streams = event.streams;
    
      streams.forEach((stream) => {
    
        const [peerId, type] = stream.id.split("_");
    
        stream.getTracks().forEach((t) => {

    
          const incomingStream: IncomingStream = {
            peerId,            
            stream,            
            type: (type as "camera" | "mic" | "screen") || (t.kind as "camera" | "mic") , 
          };
    
          if (this.onIncomingStream) {
            this.onIncomingStream(incomingStream);
          }
    
        });
      });
    };

    this.peerConnection.onicecandidate = (event) => {
      if (event.candidate) {
        console.log("ICE candidate ready to send", event.candidate);
      }
    };
  }

  async captureLocalMedia(): Promise<MediaStream> {
    this.localStream = await navigator.mediaDevices.getUserMedia({
      video: true,
      audio: true,
    });

    Object.defineProperty(this.localStream, "id",{value: `${this.user.id}-camera`})

    this.localStream.getTracks().forEach(track => {
      Object.defineProperty(track, "id", {value: 'cameraTrack'})
    });

    if (!this.peerConnection) this.createPeerConnection();

    this.localStream.getTracks().forEach((track) => {
      this.peerConnection!.addTrack(track, this.localStream!);
    });

    return this.localStream;
  }

  getLocalStream() {
    return this.localStream;
  }

  async createOffer(): Promise<RTCSessionDescriptionInit> {
    if (!this.peerConnection) this.createPeerConnection();
    const offer = await this.peerConnection!.createOffer();

    let customizeSDP = offer.sdp

    customizeSDP = customizeSDP!
    .replace(/(m=audio[\s\S]*?a=msid:)([^\s]+) ([^\s]+)\r?\n/, `$1${this.user.id}_mic micTrack\n`)
    .replace(/(m=video[\s\S]*?a=msid:)([^\s]+) ([^\s]+)\r?\n/, `$1${this.user.id}_camera cameraTrack\n`);

    await this.peerConnection!.setLocalDescription({type: offer.type, sdp: customizeSDP});
    return {type: offer.type, sdp: customizeSDP};
  }

  async createAnswer(offer: RTCSessionDescriptionInit): Promise<RTCSessionDescriptionInit> {
    if (!this.peerConnection) this.createPeerConnection();
    await this.peerConnection!.setRemoteDescription(offer);
    await this.processQueuedIceCandidates();
    const answer = await this.peerConnection!.createAnswer();
    await this.peerConnection!.setLocalDescription(answer);
    return answer;
  }

  async processRemoteSessionDescription(sdp: RTCSessionDescriptionInit) {
    if (!this.peerConnection) this.createPeerConnection();
    await this.peerConnection!.setRemoteDescription(sdp);
    await this.processQueuedIceCandidates();
  }

  async addRemoteIceCandidate(ice: RTCIceCandidateInit) {
    if (!this.peerConnection) return;
    
    if (this.peerConnection.remoteDescription === null) {
      this.queueIceCandidate(ice);
      return;
    }

    await this.peerConnection.addIceCandidate(new RTCIceCandidate(ice));
  }

  queueIceCandidate(candidate: RTCIceCandidateInit) {
    this.queuedIceCandidates.push(new RTCIceCandidate(candidate));
  }

  private async processQueuedIceCandidates() {
    for (const candidate of this.queuedIceCandidates) {
      try {
        await this.peerConnection!.addIceCandidate(candidate);
      } catch (error) {
        console.error("Failed to add queued ICE candidate:", error);
      }
    }
    this.queuedIceCandidates = [];
  }

  onRemoteStreamReceived(callback: (stream: IncomingStream) => void) {
    this.onIncomingStream = callback;
  }

  getPeerConnectionInstance(): RTCPeerConnection | undefined {
    return this.peerConnection;
  }

  cleanup() {
    if (this.peerConnection) {
      this.peerConnection.close();
      this.peerConnection = undefined;
    }
    this.localStream?.getTracks().forEach((t) => t.stop());
    this.localStream = undefined;
    this.queuedIceCandidates = [];
  }
} 