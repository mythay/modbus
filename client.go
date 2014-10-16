// Copyright 2014 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license.  See the LICENSE file for details.
package modbus

import (
	"encoding/binary"
)

type client struct {
	encoder     Encoder
	decoder     Decoder
	transporter Transporter
}

// Request:
//  Function code         : 1 byte (0x01)
//  Starting address      : 2 bytes
//  Quantity of coils     : 2 bytes
// Response:
//  Function code         : 1 byte (0x01)
//  Byte count            : 1 byte
//  Coil status           : N* bytes (=N or N+1)
func (mb *client) ReadCoils(address, quantity uint16) (results []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadCoils,
		Data:         dataBlock(address, quantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	if count != (len(response.Data) - 1) {
		err = ErrResponseSize
		return
	}
	results = response.Data[1:]
	return
}

// Request:
//  Function code         : 1 byte (0x02)
//  Starting address      : 2 bytes
//  Quantity of inputs    : 2 bytes
// Response:
//  Function code         : 1 byte (0x02)
//  Byte count            : 1 byte
//  Input status          : N* bytes (=N or N+1)
func (mb *client) ReadDiscreteInputs(address, quantity uint16) (results []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadDiscreteInputs,
		Data:         dataBlock(address, quantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	if count != (len(response.Data) - 1) {
		err = ErrResponseSize
		return
	}
	results = response.Data[1:]
	return
}

// Request:
//  Function code         : 1 byte (0x03)
//  Starting address      : 2 bytes
//  Quantity of registers : 2 bytes
// Response:
//  Function code         : 1 byte (0x03)
//  Byte count            : 1 byte
//  Register value        : Nx2 bytes
func (mb *client) ReadHoldingRegisters(address, quantity uint16) (results []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadHoldingRegisters,
		Data:         dataBlock(address, quantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	if count != (len(response.Data) - 1) {
		err = ErrResponseSize
		return
	}
	results = response.Data[1:]
	return
}

// Request:
//  Function code         : 1 byte (0x04)
//  Starting address      : 2 bytes
//  Quantity of registers : 2 bytes
// Response:
//  Function code         : 1 byte (0x04)
//  Byte count            : 1 byte
//  Input registers       : N bytes
func (mb *client) ReadInputRegisters(address, quantity uint16) (results []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadInputRegisters,
		Data:         dataBlock(address, quantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	if count != (len(response.Data) - 1) {
		err = ErrResponseSize
		return
	}
	results = response.Data[1:]
	return
}

// Request:
//  Function code         : 1 byte (0x05)
//  Output address        : 2 bytes
//  Output value          : 2 bytes
// Response:
//  Function code         : 1 byte (0x05)
//  Output address        : 2 bytes
//  Output value          : 2 bytes
func (mb *client) WriteSingleCoil(address, value uint16) (results []byte, err error) {
	// The requested ON/OFF state can only be 0xFF00 and 0x0000
	if value != 0xFF00 && value != 0x0000 {
		err = ErrBadRequest
		return
	}
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteSingleCoil,
		Data:         dataBlock(address, value),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 4 {
		err = ErrResponseSize
		return
	}
	if address != binary.BigEndian.Uint16(response.Data) {
		err = ErrResponse
		return
	}
	results = response.Data[2:]
	return
}

// Request:
//  Function code         : 1 byte (0x06)
//  Register address      : 2 bytes
//  Register value        : 2 bytes
// Response:
//  Function code         : 1 byte (0x06)
//  Register address      : 2 bytes
//  Register value        : 2 bytes
func (mb *client) WriteSingleRegister(address, value uint16) (results []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteSingleRegister,
		Data:         dataBlock(address, value),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 4 {
		err = ErrResponseSize
		return
	}
	if address != binary.BigEndian.Uint16(response.Data) {
		err = ErrResponse
		return
	}
	results = response.Data[2:]
	return
}

// Request:
//  Function code         : 1 byte (0x0F)
//  Starting address      : 2 bytes
//  Quantity of outputs   : 2 bytes
//  Byte count            : 1 byte
//  Outputs value         : N* bytes
// Response:
//  Function code         : 1 byte (0x0F)
//  Starting address      : 2 bytes
//  Quantity of outputs   : 2 bytes
func (mb *client) WriteMultipleCoils(address, quantity uint16, value []byte) (results []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteMultipleCoils,
		Data:         dataBlockSuffix(value, address, quantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 4 {
		err = ErrResponseSize
		return
	}
	if address != binary.BigEndian.Uint16(response.Data) {
		err = ErrResponse
		return
	}
	results = response.Data[2:]
	return
}

// Request:
//  Function code         : 1 byte (0x10)
//  Starting address      : 2 bytes
//  Quantity of outputs   : 2 bytes
//  Byte count            : 1 byte
//  Registers value       : N* bytes
// Response:
//  Function code         : 1 byte (0x10)
//  Starting address      : 2 bytes
//  Quantity of registers : 2 bytes
func (mb *client) WriteMultipleRegisters(address, quantity uint16, value []byte) (results []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteMultipleRegisters,
		Data:         dataBlockSuffix(value, address, quantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 4 {
		err = ErrResponseSize
		return
	}
	if address != binary.BigEndian.Uint16(response.Data) {
		err = ErrResponse
		return
	}
	results = response.Data[2:]
	return
}

// Request:
//  Function code         : 1 byte (0x16)
//  Reference address     : 2 bytes
//  AND-mask              : 2 bytes
//  OR-mask               : 2 bytes
// Response:
//  Function code         : 1 byte (0x16)
//  Reference address     : 2 bytes
//  AND-mask              : 2 bytes
//  OR-mask               : 2 bytes
func (mb *client) MaskWriteRegister(address, andMask, orMask uint16) (results []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeMaskWriteRegister,
		Data:         dataBlock(address, andMask, orMask),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	// Fixed response length
	if len(response.Data) != 6 {
		err = ErrResponseSize
		return
	}
	if address != binary.BigEndian.Uint16(response.Data) {
		err = ErrResponse
		return
	}
	if andMask != binary.BigEndian.Uint16(response.Data[2:]) {
		err = ErrResponse
		return
	}
	if orMask != binary.BigEndian.Uint16(response.Data[4:]) {
		err = ErrResponse
		return
	}
	results = response.Data[2:]
	return
}

// Request:
//  Function code         : 1 byte (0x17)
//  Read starting address : 2 bytes
//  Quantity to read      : 2 bytes
//  Write starting address: 2 bytes
//  Quantity to write     : 2 bytes
//  Write byte count      : 1 byte
//  Write registers value : N* bytes
// Response:
//  Function code         : 1 byte (0x17)
//  Byte count            : 2 bytes
//  Read registers value  : Nx2 bytes
func (mb *client) ReadWriteMultipleRegisters(readAddress, readQuantity, writeAddress, writeQuantity uint16, value []byte) (results []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadWriteMultipleRegisters,
		Data:         dataBlockSuffix(value, readAddress, readQuantity, writeAddress, writeQuantity),
	}
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	if count != (len(response.Data) - 1) {
		err = ErrResponse
		return
	}
	results = response.Data[1:]
	return
}

// Send a request and check possible exception in the response
func (mb *client) send(request *ProtocolDataUnit) (response *ProtocolDataUnit, err error) {
	aduRequest, err := mb.encoder.Encode(request)
	if err != nil {
		return
	}
	aduResponse, err := mb.transporter.Send(aduRequest)
	if err != nil {
		return
	}
	response, err = mb.decoder.Decode(aduResponse)
	if err != nil {
		return
	}
	// Check correct function code returned (exception)
	if response.FunctionCode != request.FunctionCode {
		err = responseError(response)
		return
	}
	if response.Data == nil || len(response.Data) == 0 {
		// Empty response
		err = ErrResponse
		return
	}
	return
}

// Creates a sequence of uint16 data
func dataBlock(value ...uint16) []byte {
	data := make([]byte, 2*len(value))
	for i, v := range value {
		binary.BigEndian.PutUint16(data[i*2:], v)
	}
	return data
}

// Creates a sequence of uint16 data and append the suffix plus its length
func dataBlockSuffix(suffix []byte, value ...uint16) []byte {
	length := 2*len(value)
	data := make([]byte, length+1+len(suffix))
	for i, v := range value {
		binary.BigEndian.PutUint16(data[i*2:], v)
	}
	data[length] = uint8(len(suffix))
	copy(data[length+1:], suffix)
	return data
}

func responseError(response *ProtocolDataUnit) error {
	mbError := &ModbusError{FunctionCode: response.FunctionCode}
	if response.Data != nil && len(response.Data) > 0 {
		mbError.ExceptionCode = response.Data[0]
	}
	return mbError
}