"use strict";
var config = require("./config.json");
var http = require('http');
var https = require('https');
var ws = require('ws');
var fs = require('fs');
var wns = require('./WebsocketNetworkServer');

const express = require("express");
const app = express();
const bodyParser = require("body-parser");

app.use(bodyParser.urlencoded({ extended: false }));
app.use(bodyParser.json());

app.use('/static', express.static('web'))

app.get("/", function (req, res) {
  
  // Sending index.html to the browser
  res.sendFile(__dirname + "/web/index.html");
});

//setup
var httpServer = null;
var httpsServer = null;
if (config.httpConfig) {
    httpServer = http.createServer(app);
    httpServer.listen(process.env.PORT || config.httpConfig.port,  function () { console.log('Listening on ' + httpServer.address().port); });
}

if (config.httpsConfig) {
    httpsServer = https.createServer({
        key: fs.readFileSync(config.httpsConfig.ssl_key_file),
        cert: fs.readFileSync(config.httpsConfig.ssl_cert_file)
    } , app);
    httpsServer.listen(process.env.PORT || config.httpsConfig.port, function () { console.log('Listening on ' + httpsServer.address().port + "(Secure)"); });
}


for (var _i = 0, _a = config.apps; _i < _a.length; _i++) {
    var _app = _a[_i];
    if (httpServer) {
        var websocketSignalingServer1 = new wns.WebsocketNetworkServer();
        //perMessageDeflate: false needs to be set to faflse turning off the compression. if set to true
        //the websocket library crashes if big messages are received (eg.128mb) no matter which payload is set!!!
        var webSocket = new ws.Server({ server: httpServer, path: _app.path, maxPayload: config.maxPayload, perMessageDeflate: true });
        websocketSignalingServer1.addSocketServer(webSocket, _app);
    }
    if (httpsServer) {
        var websocketSignalingServer2 = new wns.WebsocketNetworkServer();
        var webSocketSecure = new ws.Server({ server: httpsServer, path: _app.path, maxPayload: config.maxPayload, perMessageDeflate: true }); //problem in the typings -> setup to only accept http not https so cast to any to turn off typechecks
        websocketSignalingServer2.addSocketServer(webSocketSecure, _app);
    }
}

