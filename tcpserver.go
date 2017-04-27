package modbus

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

// tcpTransporter implements Transporter interface.
type TcpServer struct {
	// Connect string
	port     int
	packager tcpPackager
	// Connect & Read timeout
	Timeout time.Duration
	// Idle timeout to close the connection
	IdleTimeout time.Duration
	// Transmission logger
	Logger *log.Logger

	// TCP connection
	mu           sync.Mutex
	conn         net.Listener
	closeTimer   *time.Timer
	lastActivity time.Time
}

func NewTcpServer(port int) (*TcpServer, error) {
	var err error
	s := &TcpServer{port: port}
	s.conn, err = net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	return s, nil
}

type serverHandler func(pdu *PDUwithSlaveid) *PDUwithSlaveid

type mbHandler interface {
	ReadHoldingRegisters(slaveid byte, address, quantity uint16) ([]uint16, error)
	WriteSingleRegister(slaveid byte, address, value uint16) error
}

func flush(c net.Conn) (err error) {
	if err = c.SetReadDeadline(time.Now()); err != nil {
		return
	}
	// Timeout setting will be reset when reading
	if _, err = ioutil.ReadAll(c); err != nil {
		// Ignore timeout error
		if netError, ok := err.(net.Error); ok && netError.Timeout() {
			err = nil
		}
	}
	return
}

func encodeMbError(uintid, fc, ex byte) *PDUwithSlaveid {
	return &PDUwithSlaveid{uintid,
		ProtocolDataUnit{
			FunctionCode: (fc | 0x80),
			Data:         []byte{ex},
		}}
}

func (mb *TcpServer) ServeModbus(handler mbHandler) {
	mb.serve(func(pdu *PDUwithSlaveid) *PDUwithSlaveid {
		var respPdu *PDUwithSlaveid
		switch pdu.FunctionCode {
		case FuncCodeReadHoldingRegisters:
			address := binary.BigEndian.Uint16(pdu.Data)
			quantity := binary.BigEndian.Uint16(pdu.Data[2:])
			if data, err := handler.ReadHoldingRegisters(pdu.SlaveID, address, quantity); err != nil {
				respPdu = encodeMbError(pdu.SlaveID, pdu.FunctionCode, ExceptionCodeIllegalDataAddress)
			} else {
				lenbyte := []byte{byte(len(data) * 2)}

				respPdu = &PDUwithSlaveid{pdu.SlaveID,
					ProtocolDataUnit{
						FunctionCode: pdu.FunctionCode,
						Data:         append(lenbyte, dataBlock(data...)...),
					}}
			}
		case FuncCodeWriteSingleRegister:
			address := binary.BigEndian.Uint16(pdu.Data)
			value := binary.BigEndian.Uint16(pdu.Data[2:])
			if err := handler.WriteSingleRegister(pdu.SlaveID, address, value); err != nil {
				respPdu = encodeMbError(pdu.SlaveID, pdu.FunctionCode, ExceptionCodeIllegalDataAddress)
			} else {
				respPdu = pdu
			}
		default:
			respPdu = encodeMbError(pdu.SlaveID, pdu.FunctionCode, ExceptionCodeIllegalFunction)
		}
		return respPdu
	})
}
func (mb *TcpServer) serve(handler serverHandler) {
	for {
		if conn, err := mb.conn.Accept(); err == nil {
			go func(c net.Conn) {
				defer c.Close()
				var data [tcpMaxLength]byte
				for {
					// Read header first
					if _, err = io.ReadFull(c, data[:tcpHeaderSize]); err != nil {
						continue
					}
					transactionId := binary.BigEndian.Uint16(data[:])
					// Read length, ignore transaction & protocol id (4 bytes)
					length := int(binary.BigEndian.Uint16(data[4:]))
					if length <= 0 {
						flush(c)
						return
					}
					if length > (tcpMaxLength - (tcpHeaderSize - 1)) {
						flush(c)
						err = fmt.Errorf("modbus: length in response header '%v' must not greater than '%v'", length, tcpMaxLength-tcpHeaderSize+1)
						return
					}
					// Skip unit id
					length += tcpHeaderSize - 1
					if _, err = io.ReadFull(c, data[tcpHeaderSize:length]); err != nil {
						return
					}

					aduRequest := data[:length]
					pdu, err := mb.packager.Decode(aduRequest)
					if err != nil {
						continue
					}
					resp := handler(pdu)
					adu := make([]byte, tcpHeaderSize+1+len(resp.Data))

					// Transaction identifier
					binary.BigEndian.PutUint16(adu, uint16(transactionId))
					// Protocol identifier
					binary.BigEndian.PutUint16(adu[2:], tcpProtocolIdentifier)
					// Length = sizeof(SlaveId) + sizeof(FunctionCode) + Data
					binary.BigEndian.PutUint16(adu[4:], uint16(1+1+len(resp.Data)))
					// Unit identifier
					adu[6] = resp.SlaveID
					// PDU
					adu[tcpHeaderSize] = resp.FunctionCode
					copy(adu[tcpHeaderSize+1:], resp.Data)
					c.Write(adu)
				}
			}(conn)
		}

	}
}

func (mb *TcpServer) logf(format string, v ...interface{}) {
	if mb.Logger != nil {
		mb.Logger.Printf(format, v...)
	}
}

// closeLocked closes current connection. Caller must hold the mutex before calling this method.
func (mb *TcpServer) Close() (err error) {
	if mb.conn != nil {
		err = mb.conn.Close()
		mb.conn = nil
	}
	return
}
