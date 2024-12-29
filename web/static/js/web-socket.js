

export class PeerChatWebsocket {
  constructor(url, peerConnection) {
    this.websocket = new WebSocket(url);
    this.peerConnection = peerConnection;
    this.messageHandlers = {};
  }
  
  handle() {
    this.websocket.onopen = (event) => this.call(this.onopen.bind(this), event);
    this.websocket.onclose = (event) => this.call(this.onclose.bind(this), event);
    this.websocket.onmessage = (event) => this.call(this.onmessage.bind(this), event);
  }
  
  async onmessage(event) {
    const obj = JSON.parse(event.data)
    switch (obj.type) {
      case 'request-offer':
        const offer = await this.peerConnection.createOffer();
        this.websocket.send(JSON.stringify(offer));
        break;
      case 'offer':
        const answer = await this.peerConnection.createAnswer(obj);
        this.websocket.send(JSON.stringify(answer));
        break;
      case 'answer':
        await this.peerConnection.addAnswer(obj);
        break;
      case 'wait':
        await this.peerConnection.wait();
        break;
      case 'error':
        throw new Error(`Server Error:`, obj.data);
    }
    
    if (this.messageHandlers) {
      const handle = this.messageHandlers[obj.type];
      this.call(handle, event);
    }
  } 
  
  call(func, event) {
    if (func) {
      func(event);
    }
  }
}
