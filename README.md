## WebRTC Remote Desktop

GO based WebRTC Remote Desktop allows you to control the computers remotely using WebRTC from modern browsers. 
<hr>

### Insipiration

The project is inspired from "WebRTC remote screen (https://github.com/rviscarra/webrtc-remote-screen)".

<hr>

### Features

* Remote screen viewing
* Mouse controls
* Keyboard
* Clipboard
* File download and upload
* Control screen bright (to reduce network load)

<hr>

### Dependencies

- Go 1.12+ (https://golang.org/doc/install)
- If you want h264 support: libx264 (included in x264-go, you'll need a C compiler / assembler to build it) (Supports Mac only)
- If you want VP8 support: libvpx (Supports Windows & Mac)

For Windows libvpx installation, follow below installation steps:
```
1. Download and install latest MYSYS2 installer from https://www.msys2.org/
2. Open MYSYS Shell
3. Install Mingw Toolchain
- 32 bit:  pacman -S mingw-w64-i686-toolchain 
- 64 bit:  pacman -S mingw-w64-x86_64-toolchain
4. pacman -S mingw-w64-x86_64-libvpx
```

<hr>

### Running in development mode

```
go mod tidy
go run -tags "h264enc" cmd/agent.go
```

<hr>

### Building for production
Build the _deployment_ package by runnning `make`. This should create a tar file with the 
binary and web directory, by default only support for h264 is included, if you want to use VP8 run `make encoders=vp8`, if you want both then `make encoders=vp8,h264`.

Copy the archive to a remote server, decompress it and run `./agent`. The `agent` application assumes the web dir. is in the same directory. 

WebRTC requires a _secure_ domain to work, the recommended approach towards this is to forward the agent port thru SSH tunneling:

```bash
ssh -L YOUR_LOCAL_PORT:localhost:xxxx 
```

### Gateway

Gateway is a websocket server which contains signaling function and fontend webpage allows you to control remote computer.

Server address and port is hard coded on the source code.

```
    npm start 
```
<hr>
