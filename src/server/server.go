package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	ntpPort         = 8200
	ntpFractionBits = 32
	ntpEpoch        = 2208988800 // NTP epoch time in seconds (1900-1970)
)

type packet struct {
	Settings       uint8
	Stratum        uint8
	Poll           int8
	RootDelay      uint32
	RootDispersion uint32
	ReferenceID    uint32
	Precision      int8
	RefTimeSec     uint32
	RefTimeFrac    uint32
	OrigTimeSec    uint32
	OrigTimeFrac   uint32
	RxTimeSec      uint32
	RxTimeFrac     uint32
	TxTimeSec      uint32
	TxTimeFrac     uint32
}

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", "0.0.0.0", ntpPort))
	if err != nil {
		fmt.Println("Error resolving address:", err)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("NTP server listening on port %d...\n", ntpPort)

	for {
		handleClient(conn)
	}
}

func handleClient(conn *net.UDPConn) {
	buffer := make([]byte, 48)

	_, clientAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error reading from client:", err)
		return
	}

	// Get current time
	currentTime := getCurrentNTPTime()
	fmt.Println("Current NTP Time:", currentTime)

	p := packet{}
	p.Settings = 0x1B
	p.Stratum = 2
	p.Poll = -6
	p.Precision = -20
	p.RefTimeSec = 0xDEADBEEF
	p.RefTimeFrac = 0xCAFE
	p.OrigTimeSec = binary.BigEndian.Uint32(buffer[40:44])
	p.OrigTimeFrac = binary.BigEndian.Uint32(buffer[44:48])
	p.RxTimeSec = currentTime.Seconds
	p.RxTimeFrac = currentTime.FractionPart
	p.TxTimeSec = currentTime.Seconds
	p.TxTimeFrac = currentTime.FractionPart

	fmt.Println("Packet Data:", p)

	response := make([]byte, 48)
	binary.BigEndian.PutUint32(response[32:], p.RxTimeSec)
	binary.BigEndian.PutUint32(response[36:], p.RxTimeFrac)
	binary.BigEndian.PutUint32(response[40:], p.TxTimeSec)
	binary.BigEndian.PutUint32(response[44:], p.TxTimeFrac)

	fmt.Println("response:", response)
	_, err = conn.WriteToUDP(response, clientAddr)
	if err != nil {
		fmt.Println("Error writing response to client:", err)
	}
}

func getCurrentNTPTime() ntpTime {
	now := time.Now()
	secs := uint32(now.Unix()) + ntpEpoch // NTP seconds since epoch
	fraction := uint32(now.Nanosecond()*(1<<ntpFractionBits)) / uint32(time.Second/time.Nanosecond)
	fmt.Printf("Current Unix Time: %d\n", now.Unix())
	fmt.Printf("Fraction Part: %d\n", fraction)
	return ntpTime{Seconds: secs, FractionPart: fraction}
}

type ntpTime struct {
	Seconds      uint32
	FractionPart uint32
}
