// Copyright 2014 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license.  See the LICENSE file for details.

package test

import (
	"log"
	"os"
	"testing"

	"github.com/tarm/serial"

	"github.com/mythay/modbus"
)

const (
	asciiDevice = "/dev/pts/21"
)

func TestASCIIClient(t *testing.T) {
	// Diagslave does not support broadcast id.
	handler := modbus.NewASCIIClientHandler(asciiDevice)
	ClientTestAll(t, modbus.NewClient(handler))
}

func TestASCIIClientAdvancedUsage(t *testing.T) {
	handler := modbus.NewASCIIClientHandler(asciiDevice)
	handler.Baud = 19200
	handler.Size = 8
	handler.Parity = serial.ParityEven
	handler.StopBits = 1
	handler.Logger = log.New(os.Stdout, "ascii: ", log.LstdFlags)
	err := handler.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer handler.Close()

	client := modbus.NewClient(handler)
	results, err := client.ReadDiscreteInputs(12, 15, 2)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	results, err = client.ReadWriteMultipleRegisters(12, 0, 2, 2, 2, []byte{1, 2, 3, 4})
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
}
