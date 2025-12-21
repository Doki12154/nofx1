package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
)

func main() {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// Encode private key to PEM format
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Escape newlines for .env file format (replace \n with \\n)
	privEnvValue := strings.ReplaceAll(string(privPEM), "\n", "\\n")

	fmt.Println("请将以下内容复制到 .env 文件的 RSA_PRIVATE_KEY= 后面：")
	fmt.Println()
	fmt.Println(privEnvValue)
}
