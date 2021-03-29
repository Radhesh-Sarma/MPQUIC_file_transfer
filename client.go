package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"time"

	quic "github.com/lucas-clemente/quic-go"
)
const BUFF_SIZE = 1024

const threshold = 5 * 1024

func FillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString = retunString + ":"
			continue
		}
		break
	}
	return retunString
}

func main(){
	quicConf := &quic.Config{
		CreatePaths: true,
	}

	fileToSend := "send.txt"
	addr := "localhost:4242"

	file, err := os.Open(fileToSend)
	if err != nil{
		log.Fatalf("Failed Creating file %s",err)
	}

	fInfo, err := file.Stat()
	if err != nil{
		fmt.Println("Error: ",err)
	}
	if fInfo.Size() <= threshold {
		quicConf.CreatePaths = false
		fmt.Println("File is small, using single path")
	} else {
		fmt.Println("File is large, using multipath")
	}
	file.Close()

	session, err := quic.DialAddr(addr, &tls.Config{InsecureSkipVerify: true}, quicConf)
	if err != nil{
		fmt.Println("Error: ",err)
	}

	fmt.Println("Session Created: ",session.RemoteAddr())

	stream, err := sess.OpenStream()
	if err != nil{
		fmt.Println("Error: ",err)
	}

	fmt.Println("stream created...")
    fmt.Println("Client connected")
    sendFile(stream, fileToSend)
    time.Sleep(2 * time.Second)
}

func sendFile(stream quic.Stream, fToSend string){
	defer stream.Close()
	file, err := os.Open(fToSend)
    if err != nil{
		fmt.Println("Error: ",err)
	}

    fInfo, err := file.Stat()
    if err != nil{
		fmt.Println("Error: ",err)
	}

	fSize := FillString(strconv.FormatInt(fInfo.Size(),10),10)
	fName := FillString(fInfo.Name(),64)

	fmt.Println("Sending file name and file size!")
	stream.Write([]byte(fSize))
	stream.Write([]byte(fName))

	sendBuff := make([]byte,BUFF_SIZE)

	var sentBytes int64
	start := time.Now()

	for {
		sentSize, err := file.Read(sendBuff)
		if err != nil{
			break
		}
		stream.Write(sendBuff)
		if err != nil {
            break
        }

		sentBytes += int64(sentSize)
	}
	elapsed := time.Since(start)
	fmt.Print("Transfer took: ",elapsed)
	fmt.Println(" seconds")
	stream.Close()
	stream.Close()
	time.Sleep(2 * time.Second)
	fmt.Println("File has been sent")
	return
}