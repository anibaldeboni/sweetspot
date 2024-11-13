package main

import (
	"crypto/rand"
	"strconv"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func main() {
	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	// Start scanning.
	println("scanning...")
	devices := make(map[string]bluetooth.ScanResult)
	err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if _, exists := devices[device.Address.String()]; !exists {
			println("found device:", device.Address.String(), device.RSSI, device.LocalName())
		}
		devices[device.Address.String()] = device

	})
	must("start scan", err)
}

func generateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func pingDevice(peripheral *bluetooth.Device, payloadLength int) error {
	// Generate random payload
	payload, err := generateRandomBytes(payloadLength)
	if err != nil {
		return err
	}

	// Send the payload (ping)
	println("sending ping...")
	serv, err := peripheral.DiscoverServices([]bluetooth.UUID{})
	for _, s := range serv {
		println("service:", s.UUID)
		chars, err := s.DiscoverCharacteristics([]bluetooth.UUID{})
		for _, c := range chars {
			c.Write(payload)
			println("characteristic:", c.UUID)
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func ping(macAddr string) {
	mac := str2ba(macAddr) // YOUR BLUETOOTH MAC ADDRESS HERE

	fd, err := unix.Socket(syscall.AF_BLUETOOTH, syscall.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	check(err)
	addr := &unix.SockaddrRFCOMM{Addr: mac, Channel: 1}

	var data = make([]byte, 50)
	logrus.Print("connecting...")
	err = unix.Connect(fd, addr)
	check(err)
	defer unix.Close(fd)
	logrus.Println("done")

	for {
		n, err := unix.Read(fd, data)
		check(err)
		if n > 0 {
			logrus.Infof("Received: %v\n", string(data[:n]))
		}
	}
}

func check(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}

// str2ba converts MAC address string representation to little-endian byte array
func str2ba(addr string) [6]byte {
	a := strings.Split(addr, ":")
	var b [6]byte
	for i, tmp := range a {
		u, _ := strconv.ParseUint(tmp, 16, 8)
		b[len(b)-1-i] = byte(u)
	}
	return b
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
