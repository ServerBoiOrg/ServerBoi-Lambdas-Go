package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"

	"golang.org/x/crypto/ssh"
)

func generateSSHKey() (public string, private []byte) {
	bitSize := 2048
	privateKey, err := generatePrivateKey(bitSize)
	if err != nil {
		log.Fatal(err.Error())
	}
	publicKeyBytes := generatePublicKey(&privateKey.PublicKey)
	private = encodePrivateKeyToPEM(privateKey)

	return string(publicKeyBytes), private
}

func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	privateBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateBytes,
	}
	privatePem := pem.EncodeToMemory(&privateBlock)
	return privatePem
}

func generatePublicKey(privatekey *rsa.PublicKey) []byte {
	publicKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		log.Fatalf("Error generating key: %v", err)
	}
	publicBytes := ssh.MarshalAuthorizedKey(publicKey)
	return publicBytes
}
