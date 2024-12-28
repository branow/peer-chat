
export class PeerConnection {
  constructor(connectionConfig){
    this.connectionConfig = connectionConfig;
    this.peerConnection = null;
    this.localStream = null;
    this.remoteStream = null;
  }

  async init() {
    this.localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: false });
    this.remoteStream = new MediaStream();
  }
  
  _createConnection() {
    if (this.peerConnection) { this.peerConnection.close(); }
    
    console.log('create connection')
    
    this.peerConnection = new RTCPeerConnection(this.connectionConfig);

    this.remoteStream.getTracks().forEach((track) => {
      track.stop();
      this.remoteStream.removeTrack(track);
    });

    this.peerConnection.ontrack = (event) => {
      event.streams[0].getTracks().forEach((track) => {
        this.remoteStream.addTrack(track);
      });
    }

    this.localStream.getTracks().forEach((track) => {
      this.peerConnection.addTrack(track, this.localStream);
    });
  }

  async createOffer() {
    this._createConnection();

    const iceCandidatePromise = new Promise((resolve) => {
      this.peerConnection.onicecandidate = (event) => {
        if (event.candidate) {
          resolve();
        }
      }
    });
    
    const offer = await this.peerConnection.createOffer();
    await this.peerConnection.setLocalDescription(offer); 

    await iceCandidatePromise;

    return this.peerConnection.localDescription;
  }

  async createAnswer(offer) {
    this._createConnection();

    const iceCandidatePromise = new Promise((resolve) => {
      this.peerConnection.onicecandidate = (event) => {
        if (event.candidate) {
          resolve();
        }
      }
    });

    await this.peerConnection.setRemoteDescription(offer);
    const answer = await this.peerConnection.createAnswer();
    await this.peerConnection.setLocalDescription(answer);

    await iceCandidatePromise;

    return this.peerConnection.localDescription;
  }

  async addAnswer(answer) {
    if (!this.peerConnection.currentRemoteDescription) {
      await this.peerConnection.setRemoteDescription(answer);
    }
  }

  wait() {
    this.remoteStream.getTracks().forEach((track) => {
      track.stop();
      this.remoteStream.removeTrack(track);
    });
  }
}
