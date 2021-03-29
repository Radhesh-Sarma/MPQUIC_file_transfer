package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"strings"
	"strconv"
	"time"

	quic "github.com/lucas-clemente/quic-go"
)

const BUFF_SIZE = 1024
const addr = "localhost:4242"

func GenerateTLSConfig() *tls.Config{
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	templt := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &templt, &templt, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}

func main(){
	quicConf := &quic.Config{
		CreatePaths: true,
	}
	listener, err := quic.ListenAddr(addr, GenerateTLSConfig(), quicConf)
	if err != nil{
		fmt.Println("Error: ",err)
		os.Exit(1)
	}

	fmt.Println("Server Started")

	session, err := listener.Accept()
	if err != nil{
		fmt.Println("Error: ",err)
		os.Exit(1)
	}

	fmt.Println("Session Created: ",session.RemoteAddr())

	stream, err := session.AcceptStream()
	if err != nil{
		fmt.Println("Error: ",err)
		os.Exit(1)
	}

	fmt.Println("Stream Created: ",stream.StreamID())

	defer stream.Close()

	fmt.Println("Connected to Server")
	BuffFName := make([]byte, 64)
	BuffFSize := make([]byte, 10)

	stream.Read(BuffFSize)
	fsize, _ := strconv.ParseInt(strings.Trim(string(BuffFSize), ":"),10,64)
	fmt.Println("File size: ",fsize)

	stream.Read(BuffFName)
	fname := strings.Trim(string(BuffFName), ":")

	fpath, err := os.Create("Receive.txt")
	if err != nil{
		log.Fatalf("Failed Creating file %s",err)
	}

	defer fpath.Close()
	var recvBuff int64
	start := time.Now()

	for {
		if(fsize - recvBuff) < BUFF_SIZE{
			recv, err := io.CopyN(fpath,stream,(fsize - recvBuff))
			if err != nil{
				fmt.Println("Error: ",err)
				os.Exit(1)
			}
			stream.Read(make([]byte, (recvBuff + BUFF_SIZE) - fsize))
			recvBuff += recv
			break
		}
		_, err := io.CopyN(fpath, stream, BUFF_SIZE)
		if err != nil{
			fmt.Println("Error: ",err)
			os.Exit(1)
		}
		recvBuff += BUFF_SIZE
	}
	elapsed := time.Since(start)
	fmt.Print("Transfer took: ",elapsed)
	fmt.Println(" seconds")
	stream.Close()
	stream.Close()
	fmt.Println("File Transfer Completed Successfully")
}