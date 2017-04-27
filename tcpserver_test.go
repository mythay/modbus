package modbus

import "testing"
import "fmt"

type h struct {
	data []uint16
}

func (s *h) ReadHoldingRegisters(slaveid byte, address, quantity uint16) ([]uint16, error) {
	var buf []uint16
	if address < uint16(len(s.data)) && address+quantity-1 < uint16(len(s.data)) {
		buf = append(buf, s.data[address:address+quantity]...)
		return buf, nil
	}
	return nil, fmt.Errorf("out of range")
}
func (s *h) WriteSingleRegister(slaveid byte, address, value uint16) error {
	if address < uint16(len(s.data)) {
		s.data[address] = value
		return nil
	}
	return fmt.Errorf("out of range")
}
func Test_tcpServer_Serve(t *testing.T) {

	// t.Run("simple modbus server", func(t *testing.T) {
	// 	p, _ := NewTcpServer(502)
	// 	p.serve(func(pdu *PDUwithSlaveid) *PDUwithSlaveid {
	// 		fmt.Println(pdu)
	// 		pdu.Data[0] = 2
	// 		pdu.Data[1] = 10
	// 		pdu.Data[2] = 10
	// 		pdu.Data = pdu.Data[:3]
	// 		return pdu
	// 	})
	// })

	t.Run("memory modbus server", func(t *testing.T) {

		p, _ := NewTcpServer(502)
		p.ServeModbus(&h{[]uint16{10001, 2, 3, 4, 5, 6, 7, 8, 9, 10}})
	})

}
