package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"io/ioutil"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func main() {
	dst := flag.String("d", "127.0.0.1", "DestinationIP address")
	id := flag.Int("id", 0, "Identifier 1-65535")
	f := flag.String("f", "", "File Path to send")
	s := flag.Int("s", 8980, "Max packet size")

	helpFlag := flag.Bool("help", false, "Display help message")

	// Parse the command-line flags
	flag.Parse()

	if *f == "" || *id == 0 {
		printHelp()
		os.Exit(0)
	}

	// Help flag
	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	// Destination IP address
	destination := *dst
	identifier := *id
	filePath := *f
	maxSize := *s

	// Create a connection
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Fatal("Error creating connection:", err)
		return
	}
	defer conn.Close()

	// Read the file
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	// Split the file into smaller chunks
	chunks := splitData(fileData, maxSize)

	// Send each chunk to the destination
	for _, chunk := range chunks {

		// Prepare the ICMP body
		name := filePath // Nama
		nameBytes := []byte(name)

		// Check the length of the name
		if len(nameBytes) > 100 {
			nameBytes = nameBytes[:100]
		} else if len(nameBytes) < 100 {
			// If the name is shorter than 100 bytes, add padding
			padding := make([]byte, 100-len(nameBytes))
			nameBytes = append(nameBytes, padding...)
		}

		// Create the final message by concatenating the name and buffer
		bodyData := append(nameBytes, chunk...)

		// Create ICMP message
		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   identifier, // Set the ICMP identifier value here
				Seq:  1,
				Data: bodyData,
			},
		}

		// Serialize the ICMP message
		msgBytes, err := msg.Marshal(nil)
		if err != nil {
			log.Fatal("Error marshaling ICMP message:", err)
			return
		}

		// Resolve destination IP address
		destIP := net.ParseIP(destination)
		if destIP == nil {
			log.Fatal("Invalid destination IP address")
			return
		}

		// Send ICMP message to destination
		_, err = conn.WriteTo(msgBytes, &net.IPAddr{IP: destIP})
		if err != nil {
			log.Fatal("Error sending ICMP message:", err)
			return
		}

		fmt.Printf("%v bytes ICMP packet sent successfully to %v!\n", len(chunk), destIP)

		time.Sleep(200 * time.Millisecond)
	}

}

// Split data into smaller chunks
func splitData(data []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}

func printHelp() {
	fmt.Println("Sending file through ICMP")
	fmt.Println("Usage: send.go [options]")
	fmt.Println("Options:")
	flag.PrintDefaults()
}