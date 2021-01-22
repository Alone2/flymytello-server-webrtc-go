package security

import (
    "io/ioutil"

    "golang.org/x/crypto/bcrypt"
)

const (
    // Path of hashed password
    path string =  "/opt/flymytello/password"
)

// Struct with password hash+salt saved 
type PasswordStruct struct {
    hash []byte
}

// https://gowebexamples.com/password-hashing/

// Hash a passwort return PasswordStruct
func HashPassword(password string) (*PasswordStruct, error) {
    // bycrypt adds salt automatically to hash
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 13)
    if err != nil {
        return &PasswordStruct{}, err
    }
    p := PasswordStruct{ 
        hash: bytes, 
    }
    return &p, err
}

// Check if a password is the same as the hashed one in PasswordStruct
func (p *PasswordStruct) CheckHash(password string) (bool, error) {
    err := bcrypt.CompareHashAndPassword(p.hash, []byte(password))
    return err == nil, nil
}

// Get PasswordStruct from file (const path)
func GetPassword() (*PasswordStruct, error) {
    dat, err := ioutil.ReadFile(path)
    if err != nil {
        return &PasswordStruct{}, err
    }
    p := PasswordStruct{
        hash: dat,
    }
    return &p, nil
}

// Save hashed password in PasswordStruct to file (const path)
func (p *PasswordStruct) Save() (error) {
    err := ioutil.WriteFile(path, p.hash, 0700)
    if err != nil {
        return err
    }
    return nil
}
