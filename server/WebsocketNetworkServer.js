"use strict";

var WebsocketNetworkServer = (function () {
    function WebsocketNetworkServer() {
        this.sockets = {};
        this.peerIDs = {};
        this.attributes = {};
        this.curId = 0;
    }
    WebsocketNetworkServer.prototype.onConnection = function (socket) {
        this.sockets[this.curId] = socket;
        this.peerIDs[this.curId] = "";
        this.attributes[this.curId] = "server";
        var thisSocketId = this.curId;
        this.curId += 1;
        var _this = this;
        var fileUploadTarget = null;
        var fileDownloadTarget = null;

        console.log("connected " + socket._socket.remoteAddress);
        
        socket.on('message', function (message, flags) {

            if(typeof message === "string")
            {
                try
                {
                    var obj = JSON.parse(message);
                    if(obj.Type == "init")
                    {
                        _this.peerIDs[thisSocketId] = obj.Data;
                        if(obj.Attribute == "client")
                            _this.attributes[thisSocketId] = "client";
                    }
                    else if(obj.Type == "session_response")
                    {
                        var response = JSON.parse(obj.Data);
                        for (var otherSocket in _this.sockets)
                        {
                            if(_this.peerIDs[otherSocket] == response.Peer)
                            {
                                var data = {
                                    Type: "session_response",
                                    Data: response.Answer
                                };
                                
                                _this.sockets[otherSocket].send(JSON.stringify(data));
                                break;
                            }
                            
                        }
                    }
                    else if(obj.Type == "session_request")
                    {
                        for (var otherSocket in _this.sockets)
                        {
                            if(_this.peerIDs[otherSocket] == obj.Data.peer)
                            {
                                var data = {
                                    Type: "session_request",
                                    Data: JSON.stringify({
                                        Offer: obj.Data.offer,
                                        Screen: obj.Data.screen,
                                        Peer: _this.peerIDs[thisSocketId]
                                    })
                                };
                                
                                _this.sockets[otherSocket].send(JSON.stringify(data));
                                break;
                            }
                            
                        } 
                    }
                    else if(obj.Type == "peers")
                    {
                        let peers = [];
                        for (var otherSocket in _this.sockets)
                        {
                            if (_this.sockets.hasOwnProperty(otherSocket) && _this.sockets[otherSocket] != socket && _this.attributes[otherSocket] == "server")
                            {
                                peers.push(_this.peerIDs[otherSocket]);
                            }
                        }
                        
                        try
                        {
                            var data = {
                                Type: "peers",
                                Data: peers
                            }
                            _this.sockets[thisSocketId].send(JSON.stringify(data));
                        }
                        catch (err) {
                            console.warn("Error in sending message " + message + ":\n" + err);
                        }
                    }
                    else if(obj.Type == "file_upload")
                    {
                        for (var otherSocket in _this.sockets)
                        {
                            if(_this.peerIDs[otherSocket] == obj.Peer)
                            {
                                var data = {
                                    Type: "file_upload",
                                    Data: obj.Data
                                };
                                
                                _this.sockets[otherSocket].send(JSON.stringify(data));
                                fileUploadTarget = _this.sockets[otherSocket];
                                break;
                            }
                        }
                    }
                    else if(obj.Type == "file_download")
                    {
                        for (var otherSocket in _this.sockets)
                        {
                            if(_this.peerIDs[otherSocket] == obj.Peer)
                            {
                                var download_info = {
                                    Name: obj.Data.Name,
                                    Path: obj.Data.Path,
                                    Peer: _this.peerIDs[thisSocketId]
                                }

                                var data = {
                                    Type: "file_download",
                                    Data: JSON.stringify(download_info),
                                };
                                
                                _this.sockets[otherSocket].send(JSON.stringify(data));
                                break;
                            }
                        }
                    } 
                    else if(obj.Type == "file_download_stream")
                    {
                        for (var otherSocket in _this.sockets)
                        {
                            if(_this.peerIDs[otherSocket] == obj.Data)
                            {
                                var data = {
                                    Type: "file_download_stream",
                                    Data: "",
                                };
                                
                                _this.sockets[otherSocket].send(JSON.stringify(data));
                                fileDownloadTarget = _this.sockets[otherSocket];
                                break;
                            }
                        }
                    } 
                }
                catch(err)
                {
                    console.warn("json parse error " + message + ":\n" + err);
                }
            }
            
            if(typeof message === "object")
            {
               if(fileUploadTarget != null)
               {
                    let byteMessage = new Uint8Array(message);
                    if(byteMessage.length >= 8)
                    {
                        let result = byteMessage[0] == 0x19 && byteMessage[1] == 0x88 && byteMessage[2] == 0x01 && byteMessage[3] == 0x28
                            && byteMessage[4] == 0xf5 && byteMessage[5] == 0x5f && byteMessage[6] == 0x68 && byteMessage[7] == 0x9c;
                        if(result)
                        {
                            fileUploadTarget.send(message);
                            if(byteMessage.length == 8)
                            {
                                fileUploadTarget = null;
                            }
                        }
                    }

               }

               if(fileDownloadTarget != null)
               {
                    let byteMessage = new Uint8Array(message);
                    if(byteMessage.length >= 8)
                    {
                        let result = byteMessage[0] == 0x19 && byteMessage[1] == 0x88 && byteMessage[2] == 0x01 && byteMessage[3] == 0x28
                            && byteMessage[4] == 0x02 && byteMessage[5] == 0x42 && byteMessage[6] == 0xac && byteMessage[7] == 0x12;
                        if(result)
                        {
                            fileDownloadTarget.send(message);
                            if(byteMessage.length == 8)
                            {
                                fileDownloadTarget = null;
                            }
                        }
                    }
               }

            }
            
        });

        socket.on('error', function (error) {
            console.error(error);
        });

        socket.on('close', function (code, message) { 
            /*var closedMessage = "{'senderId':'" + _this.peerIDs[thisSocketId] + "','messageType':'closed','room':'adobarclient'}";
            for (var otherSocket in _this.sockets)
            {
              if ( otherSocket != thisSocketId)
              {
                try
                {
                  _this.sockets[otherSocket].send(closedMessage);
                }
                catch (err) {
                  console.warn("Error in sending message " + closedMessage + ":\n" + err);
                }
              }
            }*/

            delete _this.sockets[thisSocketId];
            delete _this.peerIDs[thisSocketId];
          
        });
    };
    //
    WebsocketNetworkServer.prototype.addSocketServer = function (websocketServer, appConfig) {
        var _this = this;
        websocketServer.on('connection', function (socket) { _this.onConnection(socket); });
    };
    return WebsocketNetworkServer;
}());
exports.WebsocketNetworkServer = WebsocketNetworkServer;
