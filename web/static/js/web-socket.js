export class PeerChatWebsocket {
  constructor(url, peerConnection) {
    this.websocket = new WebSocket(url);
    this.peerConnection = peerConnection;
    this.messageHandlers = {};
    this.holdInterval = 30 * 1000; // milliseconds
  }

  handle() {
    this.websocket.onopen = (event) => this.call(this.onopen.bind(this), event);
    this.websocket.onclose = (event) =>
      this.call(this.onclose.bind(this), event);
    this.websocket.onmessage = (event) =>
      this.call(this.onmessage.bind(this), event);
    this.websocket.onerror = (event) =>
      this.call(this.onerror.bind(this), event);
  }

  async onmessage(event) {
    const obj = JSON.parse(event.data);
    switch (obj.type) {
      case "request-offer":
        const offer = await this.peerConnection.createOffer();
        this.websocket.send(JSON.stringify(offer));
        this.unhold();
        break;
      case "offer":
        const answer = await this.peerConnection.createAnswer(obj);
        this.websocket.send(JSON.stringify(answer));
        this.hold();
        break;
      case "answer":
        await this.peerConnection.addAnswer(obj);
        this.hold();
        break;
      case "wait":
        await this.peerConnection.wait();
        this.hold();
        break;
      case "error":
        throw new Error(`Server Error:`, obj.data);
        this.unhold();
    }

    if (this.messageHandlers) {
      const handle = this.messageHandlers[obj.type];
      this.call(handle, event);
    }
  }

  hold() {
    if (!this.holdIntervalId) {
      this.holdIntervalId = setInterval(() => {
        this.websocket.send(JSON.stringify(HoldMessage));
      }, this.holdInterval);
    }
  }

  unhold() {
    if (this.holdIntervalId) {
      clearInterval(this.holdIntervalId);
      this.holdIntervalId = 0;
    }
  }

  call(func, event) {
    if (func) {
      func(event);
    }
  }
}

const HoldMessage = {
  type: "hold",
  data: "",
  sdp: "",
};
