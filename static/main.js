
class PeerConnection {
  constructor(connectionConfig) {
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
}

const connectionConfig = {
  iceServers: [
    {
      urls: ['stun:stun0.l.google.com:19302', 'stun:stun2.l.google.com:19302']
    }
  ]
}

let peerConnection;
const init = async () => {
  peerConnection = new PeerConnection(connectionConfig);  
  await peerConnection.init();
  
  document.getElementById('user-1').srcObject = peerConnection.localStream;
  document.getElementById('user-2').srcObject = peerConnection.remoteStream;
  
  initWebsocket(peerConnection)
}


const initWebsocket = (peerConnection) => {
  websocket = new WebSocket('ws://localhost:8080/ws')
  websocket.onopen = () => {
    console.log('open:', new Date().toISOString())
  }
  websocket.onclose = () => {
    console.log('close:', new Date().toISOString())
  }
  websocket.onmessage = async (event) => {
    console.log('receive: ', event);
    const obj = JSON.parse(event.data)
    switch (obj.type) {
      case 'request-offer':
        const offer = await peerConnection.createOffer();
        websocket.send(JSON.stringify(offer));
        console.log('send offer');
        break;
      case 'offer':
        const answer = await peerConnection.createAnswer(obj);
        websocket.send(JSON.stringify(answer));
        console.log('send answer');
        break;
      case 'answer':
        await peerConnection.addAnswer(obj);
        console.log('add answer');
        break;
      default:
        throw new Error(`Invalid object type ${obj.type}`)
    }
  }
}

const initManual = (peerConnection) => {
  const offerTextArea = document.getElementById('offer');
  const answerTextArea = document.getElementById('answer');

  document.getElementById('create-offer').addEventListener('click', async () => {
    const offer = await peerConnection.createOffer();
    offerTextArea.value = JSON.stringify(offer);
  });

  document.getElementById('create-answer').addEventListener('click', async () => {
    const offer = JSON.parse(offerTextArea.value);
    const answer = await peerConnection.createAnswer(offer);
    answerTextArea.value = JSON.stringify(answer);
  });

  document.getElementById('add-answer').addEventListener('click', async () => {
    const answer = JSON.parse(answerTextArea.value);
    await peerConnection.addAnswer(answer);
  });
}

init();