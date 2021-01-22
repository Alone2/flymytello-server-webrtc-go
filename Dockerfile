FROM golang:1.15

WORKDIR /go/src/flymytello-server-webrtc-go/

ENV PUBLIC_CHAIN_CERT=""
ENV PRIVATE_KEY_CERT=""

RUN go get github.com/pion/webrtc/ && \
    go get github.com/baiyufei/rtp && \
    go get github.com/gorilla/websocket && \
    go get github.com/SMerrony/tello && \
    go get golang.org/x/term && \
    # Patch
    cp -f /go/src/github.com/baiyufei/rtp/codecs/h264_packet.go /go/src/github.com/pion/rtp/codecs/h264_packet.go && \
    mkdir /opt/flymytello

# My Patch
COPY ./docker/h264reader.go /go/src/github.com/pion/webrtc/pkg/media/h264reader/h264reader.go
COPY ./ .
RUN go build . && \
    cd setup && \
    go build .

CMD ./flymytello-server-webrtc-go "$PUBLIC_CHAIN_CERT" "$PRIVATE_KEY_CERT"
