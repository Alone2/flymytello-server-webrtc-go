package main

import (
	"flymytello-server-webrtc-go/security"
	"fmt"

	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	fmt.Println("Type your password: (invisible)")

	// get password from user
	passwd, err := terminal.ReadPassword(0)
	if err != nil {
		panic(err)
	}

	// hash and save password
	p, err := security.HashPassword(string(passwd))
	if err != nil {
		panic(err)
	}
	err = p.Save()
	if err != nil {
		panic(err)
	}
	fmt.Println("Password hashed successfully")
}
