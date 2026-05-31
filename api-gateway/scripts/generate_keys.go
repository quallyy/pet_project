package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
)

func main() {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	// Create keys directory
	err = os.MkdirAll("keys", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	// Save private key
	privateFile, err := os.Create("keys/private.pem")
	if err != nil {
		log.Fatal(err)
	}
	defer privateFile.Close()

	privateBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	err = pem.Encode(privateFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateBytes,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Save public key
	publicFile, err := os.Create("keys/public.pem")
	if err != nil {
		log.Fatal(err)
	}
	defer publicFile.Close()

	publicBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	err = pem.Encode(publicFile, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicBytes,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("RSA keys generated in /keys")
}