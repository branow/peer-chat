import { PeerConnection } from "./peer-connection.js";
import { PeerChatWebsocket } from "./web-socket.js";

const peerConnectionConfig = {
  iceServers: [
    {
      urls: ['stun:stun0.l.google.com:19302', 'stun:stun2.l.google.com:19302']
    }
  ]
}

const init = async () => {
  const peerConnection = new PeerConnection(peerConnectionConfig);  
  await peerConnection.init();
  
  document.getElementById('user-1').srcObject = peerConnection.localStream;
  document.getElementById('user-2').srcObject = peerConnection.remoteStream;
  
  initWebsocket(peerConnection)
}

const initWebsocket = (peerConnection) => {
  const websocket = new PeerChatWebsocket(peerConnection);
  websocket.handle();
  websocket.onopen = () => {
    console.log('open:', new Date().toISOString())
  }
  websocket.onclose = () => {
    console.log('close:', new Date().toISOString())
  }
  websocket.messageHandlers['request-offer'] = () => console.log('send offer')
  websocket.messageHandlers['offer'] = () => console.log('send answer')
  websocket.messageHandlers['request-answer'] = () => console.log('add answer')
  websocket.messageHandlers['wait'] = (event) => { 
    const obj = JSON.parse(event.data)
    switch (obj.data) {
      case 'Wait for peer':
        console.log('Please, wait for peer to connect');
        break;
      case 'Wait for room':
        console.log('Please, wait the room is full');
        break;
      default:
        consol.log('Unknown reason:', obj.data);
    }
  }
}

init();