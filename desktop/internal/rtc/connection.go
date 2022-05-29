package rtc

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"webrtc-rdp/internal/encoders"
	"webrtc-rdp/internal/rdisplay"

	"github.com/go-vgo/robotgo"
	"github.com/google/uuid"
	"github.com/micmonay/keybd_event"
	"github.com/pion/sdp"
	"github.com/pion/webrtc/v2"
	"github.com/shirou/gopsutil/disk"
	"github.com/tidwall/sjson"
)

// RemoteScreenPeerConn is a webrtc.PeerConnection wrapper that implements the
// PeerConnection interface
type RemoteScreenPeerConn struct {
	connection *webrtc.PeerConnection
	stunServer string
	track      *webrtc.Track
	streamer   videoStreamer
	grabber    rdisplay.ScreenGrabber
	encService encoders.Service
}

func findBestCodec(sdp *sdp.SessionDescription, encService encoders.Service, h264Profile string) (*webrtc.RTPCodec, encoders.VideoCodec, error) {
	var h264Codec *webrtc.RTPCodec
	var vp8Codec *webrtc.RTPCodec
	for _, md := range sdp.MediaDescriptions {
		for _, format := range md.MediaName.Formats {
			if format == "webrtc-datachannel" {
				continue
			}

			intPt, _ := strconv.Atoi(format)
			payloadType := uint8(intPt)
			sdpCodec, err := sdp.GetCodecForPayloadType(payloadType)
			if err != nil {
				return nil, encoders.NoCodec, fmt.Errorf("Can't find codec for %d", payloadType)
			}

			if sdpCodec.Name == webrtc.H264 && h264Codec == nil {
				packetSupport := strings.Contains(sdpCodec.Fmtp, "packetization-mode=1")
				supportsProfile := strings.Contains(sdpCodec.Fmtp, fmt.Sprintf("profile-level-id=%s", h264Profile))
				if packetSupport && supportsProfile {
					h264Codec = webrtc.NewRTPH264Codec(payloadType, sdpCodec.ClockRate)
					h264Codec.SDPFmtpLine = sdpCodec.Fmtp
				}
			} else if sdpCodec.Name == webrtc.VP8 && vp8Codec == nil {
				vp8Codec = webrtc.NewRTPVP8Codec(payloadType, sdpCodec.ClockRate)
				vp8Codec.SDPFmtpLine = sdpCodec.Fmtp
			}
		}
	}
	if vp8Codec != nil && encService.Supports(encoders.VP8Codec) {
		return vp8Codec, encoders.VP8Codec, nil
	}

	if h264Codec != nil && encService.Supports(encoders.H264Codec) {
		return h264Codec, encoders.H264Codec, nil
	}

	return nil, encoders.NoCodec, fmt.Errorf("Couldn't find a matching codec")
}

func newRemoteScreenPeerConn(stunServer string, grabber rdisplay.ScreenGrabber, encService encoders.Service) *RemoteScreenPeerConn {
	return &RemoteScreenPeerConn{
		stunServer: stunServer,
		grabber:    grabber,
		encService: encService,
	}
}

func getTrackDirection(sdp *sdp.SessionDescription) webrtc.RTPTransceiverDirection {
	for _, mediaDesc := range sdp.MediaDescriptions {
		if mediaDesc.MediaName.Media == "video" {
			if _, recvOnly := mediaDesc.Attribute("recvonly"); recvOnly {
				return webrtc.RTPTransceiverDirectionRecvonly
			} else if _, sendRecv := mediaDesc.Attribute("sendrecv"); sendRecv {
				return webrtc.RTPTransceiverDirectionSendrecv
			}
		}
	}
	return webrtc.RTPTransceiverDirectionInactive
}

// ProcessOffer handles the SDP offer coming from the client,
// return the SDP answer that must be passed back to stablish the WebRTC
// connection.
func (p *RemoteScreenPeerConn) ProcessOffer(strOffer string) (string, error) {
	sdp := sdp.SessionDescription{}
	err := sdp.Unmarshal(strOffer)
	if err != nil {
		return "", err
	}

	webrtcCodec, encCodec, err := findBestCodec(&sdp, p.encService, "42e01f")
	if err != nil {
		return "", err
	}

	mediaEngine := webrtc.MediaEngine{}
	mediaEngine.RegisterCodec(webrtcCodec)

	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	pcconf := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{p.stunServer},
			},
			/*{
				URLs: []string{"stun:openrelay.metered.ca:80"},
			},*/
			/*{
				URLs:           []string{"turn:openrelay.metered.ca:80"},
				Username:       "openrelayproject",
				Credential:     "openrelayproject",
				CredentialType: webrtc.ICECredentialTypePassword,
			},
			{
				URLs:           []string{"turn:openrelay.metered.ca:443"},
				Username:       "openrelayproject",
				Credential:     "openrelayproject",
				CredentialType: webrtc.ICECredentialTypePassword,
			},*/
			/*{
				URLs:           []string{"turn:openrelay.metered.ca:443?transport=tcp"},
				Username:       "openrelayproject",
				Credential:     "openrelayproject",
				CredentialType: webrtc.ICECredentialTypePassword,
			},*/
		},
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
	}

	peerConn, err := api.NewPeerConnection(pcconf)
	if err != nil {
		return "", err
	}

	p.connection = peerConn

	peerConn.OnICEConnectionStateChange(func(connState webrtc.ICEConnectionState) {
		if connState == webrtc.ICEConnectionStateConnected {
			p.start()
		}

		if connState == webrtc.ICEConnectionStateDisconnected {
			p.Close()
		}

		log.Printf("Connection state: %s \n", connState.String())
	})

	track, err := peerConn.NewTrack(
		webrtcCodec.PayloadType,
		uint32(rand.Int31()),
		uuid.New().String(),
		fmt.Sprintf("remote-screen"),
	)

	direction := getTrackDirection(&sdp)

	if direction == webrtc.RTPTransceiverDirectionSendrecv {
		_, err = peerConn.AddTrack(track)
	} else if direction == webrtc.RTPTransceiverDirectionRecvonly {
		_, err = peerConn.AddTransceiverFromTrack(track, webrtc.RtpTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendonly,
		})
	} else {
		return "", fmt.Errorf("Unsupported transceiver direction")
	}

	offerSdp := webrtc.SessionDescription{
		SDP:  strOffer,
		Type: webrtc.SDPTypeOffer,
	}
	err = peerConn.SetRemoteDescription(offerSdp)
	if err != nil {
		return "", err
	}

	p.track = track

	answer, err := peerConn.CreateAnswer(nil)
	if err != nil {
		return "", err
	}

	screen := p.grabber.Screen()
	sourceSize := image.Point{
		screen.Bounds.Dx(),
		screen.Bounds.Dy(),
	}

	encoder, err := p.encService.NewEncoder(encCodec, sourceSize, p.grabber.Fps())
	if err != nil {
		return "", err
	}

	size, err := encoder.VideoSize()
	if err != nil {
		return "", err
	}

	p.streamer = newRTCStreamer(p.track, &p.grabber, &encoder, size)

	err = peerConn.SetLocalDescription(answer)
	if err != nil {
		return "", err
	}

	directory := ""
	const (
		DRIVE     = 0
		DIRECTORY = 1
		FILE      = 2
	)

	type DirectoryFormat struct {
		Name string
		Type int
	}

	type DiretoryMessage struct {
		Current     string
		Directories []DirectoryFormat
	}

	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		panic(err)
	}

	// Register data channel creation handling
	peerConn.OnDataChannel(func(d *webrtc.DataChannel) {

		if d.Label() == "messages" {
			type WSSMessage struct {
				Command string
				Data    interface{}
			}

			incomingMessage := &WSSMessage{}

			clipboardText := ""
			ticker := time.NewTicker(500 * time.Millisecond)
			done := make(chan bool)
			go func() {

				for {

					select {
					case <-done:
						return
					case t := <-ticker.C:
						text, err := robotgo.ReadAll()
						if err == nil {
							if clipboardText != text {
								clipboardText := text
								response, _ := sjson.Set("", "command", "clipboard")
								response, _ = sjson.Set(response, "data", clipboardText)
								d.SendText(response)
							}

						}
						_ = t
					}

					time.Sleep(300 * time.Millisecond)
				}
			}()

			// Register text message handling
			d.OnMessage(func(msg webrtc.DataChannelMessage) {

				if err = json.Unmarshal(msg.Data, incomingMessage); err != nil {
					log.Fatal(err)
					return
				}

				switch incomingMessage.Command {
				case "color":
					m, ok := incomingMessage.Data.(map[string]interface{})
					if !ok {
						return
					}
					p.streamer.cmMod(m["colorMode"].(string)[0] - '0')

				case "screensize":
					screen := p.grabber.Screen()
					screenSize := make(map[string]interface{})
					screenSize["width"] = screen.Bounds.Dx()
					screenSize["height"] = screen.Bounds.Dy()

					// Create response
					response, _ := sjson.Set("", "command", incomingMessage.Command)
					response, _ = sjson.Set(response, "data", screenSize)
					err := d.SendText(response)
					if err != nil {
						log.Println(err)
					}
					break

				case "directory":
					m, ok := incomingMessage.Data.(map[string]interface{})
					if !ok {
						return
					}

					var directories []DirectoryFormat
					_directory := m["directory"].(string)

					if _directory != "" {
						directory = _directory
					}

					if directory == "" {
						dir, err := os.UserHomeDir()
						if err == nil {
							directory = dir
						}
					}

					r, _ := regexp.Compile(`^\w:\\..\\$`)

					if r.MatchString(directory) {
						partitions, _ := disk.Partitions(false)
						for _, partition := range partitions {

							directoryFormat := DirectoryFormat{partition.Mountpoint, DRIVE}
							directories = append(directories, directoryFormat)
						}
						directory = ""
					} else {
						files, err := ioutil.ReadDir(directory)
						if err != nil {
							log.Println(err)
							break
						}

						for _, f := range files {
							if f.IsDir() {
								directoryFormat := DirectoryFormat{f.Name(), DIRECTORY}
								directories = append(directories, directoryFormat)
							} else {
								directoryFormat := DirectoryFormat{f.Name(), FILE}
								directories = append(directories, directoryFormat)
							}
						}

						r, _ := regexp.Compile(`..\\$`)
						if r.MatchString(directory) {
							directory = filepath.Dir(directory)
							r, _ = regexp.Compile(`\\$`)
							if r.MatchString(directory) {
								directory = directory[0 : len(directory)-1]
							}
						}
					}

					message := DiretoryMessage{directory, directories}
					response, _ := sjson.Set("", "command", "directory")
					response, _ = sjson.Set(response, "data", message)
					err := d.SendText(response)
					if err != nil {
						log.Println(err)
					}

					break

				case "mousemove":
					m, ok := incomingMessage.Data.(map[string]interface{})
					if !ok {
						return
					}

					// string to int
					x, err := strconv.Atoi(m["x"].(string))
					if err != nil {
						return
					}

					y, err := strconv.Atoi(m["y"].(string))
					if err != nil {
						return
					}

					robotgo.MoveMouse(x, y)
					break

				case "mousedown":
					robotgo.MouseToggle("down")
					break
				case "mouseup":
					robotgo.MouseToggle("up")
					break
				case "clipboard":
					m, ok := incomingMessage.Data.(map[string]interface{})
					if !ok {
						return
					}
					clipboardText := m["clipboardText"].(string)

					robotgo.WriteAll(clipboardText)
					break
				case "click":
					m, ok := incomingMessage.Data.(map[string]interface{})
					if !ok {
						return
					}

					robotgo.Click(m["button"].(string), false)
					break

				case "mousedrag":
					m, ok := incomingMessage.Data.(map[string]interface{})
					if !ok {
						return
					}

					x, err := strconv.Atoi(m["x"].(string))
					if err != nil {
						return
					}

					y, err := strconv.Atoi(m["y"].(string))
					if err != nil {
						return
					}
					fmt.Println(m)
					robotgo.Drag(x, y, m["button"].(string))
					break

				case "dblclick":
					m, ok := incomingMessage.Data.(map[string]interface{})
					if !ok {
						return
					}

					robotgo.Click(m["button"].(string), true)
					break

				case "mousescroll":
					m, ok := incomingMessage.Data.(map[string]interface{})
					if !ok {
						return
					}
					robotgo.ScrollMouse(5, m["direction"].(string))
					break

				case "keydown":
					m, ok := incomingMessage.Data.(map[string]interface{})
					if !ok {
						return
					}
					ctrl := m["ctrl"].(bool)
					shift := m["shift"].(bool)

					kb.SetKeys(keybd_event.VK_ENTER)

					if ctrl {
						kb.HasCTRL(true)
					} else if shift {
						kb.HasSHIFT(true)
					}

					err = kb.Launching()
					if err != nil {
						fmt.Println(err)
					}

					//robotgo.KeyTap(config.RobotGoJSKeyMap[m["keyCode"].(float64)])

					break

				case "terminate":
					err := peerConn.Close()
					if err != nil {
						log.Panic(err)
					}
				}
			})
		} else {

		}
	})

	return answer.SDP, nil
}

func (p *RemoteScreenPeerConn) start() {
	p.streamer.start()
}

// Close Stops the video streamer and closes the WebRTC peer connection
func (p *RemoteScreenPeerConn) Close() error {

	if p.streamer != nil {
		p.streamer.close()
	}

	if p.connection != nil {
		return p.connection.Close()
	}
	return nil
}
