package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	// this is not a syntax error
	if len(os.Args) == 1 {
		printUsage()
		os.Exit(0)
	}

	// this totally is
	if len(os.Args) < 3 {
		printBadArgument(fmt.Errorf("Not enough arguments"))
	}

	serverPort := os.Args[1]
	clientPort := os.Args[2]

	listen(serverPort, clientPort)
}

func listen(serverPort, clientPort string) {
	ln, err := getListener(serverPort)
	if err != nil {
		printBadArgument(err)
	}
	defer ln.Close()

	for {
		if conn, err := ln.Accept(); err != nil {
			fmt.Printf("ERR: %s\n", err)
		} else {
			go pipeConn(conn, clientPort)
		}
	}
}

func pipeConn(recvConn net.Conn, f string) {
	defer recvConn.Close()

	sendConn, err := net.Dial("tcp", f)
	if err != nil {
		fmt.Printf("ERR: %s\n", err)
		return
	}
	defer sendConn.Close()

	done := make(chan bool, 2)

	p := func(in io.Reader, out io.Writer) {
		if _, err := io.Copy(out, in); err != nil {
			fmt.Printf("ERR: %s\n", err)
		}
	}

	log.Print("Beginning piping")

	go p(recvConn, sendConn)
	go p(sendConn, recvConn)

	<-done
	<-done

	log.Print("Ending piping")
}

func printUsage() {
	fmt.Printf(
		`portpipe: A tiny utility to pipe TCP requests to the server_port into
the client_port

SYNOPSIS
    %s [server_address:server_port] [client_address:client_port]
`, os.Args[0])
}

func printBadArgument(err error) {
	fmt.Println(err)

	printUsage()

	os.Exit(1)
}

func getListener(p string) (net.Listener, error) {
	addr, err := net.ResolveTCPAddr("tcp", p)

	if err != nil {
		return nil, err
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	return ln, nil
}
