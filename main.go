package main

import (
    "fmt"
    "os"
    "time"

    "flymytello-server-webrtc-go/tello"
    "flymytello-server-webrtc-go/security"
    "flymytello-server-webrtc-go/rtc"
)

// main, everything starts here
func main() {
    if len(os.Args) < 3 {
        fmt.Println("Syntax: ./flymytello-server-webrtc-go [PATH_TO_PUB_KEY] [PATH_TO_PRIV_KEY]")
        return;
    }

    // Check for password
    passwordTool, err := security.GetPassword()
    if err != nil {
        fmt.Println("No password found...")
        fmt.Println("Generate one using setup.sh or ./setup/")
    }
    // Wait till password available
    for err != nil{
        passwordTool, err = security.GetPassword()
        time.Sleep(time.Second * 3)
    }
    fmt.Println("hashed password found")

    // Check if certificates are here
    if os.Args[1] == "" || os.Args[2] == "" {
        fmt.Println("No private or/and public key path as argument (TLS certificate)")
        fmt.Println("Syntax: ./flymytello-server-webrtc-go [PATH_TO_PUB_KEY] [PATH_TO_PRIV_KEY]")
        return
    }

    // New drone
    t, err := tello.NewTellodrone()
    if err != nil {
        fmt.Println("Cannot bind to drone ports")
    }

    // new webserver for signaling
    w, err := rtc.NewTelloWebserver(5001, passwordTool, os.Args[1], os.Args[2])
    if err != nil {
        fmt.Println("cannot create webserver")
    }
    fmt.Println("started webserver...")

    // Fire up webrtc for every incoming connection
    for {
        sig := w.GetSignalerConn()
        fmt.Println("establishing new connection...")
        err = rtc.InitializeRTCVideo(sig, t)
        if err != nil {
            fmt.Println(err)
        } else {
            fmt.Println("connection established")
        }
    }
}

