package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

const (
	ntpPort           = 8200
	initialPollRate   = 16 * time.Second
	maxBackoffPolling = 36 * time.Hour
	burstInterval     = 2 * time.Second
	minOffsetInterval = 256 * time.Second
	maxOffsetInterval = 512 * time.Second
	numBurstPackets   = 5
	timeoutDuration   = 5 * time.Second
)

var offset int64 // Global variable to hold the offset

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
	serverAddr := "localhost"
	//serverAddr := "us.pool.ntp.org"
	client := &ntpClient{
		serverAddr: serverAddr,
	}

	err := client.start()
	if err != nil {
		log.Fatal(err)
	}
}

type ntpClient struct {
	serverAddr string
	Delays     []float64 // Slice to hold delay values per iteration
	Offsets    []float64 // Slice to hold offset values per iteration
}

func (c *ntpClient) start() error {
	// Create delay plot
	delayPlot := plot.New()

	// Set delay plot title and axis labels
	delayPlot.Title.Text = "NTP Delay vs Iterations"
	delayPlot.X.Label.Text = "Iteration"
	delayPlot.Y.Label.Text = "Milliseconds"

	// Create offset plot
	offsetPlot := plot.New()

	// Set offset plot title and axis labels
	offsetPlot.Title.Text = "NTP Offset vs Iterations"
	offsetPlot.X.Label.Text = "Iteration"
	offsetPlot.Y.Label.Text = "Milliseconds"

	// Calculate the total number of bursts for one hour
	totalBursts := int((time.Hour).Minutes())

	for burst := 0; burst < totalBursts/10; burst++ {
		var totalOffset int64
		var totalDelay float64

		// Perform an eight-packet burst every 4 minutes
		for i := 0; i < numBurstPackets; i++ {
			delay, off, err := c.sendRequest()
			if err != nil {
				return err
			}
			totalDelay += delay
			totalOffset += off
			time.Sleep(burstInterval)
		}

		// Update the global offset after each burst
		offset += totalOffset / int64(numBurstPackets)

		// Add delay and offset values to slices
		c.Delays = append(c.Delays, totalDelay/float64(numBurstPackets))
		c.Offsets = append(c.Offsets, float64(totalOffset / int64(numBurstPackets)))

		// Sleep for 1 minute before the next burst
		time.Sleep(1 * time.Minute)
	}

	// Plot delay values
	delayPoints := make(plotter.XYs, len(c.Delays))
	for i, d := range c.Delays {
		delayPoints[i].X = float64(i)
		delayPoints[i].Y = d
	}
	line, points, err := plotter.NewLinePoints(delayPoints)
	if err != nil {
		return err
	}
	line.Color = plotter.DefaultLineStyle.Color
	points.Shape = plotter.DefaultGlyphStyle.Shape
	points.Color = plotter.DefaultGlyphStyle.Color
	delayPlot.Add(line, points)

	// Save delay plot to a file
	if err := delayPlot.Save(6*vg.Inch, 4*vg.Inch, "ntp_delay.png"); err != nil {
		return err
	}

	// Plot offset values
	offsetPoints := make(plotter.XYs, len(c.Offsets))
	for i, o := range c.Offsets {
		offsetPoints[i].X = float64(i)
		offsetPoints[i].Y = o
	}
	line, points, err = plotter.NewLinePoints(offsetPoints)
	if err != nil {
		return err
	}
	line.Color = plotter.DefaultLineStyle.Color
	points.Shape = plotter.DefaultGlyphStyle.Shape
	points.Color = plotter.DefaultGlyphStyle.Color
	offsetPlot.Add(line, points)

	// Save offset plot to a file
	if err := offsetPlot.Save(6*vg.Inch, 4*vg.Inch, "ntp_offset.png"); err != nil {
		return err
	}

	return nil
}

func (c *ntpClient) sendRequest() (float64, int64, error) {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", c.serverAddr, ntpPort))
	if err != nil {
		return 0, 0, err
	}
	defer conn.Close()

	req := &packet{Settings: 0x1B}
	req.OrigTimeSec, req.OrigTimeFrac = getTime()

	buffer := make([]byte, 48)
	binary.BigEndian.PutUint32(buffer[40:], req.OrigTimeSec)
	binary.BigEndian.PutUint32(buffer[44:], req.OrigTimeFrac)
	buffer[0] = req.Settings

	//fmt.Println("sending request", buffer)
	T1 := time.Now().Add(time.Duration(offset) * time.Millisecond).UTC()
	_, err = conn.Write(buffer)
	if err != nil {
		return 0, 0, err
	}

	responseBuffer := make([]byte, 48)
	err = conn.SetReadDeadline(time.Now().Add(timeoutDuration))
	if err != nil {
		return 0, 0, err
	}

	_, err = conn.Read(responseBuffer)
	if err != nil {
		return 0, 0, err // Return on error, including timeout
	}

	responsePacket := parsePacket(responseBuffer)
	//fmt.Println("packed received", responsePacket)
	// Calculate delay and offset

	T4 := time.Now().Add(time.Duration(offset) * time.Millisecond).UTC() // Client receive time
	T2 := toTime(responsePacket.RxTimeSec, responsePacket.RxTimeFrac)       // Server receive time
	T3 := toTime(responsePacket.TxTimeSec, responsePacket.TxTimeFrac)       // Server transmit time

	delay := (T4.Sub(T1) - T3.Sub(T2)).Seconds() * 1000 // in milliseconds
	newOffset := ((T2.Sub(T1) + T3.Sub(T4)).Seconds() / 2) * 1000 // in milliseconds
	fmt.Printf("T1: %v T2: %v T3: %v T4: %v\n", T1, T2, T3, T4)
	//fmt.Printf("Delay: %.2f ms\n", delay)
	//fmt.Printf("New Offset: %.2f ms\n", newOffset)

	return delay, int64(newOffset), nil
}

func getTimeInSeconds() uint32 {
	now := time.Now()
	return uint32(now.Unix())
}


func toTime(seconds uint32, fraction uint32) time.Time {
	secs := int64(seconds) - 2208988800 // NTP epoch to Unix epoch
	nanos := int64(fraction) * 1e9 >> 32

	return time.Unix(secs, nanos).UTC()
}

func parsePacket(buffer []byte) *packet {
	p := &packet{
		Settings:       buffer[0],
		Stratum:        buffer[1],
		Poll:           int8(buffer[2]),
		RootDelay:      binary.BigEndian.Uint32(buffer[4:8]),
		RootDispersion: binary.BigEndian.Uint32(buffer[8:12]),
		ReferenceID:    binary.BigEndian.Uint32(buffer[12:16]),
		Precision:      int8(buffer[3]),
		RefTimeSec:     binary.BigEndian.Uint32(buffer[16:20]),
		RefTimeFrac:    binary.BigEndian.Uint32(buffer[20:24]),
		OrigTimeSec:    binary.BigEndian.Uint32(buffer[24:28]),
		OrigTimeFrac:   binary.BigEndian.Uint32(buffer[28:32]),
		RxTimeSec:      binary.BigEndian.Uint32(buffer[32:36]),
		RxTimeFrac:     binary.BigEndian.Uint32(buffer[36:40]),
		TxTimeSec:      binary.BigEndian.Uint32(buffer[40:44]),
		TxTimeFrac:     binary.BigEndian.Uint32(buffer[44:48]),
	}
	return p
}

func getTime() (uint32, uint32) {
	now := time.Now()
	secs := uint32(now.Unix()) + 2208988800 // NTP epoch
	nanos := uint32(now.Nanosecond())

	// Convert nanoseconds to 32-bit fraction
	fracs := uint32(uint64(nanos) * (1 << 32) / uint64(time.Second))

	return secs, fracs
}