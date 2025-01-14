export class PeerConnection {
  constructor(connectionConfig, localStream, remoteStream) {
    this.connectionConfig = connectionConfig;
    this.localStream = localStream;
    this.remoteStream = remoteStream;
    this.peerConnection = null;
    this.iceCandidateTimeout = 2 * 1000; // milliseconds
  }

  _createConnection() {
    if (this.peerConnection) {
      this.peerConnection.close();
    }

    this.peerConnection = new RTCPeerConnection(this.connectionConfig);

    this.remoteStream.getTracks().forEach((track) => {
      track.stop();
      this.remoteStream.removeTrack(track);
    });

    this.peerConnection.ontrack = (event) => {
      event.streams[0].getTracks().forEach((track) => {
        this.remoteStream.addTrack(track);
      });
    };

    this.localStream.getTracks().forEach((track) => {
      this.peerConnection.addTrack(track, this.localStream);
    });
  }

  async createOffer() {
    this._createConnection();

    const iceCandidatePromise = this.waitIceCandidates();

    const offer = await this.peerConnection.createOffer();
    await this.peerConnection.setLocalDescription(offer);

    await iceCandidatePromise;

    return this.peerConnection.localDescription;
  }

  async createAnswer(offer) {
    this._createConnection();

    const iceCandidatePromise = this.waitIceCandidates();

    await this.peerConnection.setRemoteDescription(offer);
    const answer = await this.peerConnection.createAnswer();
    await this.peerConnection.setLocalDescription(answer);

    await iceCandidatePromise;

    return this.peerConnection.localDescription;
  }

  async waitIceCandidates() {
    let icIsAdded = false;
    this.peerConnection.onicecandidate = (event) => {
      if (event.candidate) {
        icIsAdded = true;
      }
    };
    return new Promise((resolve) => {
      const intervalId = setInterval(() => {
        if (icIsAdded) {
          resolve();
          clearInterval(intervalId);
        }
      }, this.iceCandidateTimeout);
    });
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
