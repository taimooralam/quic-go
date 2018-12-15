
package main

import (
	"bytes"
	"bufio"
	"log"
	//"io/ioutil"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"encoding/gob"
	"fmt"
	"math/big"
	"syscall"
	"os"

	quic "github.com/lucas-clemente/quic-go"
)


const addr = "127.0.0.1:4242"
var pipe_length = "/tmp/pipe_length.log"
var pipe_data = "/tmp/pipe_data.log"

 type Packet struct {
          SequenceNumber int32
          TimeStamp int64
          Data string
 }

func check(e error){
	if e!= nil {
		panic(e)
	}
}

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {
	fmt.Printf("Hello world\n")
	//go func() { log.Fatal(echoServer()) }()
	echoServer()
}

// Start a server that echos all data on the first stream opened by the client
func echoServer() error {

	//create the two pipes, one for sending the length and one for sending the data
	os.Remove(pipe_length);
	os.Remove(pipe_data);

	log.Print("making fifo")
	err := syscall.Mkfifo(pipe_length, 0777);
	if err != nil {
		log.Fatal("Make named pipe error", err)
		return err
	}

	log.Print("fifomade")

	err = syscall.Mkfifo(pipe_data, 0777);
	if err != nil {
		log.Fatal("Make named pipe error: ", err)
		return err
	}
	
	log.Println("Opening the two pipes for data")
	file_length, err := os.OpenFile(pipe_length, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0777)
	if err != nil {
		log.Fatal("Open named pipe file_length error:", err)
		return err
	}

	file_data, err := os.OpenFile(pipe_data, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0777)
	if err != nil {
		log.Fatal("Open named pipe file_data error:", err)
		return err
	}



	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	check(err)

	sess, err := listener.Accept()
	check(err)

	stream, err := sess.AcceptStream()
	check(err)

	//create a reader
	reader := bufio.NewReader(stream)



	
	//i := 20
	//file_length.WriteString(fmt.Sprintf("%zu", i))
	//file_length.WriteString(uint32(i)) //convert the length to unint32
	//file_data.WriteString("abddhftysdfasdfadsfadsfasdfadsfasdfhabdd~")




	//i := 0
	for {
		//read bytes until ` is encountered
		line, err := reader.ReadBytes('`')

		if err == nil{

			//gob decode into packet
			decBuf := bytes.NewBuffer(line)
			packet := Packet{}
			err = gob.NewDecoder(decBuf).Decode(&packet)

			//print the length of packet.data
			log.Println("Length of the packet", len(packet.Data))
			//log.Println(packet.Data[len(packet.Data)-51:])

			fmt.Printf("%d", len(packet.Data))
			file_length.WriteString(fmt.Sprintf("%d", len(packet.Data))) //convert the length to unint32
			//file_data.WriteString(packet.Data)
			file_data.Write([]byte(packet.Data))
			//err = ioutil.WriteFile(pipe_data,[]byte(packet.Data),0777)
			//if err != nil {
			//	log.Fatal("Error occured", err)
			//}

		//log.Print(file_length)
		//log.Print(file_data)
		}
	}
	return err
}



// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
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
