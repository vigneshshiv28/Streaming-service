export interface IncomingStream {
  peerId: string;
  stream: MediaStream;
  type: "camera" | "screen";
}

export class PeerConnectionManager {
  private peerConnection?: RTCPeerConnection;
  private localStream?: MediaStream;
  private onIncomingStream?: (stream: IncomingStream) => void;

   constructor() {
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
      if (this.onIncomingStream) {
        const incomingStream: IncomingStream = {
          peerId: "remote-peer", // TODO: set actual peer ID from signaling
          stream: event.streams[0],
          type: "camera",
        };
        this.onIncomingStream(incomingStream);
      }
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
    await this.peerConnection!.setLocalDescription(offer);
    return offer;
  }

  async createAnswer(offer: RTCSessionDescriptionInit): Promise<RTCSessionDescriptionInit> {
    if (!this.peerConnection) this.createPeerConnection();
    await this.peerConnection!.setRemoteDescription(offer);
    const answer = await this.peerConnection!.createAnswer();
    await this.peerConnection!.setLocalDescription(answer);
    return answer;
  }

  async processRemoteSessionDescription(sdp: RTCSessionDescriptionInit) {
    if (!this.peerConnection) this.createPeerConnection();
    await this.peerConnection!.setRemoteDescription(sdp);
  }

  async addRemoteIceCandidate(ice: RTCIceCandidateInit) {
    if (!this.peerConnection) return;
    await this.peerConnection.addIceCandidate(new RTCIceCandidate(ice));
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
  }
}