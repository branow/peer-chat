# Free Peer-to-Peer Internet Call Website

This project is a simple website for free peer-to-peer internet call communication. It allows users to easily create or join rooms for real-time video calls with a maximum of 2 participants per room. The service requires no registration, and you can meet anyone with just a single click.

## Demo

Check out the live demo here: [Website Demo](https://peerchat.top/)

<p align="center">
  <img alt="Using Run function" src="https://raw.githubusercontent.com/branow/peer-chat/main/screenshot.png">
</p>

## Features

- Create public or private rooms.
- Join existing rooms without any registration.
- Peer-to-peer communication (max 2 participants per room).
- Real-time communication using WebRTC.

## Technologies Used

- **Go**: Backend server.
- **JavaScript**: Client-side logic.
- **HTMX**: For dynamic content loading.
- **Websockets**: Real-time bidirectional communication.
- **RTCPeerConnection**: For establishing peer-to-peer connections.

## Getting Started

1. Clone the repository.
2. Install dependencies. 
- ```go mod tidy```
3. Run the server and access the website through your browser (localhost:8080).
* ```go build ```
* ```peer-chat.exe -s=false``` // windows
* ```./peer-chat -s=false``` // linux


## License

This project is licensed under the terms of the [MIT license](https://github.com/branow/peer-chat/blob/main/LICENSE).