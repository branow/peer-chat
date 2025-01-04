import { PeerConnection } from "./peer-connection.js";
import { PeerChatWebsocket } from "./web-socket.js";

const peerConnectionConfig = {
  iceServers: [
    {
      urls: ['stun:stun0.l.google.com:19302', 'stun:stun2.l.google.com:19302']
    }
  ]
};

const hostname = window.location.hostname;
const port = window.location.port;
const protocol = secured ? 'wss' : 'ws';
const URL = `${protocol}://${hostname}:${port}/ws/room/${room.id}`;

class Page {
  constructor() {
    this.localVideo = document.getElementById('local-video');
    this.remoteVideo = document.getElementById('remote-video');
    this.loaderContainer = document.getElementById('loader-container');
    this.loaderMessage = document.getElementById('loader-message');
    this.disconnectBtn = document.getElementById('disconnect-control');
    this.microBtn = document.getElementById('micro-control');
    this.cameraBtn = document.getElementById('camera-control');
    this.muteMicro = document.getElementById('mute-micro');
    this.muteCamera = document.getElementById('mute-camera');
    this.errorMessageContainer = document.querySelector('.err-message-container');
    this.errorMessageContainer.style.display = "none";
    this.errorMessage = document.querySelector('.err-message');
    
    this.microBtn.addEventListener('click', () => {
      const isOn = this.muteMicro.style.visibility === 'hidden';
      if (isOn) {
        this.turnOffMicrophone();
        this.muteMicro.style.visibility = '';
      } else {
        this.turnOnMicrophone();
        this.muteMicro.style.visibility = 'hidden';
      }
    });
    this.cameraBtn.addEventListener('click', () => {
      const isOn = this.muteCamera.style.visibility === 'hidden';
      if (isOn) {
        this.turnOffCamera();
        this.muteCamera.style.visibility = '';
      } else {
        this.turnOnCamera();
        this.muteCamera.style.visibility = 'hidden';
      }
    });
    this.disconnectBtn.addEventListener('click', () => {
      window.location.href = "/home"
    });
  } 
  setStreams(local, remote) {
    this.localVideo.srcObject = local;
    this.remoteVideo.srcObject = remote;
  }
  setLoading(message) {
    this.loaderContainer.style.visibility = "";
    this.loaderMessage.innerText = message;
  }
  hideLoading() {
    this.loaderContainer.style.visibility = "hidden";
    this.loaderMessage.text = "";
  }
  setError(message) {
    this.errorMessage.innerHTML = message;
    this.errorMessageContainer.style.display = "flex";
  }
  turnOnMicrophone() {}
  turnOffMicrophone() {}
  turnOnCamera() {}
  turnOffCamera() {}
}

const init = async () => {
  const page = new Page();
  page.hideLoading();
  
  // Setup streams calling
  const localStream = await navigator.mediaDevices.getUserMedia({
    video: {
      width: { min: 640, ideal: 1920, max: 1920 },
      height: { min: 480, ideal: 1080, max: 1080 },
    },
    // video: true,
    audio: true
  });
  const remoteStream = new MediaStream();
  page.setStreams(localStream, remoteStream);

  const peerConnection = new PeerConnection(peerConnectionConfig, localStream, remoteStream);  
  
  localStream.getTracks().find(track => track.kind === 'video').enabled = false;
  page.turnOnCamera = () => {
    const video = localStream.getTracks().find(track => track.kind === 'video');
    if (video) video.enabled = true;
  }
  page.turnOffCamera = () => {
    const video = localStream.getTracks().find(track => track.kind === 'video');
    if (video) video.enabled = false;
  }
  page.turnOnMicrophone = () => {
    const audio = localStream.getTracks().find(track => track.kind === 'audio');
    if (audio) audio.enabled = true;
  }
  page.turnOffMicrophone = () => {
    const audio = localStream.getTracks().find(track => track.kind === 'audio');
    if (audio) audio.enabled = false;
  }
  page.microBtn.click();
  page.microBtn.click();
  page.cameraBtn.click();

  const websocket = new PeerChatWebsocket(URL, peerConnection);
  websocket.onerror = () => {
    const message = locale.get("room-error");
    page.setError(message);
  }
  websocket.messageHandlers['wait'] = (event) => { 
    const obj = JSON.parse(event.data)
    switch (obj.data) {
      case 'Wait for peer':
        page.setLoading(locale.get("room-wait-peer"));
        break;
      case 'Wait for room':
        page.setLoading(locale.get("room-wait-room"));
        break;
      default:
        page.setLoading(locale.get("room-wait-unknown"))
        throw new Error(`Unknown reason to wait: ${obj.data}`)
    }
  }
  websocket.messageHandlers['error'] = (event) => {
    console.log(event.data)
  }
  websocket.messageHandlers['offer'] = () => page.hideLoading();
  websocket.messageHandlers['answer'] = () => page.hideLoading();
  websocket.handle();
}

init();