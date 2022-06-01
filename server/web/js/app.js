
var peerConnection = null;
var dataChannel = null;
var colorMode = 0;
var oldPeers = null;
var selectedPeer = null;
var downloadBuffer = null;

let resolutionMap = {
  screenWidth: 0,
  screenHeight: 0,
  canvasWidth: 0,
  canvasHeight: 0,
};

var remoteVideo ;
var remoteCanvas ;
var explorerDirectory = "";
var selectedDirectory = "";
var selectedFile = "";
var webSocket = new WebSocket("wss://127.0.0.1:19890");

function peerChange()
{
  selectedPeer = document.getElementById("peer-list").value;
  console.log(selectedPeer);
}

function modeChange()
{
  colorMode = document.getElementById("color-mode").value;
  sendDataMessage("color", {colorMode,});
}

function sendDataMessage(command, data) {
    if (dataChannel) {
      // Send cordinates
      dataChannel.send(
        JSON.stringify({
          command: command,
          data: data,
        })
      );
    }
  }

document.addEventListener("DOMContentLoaded", () => {
  let selectedScreen = 0;
  remoteVideo = document.querySelector("#remote-video");
  remoteCanvas = document.querySelector("#remote-canvas");
  
  // Disable right click context on canvas
  remoteCanvas.oncontextmenu = function (e) {
    e.preventDefault();
  };

  const startStop = document.querySelector("#start-stop");

  function showError(error) {
    const errorNode = document.querySelector("#error");
    if (errorNode.firstChild) {
      errorNode.removeChild(errorNode.firstChild);
    }
    errorNode.appendChild(document.createTextNode(error.message || error));
  }

  remoteVideo.onplaying = () => {
    setInterval(() => {
      resizeCanvas(remoteCanvas, remoteVideo);
    }, 1000);
  };

  startStop.addEventListener("click", () => {

    if(!(webSocket && webSocket.readyState == 1) || selectedPeer == null || selectedPeer == "")
      return;

    enableStartStop(false);

    const userMediaPromise =
      adapter.browserDetails.browser === "safari"
        ? window.navigator.mediaDevices.getUserMedia({ video: true })
        : Promise.resolve(null);
    if (!peerConnection) {
      userMediaPromise.then((stream) => {

        return startRemoteSession(selectedScreen, remoteVideo, stream)
          .then(() => {
            
          })
          .catch(showError)
          .then(() => {
            
          });

      });
    } else {
      disconnectSession();
      remoteVideo.style.setProperty("visibility", "collapse");
    }
  });

  function uuidv4() {
    return ([1e7]+-1e3+-4e3+-8e3+-1e11).replace(/[018]/g, c =>
      (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
    );
  }

  function concatTypedArrays(a, b) { // a, b TypedArray of same type
      var c = new (a.constructor)(a.length + b.length);
      c.set(a, 0);
      c.set(b, a.length);
      return c;
  }

  function upload_fin()
  {
    if(selectedPeer != null)
    {
       let magic = new Uint8Array(8);
        magic[0] = 0x19; magic[1] = 0x88; magic[2] = 0x01; magic[3] = 0x28;
        magic[4] = 0xf5; magic[5] = 0x5f; magic[6] = 0x68; magic[7] = 0x9c;
        webSocket.send(magic);
    }
  }
  document.getElementById('file-upload').addEventListener('change', async (event) => {
      if(selectedPeer == null) return;
      file = event.target.files[0];
      fileReader = new FileReader();
      if (file == undefined || file.size === 0) {
        return ;
      }
      let offset = 0;
      fileReader.addEventListener('error', error => {console.error('Error reading file:', error); upload_fin();});
      fileReader.addEventListener('abort', event => {console.log('File reading aborted:', event); upload_fin();});
      fileReader.addEventListener('load', e => {
        
        const span = document.getElementById('upload-label');
        span.innerHTML  = "Cancel (" + Math.floor((offset / file.size) * 100) + "%)";

        $('.status-bar').text("Uploading (" + Math.floor((offset / file.size) * 100) + "%)");
        let bytes = new Uint8Array(e.target.result);
        let magic = new Uint8Array(8);
        magic[0] = 0x19; magic[1] = 0x88; magic[2] = 0x01; magic[3] = 0x28;
        magic[4] = 0xf5; magic[5] = 0x5f; magic[6] = 0x68; magic[7] = 0x9c;
        let mergedBytes = new Uint8Array(bytes.length + 8)

        mergedBytes.set(magic);
        mergedBytes.set(bytes, 8);
        webSocket.send(mergedBytes);

        offset += e.target.result.byteLength;
        
        if (offset < file.size) {
          readSlice(offset);
        }
        else
        {
          const span = document.getElementById('upload-label');
          span.innerHTML  = 'Upload';
          $('.status-bar').text(file.name + " successfully uploaded");
          upload_fin();
        }
      });

      const readSlice = o => {
        const slice = file.slice(offset, o + MAXIMUM_CHUNK_SIZE);
        fileReader.readAsArrayBuffer(slice);
      };

      webSocket.send(JSON.stringify({
          Type: "file_upload",
          Data: JSON.stringify({Name: file.name , Path: selectedDirectory}) ,
          Peer: selectedPeer
        }));

      readSlice(0);


    });

  
  webSocket.onopen = function (event) {
    
    var msg = {
          Type: "init",
          Data: uuidv4(),
          Attribute:"client"
        };
    webSocket.send(JSON.stringify(msg));

    setInterval(async () => {

      if(webSocket && webSocket.readyState == 1)
      {
        msg = {
          Type: "peers",
          Data: ""
        };
        webSocket.send(JSON.stringify(msg));

      }
    }, 1000);

  };

  webSocket.onmessage = async function  (event) {

    if(typeof event.data === "string")
    {
      var msg = JSON.parse(event.data);
      if(msg.Type == "peers")
      {
        if(JSON.stringify(oldPeers) != JSON.stringify(msg.Data) )
        {
          oldPeers = msg.Data;
          var x = document.getElementById("peer-list");
          var options = document.querySelectorAll('#peer-list option');
          options.forEach(o => o.remove());

          if(msg.Data.length == 0)
          {
            var option = document.createElement("option");
            option.value = "";
            option.text = "No Peer";
            x.add(option);
            selectedPeer = null;
          }
          else
          {
            for(var id in msg.Data)
            {
              var option = document.createElement("option");
              option.value = msg.Data[id];
              option.text = msg.Data[id];
              
              if(selectedPeer == null)
                selectedPeer = msg.Data[id];
              
              if(msg.Data[id] == selectedPeer)
              {
                option.selected = true;
              }
              x.add(option);
            }
          }
        }
      }
      else if(msg.Type == "session_response")
      {
        peerConnection.setRemoteDescription(
          new RTCSessionDescription({
            sdp: msg.Data,
            type: "answer",
          })
        );

        enableStartStop(true);
        CtrlPannel(true);
        setStartStopTitle("Disconnect");
        remoteVideo.style.setProperty("visibility", "visible");
      }
      else if (msg.Type == "session_request")
      {
        // No implemented for brwoser-browser
      }
    } 
    else if(typeof event.data === "object")
    {
      let arrayBuffer = await event.data.arrayBuffer();
      let bytes = new Uint8Array(arrayBuffer);
      let bytesBody = new Uint8Array(arrayBuffer.slice(8));
      if(bytes.length >= 8)
      {
        let result = bytes[0] == 0x19 && bytes[1] == 0x88 && bytes[2] == 0x01 && bytes[3] == 0x28
                        && bytes[4] == 0x02 && bytes[5] == 0x42 && bytes[6] == 0xac && bytes[7] == 0x12;
        if(result)
        {
          if(bytes.length == 8)
          {
            let blob = new Blob([downloadBuffer]);
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.style.display = 'none';
            a.href = url;
            a.download = selectedFile;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            downloadBuffer = null;
            $('.status-bar').text(selectedFile + " successfully downloaded");
          }
          else 
          {
            if(downloadBuffer == null)
            {
              downloadBuffer = bytesBody;
            }
            else
            {
              let _bytesBody = new Uint8Array(downloadBuffer.length + bytesBody.length);
              _bytesBody.set(downloadBuffer);
              _bytesBody.set(bytesBody , downloadBuffer.length);
              downloadBuffer = _bytesBody;
            }
            $('.status-bar').text("Downloading ... ");
          }
        }
      }
      
    }
    else
    {
      console.log(typeof event.data);
    }

  }

  
  function createOffer(pc, { audio, video }) {

    setStartStopTitle("Waiting signal...");
    return new Promise((accept, reject) => {
      pc.onicecandidate = (evt) => {
        if (!evt.candidate) {
          // ICE Gathering finished
          const { sdp: offer } = pc.localDescription;
          accept(offer);
          setStartStopTitle("Connecting ...");
        }
      };
      pc.createOffer({
        offerToReceiveAudio: audio,
        offerToReceiveVideo: video,
      })
        .then((ld) => {
          pc.setLocalDescription(ld);
        })
        .catch(reject);
    });
  }

  function keydownListener(event)
  {
      sendDataMessage("keydown", {
        keyCode: event.keyCode,
        ctrl:event.ctrlKey,
        shift:event.shiftKey
      });
  }

  function disableEvents() {
    const remoteCanvas = document.getElementById("remote-canvas");
    let newCanvas = remoteCanvas.cloneNode(true);
    remoteCanvas.replaceWith(newCanvas);
    // Disable right click context on canvas
    newCanvas.oncontextmenu = function (e) {
      e.preventDefault();
    };
    
    document.removeEventListener("keydown", keydownListener);
  }

  function enableEvents(dataChannel) {
    // Start sending mouse cordinates on mouse move in canvas
    const remoteCanvas = document.getElementById("remote-canvas");

    // On Mouse move
    remoteCanvas.addEventListener("mousemove", (event) => {
      // Get cordinates
      const cordinates = scaleCordinatesToOriginalScreen(event);

        // Send cordinates
        sendDataMessage("mousemove", {
          x: cordinates.x,
          y: cordinates.y,
        });

    });

    var mouseDownT = null;
    var leftClickHandler = null;

    function checkMouseDownEvent()
    {
      let button = "left";
      leftClickHandler = null;
      sendDataMessage("mousedown", {button,});
      
    }

    // On Mouse Click
    remoteCanvas.addEventListener("mousedown", (event) => {
      let button = "left";
      switch (event.button) {
        case 0:
          button = "left";
          break;

        case 1:
          button = "center";
          break;

        case 2:
          button = "right";
          break;

        default:
          button = "left";
      }

      if(button == "left")
      {
        mouseDownT = Date.now();
        leftClickHandler = setTimeout(checkMouseDownEvent, 100);
      } 
      else
      {
        sendDataMessage("click", {button,});  
      }
      
    });

    remoteCanvas.addEventListener("mouseup", (event) => {
      
      switch (event.button) {
        case 0:
          if(leftClickHandler != null || ( (Date.now() - mouseDownT) < 100))
          {
            clearTimeout(leftClickHandler);
            leftClickHandler = null;
            let button = "left";
            sendDataMessage("click", {button,});
          }
          else
          {
            sendDataMessage("mouseup", {});
          }
          break;

        default:
      }

     
      
    });
    // On Mouse Double Click
    remoteCanvas.addEventListener("dblclick", (event) => {
      let button = "left";

      switch (event.which) {
        case 1:
          button = "left";
          break;

        case 2:
          button = "center";
          break;

        case 3:
          button = "right";
          break;

        default:
          button = "left";
      }

      sendDataMessage("dblclick", {
        button,
      });
    });

    // On Mouse Scroll
    remoteCanvas.addEventListener("wheel", (event) => {
      const delta = Math.sign(event.deltaY);
      const direction = delta > 0 ? "down" : "up";
      sendDataMessage("mousescroll", {
        direction,
      });
    });

    document.addEventListener("keydown", keydownListener);
  
  }

  var _clipboardText = "";
  setInterval(async () => {

    try{
      var clipboardText = await window.navigator.clipboard.readText(); //window.clipboardData.getData('Text');
      if(clipboardText != _clipboardText)
      {
        sendDataMessage("clipboard", {
          clipboardText,
        });
        _clipboardText = clipboardText;
      }
    }catch (e){
      
    }
  }, 500);

  function scaleCordinatesToOriginalScreen(event) {
    const remoteCanvas = document.getElementById("remote-canvas");
    // Get canvas size
    const rect = remoteCanvas.getBoundingClientRect();
    // Get mouse cordinates on canvas
    const x = (event.clientX - rect.left).toFixed(0);
    const y = (event.clientY - rect.top).toFixed(0);
    // Calculate screen percentage based on canvas
    const xPer = (x / resolutionMap.canvasWidth) * 100;
    const yPer = (y / resolutionMap.canvasHeight) * 100;
    // Map percentage to original screen
    return {
      x: ((resolutionMap.screenWidth * xPer) / 100).toFixed(0),
      y: ((resolutionMap.screenHeight * yPer) / 100).toFixed(0),
    };
  }

  const MAXIMUM_CHUNK_SIZE = 512 * 1024;
  const END_OF_FILE_MESSAGE = 'EOF';
  let fileReader;

  function startRemoteSession(screen, remoteVideoNode, stream) {
    let pc;

    return Promise.resolve()
      .then(() => {

        setStartStopTitle("Preparing ...");

        pc = new RTCPeerConnection({
          iceServers: [
            {
              urls: "stun:stun1.l.google.com:19302",
            },
          
            /*{
              urls: "turn:ntk-turn-1.xirsys.com:80?transport=udp",
              username: "",
              credential: "",
            },
            {
              urls: "turn:ntk-turn-1.xirsys.com:80?transport=tcp",
              username: "",
              credential: "",
            },*/
        ],
        });

        function sleep(ms) {
          return new Promise(resolve => setTimeout(resolve, ms));
        } 

        dataChannel = pc.createDataChannel("messages");

        dataChannel.onopen = function (event) {
          enableEvents(dataChannel);

          // Fetch screen size from server
          sendDataMessage("screensize", {});
        };

        dataChannel.onmessage = function (event) {
          try {
            const message = JSON.parse(event.data);
            switch (message.command) {
              case "screensize":
                resolutionMap.screenHeight = message.data.height;
                resolutionMap.screenWidth = message.data.width;
                break;

              case "mousepose":
                console.log(message);
                break;
              case "clipboard":
                try{
                  navigator.clipboard.writeText(message.data);
                }catch(e){}
                break;
              case "directory":
                
                $('.list-group').empty();

                if(message.data.Current != "")
                  $( ".list-group" ).append( '<li class="list-group-item explorer-item" Name=".." Type=1>..</li>');

                explorerDirectory = message.data.Current;
                if(explorerDirectory == "")
                  $('.info-bar').text("");
                else
                  $('.info-bar').text("Path: " + explorerDirectory);

                message.data.Directories.forEach((item, index) => {
                  let icon = '';
                  if(item.Type == 0)
                    icon = "bi-hdd-fill";
                  else if(item.Type == 1)
                    icon = "bi-folder-fill";
                  else
                    icon = "bi-file-text-fill";
                  $( ".list-group" ).append( '<li class="list-group-item explorer-item" Name="' + item.Name + '" Type=' + item.Type + '><i class="' + icon + '" style="padding-right:10px"></i>' + item.Name + '</li>');
                  
                });

                $( ".explorer-item" ).click(function() {
                    $(".explorer-item").removeClass("selected");
                    $(this).addClass("selected");
                });

                $( ".explorer-item" ).dblclick(function() {

                  if($(this).attr("Type") == 0 || $(this).attr("Type") == 1)
                  {
                    let newDirectory;
                    if($(this).attr("Name") == "..")
                    {
                      if(explorerDirectory == "")
                        newDirectory = ""
                      else
                        newDirectory =  explorerDirectory  + "\\..\\";
                    }
                    else
                    {
                      if($(this).attr("Type") == 0)
                        newDirectory =  $(this).attr("Name")  + "\\";
                      else
                        newDirectory =  explorerDirectory + "\\" + $(this).attr("Name");
                    }
                    var directory = newDirectory;
                    sendDataMessage("directory", {directory,});
                  }
                });

                $(".explorer-item").contextmenu(function(event) {
                  
                    // Avoid the real one
                    event.preventDefault();
                    
                    // Show contextmenu
                    $(".custom-menu").finish().toggle(100).
                    
                    // In the right position (the mouse)
                    css({
                        top: event.pageY + "px",
                        left: event.pageX + "px"
                    });

                    $(".custom-menu").empty();
                    if($(this).attr('Type') == 0 || $(this).attr('Type') == 1)
                    {
                      if($(this).hasClass("selected"))
                      {
                        if($(this).attr('Type') == 0)
                          $(".custom-menu").append('<li data-action="upload" data="' + $(this).attr("Name") + '"><i class="bi-cloud-upload-fill" style="padding-right:10px"></i> Upload </li>');
                        else
                          $(".custom-menu").append('<li data-action="upload" data="' + explorerDirectory  + "\\" + $(this).attr("Name") + '"><i class="bi-cloud-upload-fill" style="padding-right:10px"></i> Upload </li>');
                      }
                      else
                      {
                        $(".custom-menu").append('<li data-action="upload" data="' + explorerDirectory  + '"><i class="bi-cloud-upload-fill" style="padding-right:10px"></i> Upload </li>');
                      }

                    } else {
                      
                      if($(this).hasClass("selected"))
                        $(".custom-menu").append('<li data-action="download" data="' + $(this).attr("Name") + '"><i class="bi-cloud-download-fill" style="padding-right:10px"></i> Download </li>');
                      else
                        $(".custom-menu").append('<li data-action="upload" data="' + explorerDirectory  + '"><i class="bi-cloud-upload-fill" style="padding-right:10px"></i> Upload </li>');
                    }

                    $(".custom-menu li").click(function(){
                      
                      // This is the triggered action name
                      switch($(this).attr("data-action")) {
                          
                          // A case for each action. Your actions here
                          case "upload": 
                            $( "#file-upload" ).trigger( "click" );
                            selectedDirectory = $(this).attr("data");
                            break;
                          case "download": 
                            selectedDirectory = explorerDirectory;
                            if(!(webSocket && webSocket.readyState == 1) || selectedPeer == null || selectedPeer == "")
                            {
                              $('.status-bar').text("Server connection has been closed.");
                            }
                            else
                            {
                              downloadBuffer = null;
                              selectedFile = $(this).attr("data");
                                webSocket.send(JSON.stringify({
                                Type: "file_download",
                                Data: {Name: selectedFile , Path: selectedDirectory} ,
                                Peer: selectedPeer
                              }));
                            }
                            break;
                      }
                    
                      // Hide it AFTER the action was triggered
                      $(".custom-menu").hide(100);
                  });
                });

                $(document).bind("mousedown", function (e) {
                    
                    // If the clicked element is not the menu
                    if (!$(e.target).parents(".custom-menu").length > 0) {
                        
                        // Hide it
                        $(".custom-menu").hide(100);
                    }
                });

                

                break;
            }
          } catch (e) {
            console.error(e);
          }
        };

        dataChannel.onclose = function (event) {
          disableEvents();
        }

        pc.oniceconnectionstatechange = function ()
        {
          if(pc.iceConnectionState == 'disconnected') {
            disconnectSession();
        }
        }

        pc.ontrack = (evt) => {
          remoteVideoNode.srcObject = evt.streams[0];
          remoteVideoNode.play();
        };

        stream &&
          stream.getTracks().forEach((track) => {
            pc.addTrack(track, stream);
          });

        return createOffer(pc, { audio: false, video: true });
      })
      .then((offer) => {

        webSocket.send(JSON.stringify({
          Type: "session_request",
          Data:{
          offer,
          screen,
          "peer": selectedPeer
          }
        }));

        peerConnection = pc;
        return;
      });
  }

  function resizeCanvas(canvas, video) {
    const w = video.offsetWidth;
    const h = video.offsetHeight;
    canvas.width = w;
    canvas.height = h;

    resolutionMap.canvasHeight = h;
    resolutionMap.canvasWidth = w;
  }

  function disconnectSession() {
    sendDataMessage("terminate", {});
    peerConnection.close();
    peerConnection = null;
    dataChannel = null;
    enableStartStop(true);
    CtrlPannel(false);
    setStartStopTitle("Connect");

    const remoteCanvas = document.getElementById("remote-canvas");
    var newRemoteCanvas = remoteCanvas.cloneNode(true);
    remoteCanvas.parentNode.replaceChild(newRemoteCanvas, remoteCanvas);

  }

  const CtrlPannel = (enabled) =>
  {
    if (enabled) {
      ctrl_btn = document.querySelector("#color-mode");
      ctrl_btn.style.setProperty("display", "")
    } else {
      
      ctrl_btn = document.querySelector("#color-mode");
      ctrl_btn.style.setProperty("display", "none")
    }
  }

  const enableStartStop = (enabled) => {
    const startStop = document.querySelector("#start-stop");
    if (enabled) {
      startStop.removeAttribute("disabled");
      
    } else {
      startStop.setAttribute("disabled", "");
      
    }
  };

  const setStartStopTitle = (title) => {
    const startStop = document.querySelector("#start-stop");
    startStop.removeChild(startStop.firstChild);
    startStop.appendChild(document.createTextNode(title));
  };

});

window.addEventListener("beforeunload", () => {
  if (peerConnection) {
    peerConnection.close();
  }
});

$( document ).ready(function() {


    $("#explorer-btn" ).click(function() {
      var directory = explorerDirectory;
      sendDataMessage("directory", {directory,});
      $('#explorer').modal('show');
  });
});

