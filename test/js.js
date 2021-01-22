let pc = new RTCPeerConnection()
var exampleSocket = new WebSocket("wss://0.0.0.0:5001/signal");

let log = msg => {
  document.getElementById('div').innerHTML += msg + '<br>'
}

pc.ontrack = function (event) {
  var el = document.createElement(event.track.kind)
  el.srcObject = event.streams[0]
  el.autoplay = true
  el.controls = true

  document.getElementById('remoteVideos').appendChild(el)
}

pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
pc.onicecandidate = event => {
  if (event.candidate === null) {
    // document.getElementById('localSessionDescription').value = JSON.stringify(pc.localDescription)
      // exampleSocket.onopen = function (event) {
          setTimeout(() => {
              exampleSocket.send(JSON.stringify(pc.localDescription))
          }, 4000)
      // }
  }
}

// Offer to receive 1 audio, and 2 video tracks
pc.addTransceiver('video', {'direction': 'sendrecv'})
pc.createOffer().then(d => pc.setLocalDescription(d)).catch(log)

exampleSocket.onmessage = function (event) {
    let sd = (event.data);
    console.log(event.data)
    pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(sd)))
} 

