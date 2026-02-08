package main

import (
	"bufio"
	"log"
	"net"
	"os"
)

func main() {
	// Resolve UDP address
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("Error resolving address:", err)
	}

	// Establish UDP connection
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal("Error establishing connection:", err)
	}
	defer conn.Close()

	// Create a reader to read from standard input
	reader := bufio.NewReader(os.Stdin)

	for {
		// Print prompt for user input
		print("> ")

		// Read a line from the console
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			continue
		}

		// Send the line as a UDP packet
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Println("Error sending packet:", err)
		}
	}
}