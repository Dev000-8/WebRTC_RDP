package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"webrtc-rdp/internal/encoders"
	"webrtc-rdp/internal/rdisplay"
	"webrtc-rdp/internal/rtc"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	httpDefaultPort   = "10000"
	defaultStunServer = "stun:stun1.l.google.com:19302" //"stun:stun.l.google.com:19302"
	maxMessageSize    = 1048576
	writeWait         = 10 * time.Second
	pongWait          = 60 * time.Second
	pingPeriod        = (pongWait * 9) / 10
)

type textMessage struct {
	Type string
	Data string
}

type sessionRequest struct {
	Offer  string `json:"Offer"`
	Screen int    `json:"Screen"`
	Peer   string `json:"Peer"`
}

type sessionResponse struct {
	Answer string `json:"Answer"`
	Peer   string `json:"Peer"`
}

type directoryMessage struct {
	Name string `json:"Name"`
	Path string `json:"Path"`
}

type downloadMessage struct {
	Name string `json:"Name"`
	Path string `json:"Path"`
	Peer string `json:"Peer"`
}

func Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func main() {

	//httpPort := flag.String("http.port", httpDefaultPort, "HTTP listen port")
	stunServer := flag.String("stun.server", defaultStunServer, "STUN server URL (stun:)")
	flag.Parse()

	var video rdisplay.Service
	video, err := rdisplay.NewVideoProvider()
	if err != nil {
		log.Fatalf("Can't init video: %v", err)
	}
	_, err = video.Screens()
	if err != nil {
		log.Fatalf("Can't get screens: %v", err)
	}

	var enc encoders.Service = &encoders.EncoderService{}
	if err != nil {
		log.Fatalf("Can't create encoder service: %v", err)
	}

	var webrtc rtc.Service = rtc.NewRemoteScreenService(*stunServer, video, enc)

	errors := make(chan error, 3)

	for {
		url := "127.0.0.1:19890"
		dialer := *websocket.DefaultDialer
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		ws, _, err := dialer.Dial("wss://"+url, nil)
		if err != nil {
			errors <- err
		} else {
			go func() {
				defer func() {
					ws.Close()
				}()
				ws.SetReadLimit(maxMessageSize)
				//c.conn.SetReadDeadline(time.Now().Add(pongWait))
				//c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
				/*name, err := os.Hostname()
				if err != nil {
					panic(err)
				}*/

				var uploadHandle *os.File = nil
				uuid := uuid.New().String()
				initData, err := json.Marshal(textMessage{"init", uuid})
				if err != nil {
					errors <- err
					return
				}
				fmt.Printf("UUID: %s\n", uuid)
				ws.WriteMessage(websocket.TextMessage, initData)
				for {
					messageType, message, err := ws.ReadMessage()
					if err != nil {
						errors <- err
						break
					}

					if messageType == websocket.TextMessage {

						var incomingMessage textMessage
						err := json.Unmarshal(message, &incomingMessage)
						if err == nil {
							switch incomingMessage.Type {
							case "session_request":
								var session sessionRequest

								err := json.Unmarshal([]byte(incomingMessage.Data), &session)
								if err == nil {
									peer, err := webrtc.CreateRemoteScreenConnection(session.Screen, 20)
									if err != nil {
										fmt.Println(err)
									} else {
										answer, err := peer.ProcessOffer(session.Offer)
										if err != nil {
											fmt.Println(err)
										} else {
											payload, err := json.Marshal(sessionResponse{
												Answer: answer,
												Peer:   session.Peer,
											})
											if err != nil {
												fmt.Println(err)
											} else {
												response, err := json.Marshal(textMessage{"session_response", string(payload)})
												if err != nil {
													fmt.Println(err)
													return
												}
												ws.WriteMessage(websocket.TextMessage, response)
											}
										}
									}
								}

								break

							case "file_download":

								var message downloadMessage
								err := json.Unmarshal([]byte(incomingMessage.Data), &message)
								if err != nil {
									fmt.Println(err)
									return
								}

								filePathName := filepath.Join(message.Path, message.Name)
								f, err := os.OpenFile(filePathName, os.O_RDONLY, 0644)
								if err != nil {
									fmt.Print(err)
									break
								}

								response, err := json.Marshal(textMessage{"file_download_stream", message.Peer})
								if err != nil {
									fmt.Println(err)
									return
								}

								err = ws.WriteMessage(websocket.TextMessage, response)
								if err != nil {
									fmt.Println(err)
									return
								}

								time.Sleep(100 * time.Millisecond)

								const BufferSize = 512 * 1024
								magicBuffer := make([]byte, 8)
								magicBuffer[0] = 0x19
								magicBuffer[1] = 0x88
								magicBuffer[2] = 0x01
								magicBuffer[3] = 0x28
								magicBuffer[4] = 0x02
								magicBuffer[5] = 0x42
								magicBuffer[6] = 0xac
								magicBuffer[7] = 0x12

								buffer := make([]byte, BufferSize)
								for {
									bytesread, err := f.Read(buffer)

									if err != nil {
										if err != io.EOF {
											fmt.Println(err)
										}

										break
									}

									mergedArr := append(magicBuffer, buffer[:bytesread]...)
									err = ws.WriteMessage(websocket.BinaryMessage, mergedArr)
									if err != nil {
										fmt.Println(err)
									}
								}

								err = ws.WriteMessage(websocket.BinaryMessage, magicBuffer)
								if err != nil {
									fmt.Println(err)
								}

								/*response, err = json.Marshal(textMessage{"file_download_end", string("")})
								if err != nil {
									fmt.Println(err)
									return
								}

								err = ws.WriteMessage(websocket.TextMessage, response)
								if err != nil {
									fmt.Println(err)
								}*/

								break

							case "file_upload":
								dir, err := os.UserHomeDir()
								downloadPath := filepath.Join(dir, "Downloads")
								res, _ := Exists(downloadPath)
								if !res {
									downloadPath = dir
								}

								if err == nil {
									var msg directoryMessage
									err := json.Unmarshal([]byte(incomingMessage.Data), &msg)
									if err != nil {
										fmt.Println(err)
										return
									}

									if msg.Path != "" {
										downloadPath = msg.Path
									}

									filePathName := filepath.Join(downloadPath, msg.Name)
									duplicateIndex := 1
									withoutExt := strings.TrimSuffix(filePathName, filepath.Ext(filePathName))
									for {

										res, _ := Exists(filePathName)
										if res {
											if filepath.Ext(filePathName) != "" {
												filePathName = withoutExt + "_" + strconv.Itoa(duplicateIndex) + filepath.Ext(filePathName)
											} else {
												filePathName = withoutExt + "_" + strconv.Itoa(duplicateIndex)
											}
											duplicateIndex++
										} else {
											break
										}
									}

									f, err := os.OpenFile(filePathName, os.O_CREATE, 0644)

									if err != nil {
										fmt.Println(err)
										break
									}
									uploadHandle = f
								}

								break

							}

						}
					} else if messageType == websocket.BinaryMessage {
						if uploadHandle != nil && len(message) >= 8 {

							result := message[0] == 0x19 && message[1] == 0x88 && message[2] == 0x01 && message[3] == 0x28 && message[4] == 0xf5 && message[5] == 0x5f && message[6] == 0x68 && message[7] == 0x9c
							if result {
								if len(message) == 8 {

									if uploadHandle != nil {
										uploadHandle.Close()
									}
									uploadHandle = nil

								} else {
									_, err := uploadHandle.Write(message[8:])
									if err != nil {
										log.Print(err)
									}
								}
							}

						}
					}

				}
			}()
		}

		err = <-errors
		log.Println(err)
		time.Sleep(1 * time.Second)
	}

	/*
		go func() {
			interrupt := make(chan os.Signal)
			signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
			errors <- fmt.Errorf("Received %v signal", <-interrupt)
		}()

		mux := http.NewServeMux()

		// Endpoint to create a new speech to text session
		mux.Handle("/api/", http.StripPrefix("/api", api.MakeHandler(webrtc, video)))

		// Serve static assets
		mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./web"))))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			http.ServeFile(w, r, "./web/index.html")
		})
	*/

	/*go func() {
		log.Printf("Server started on port %s", *httpPort)
		errors <- http.ListenAndServeTLS(fmt.Sprintf(":%s", *httpPort), "certs/localhost.crt", "certs/localhost.key", mux)
	}()*/

	err = <-errors
	log.Printf("%s, exiting.", err)
}
