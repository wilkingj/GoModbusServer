package mbserver

import (
	"errors"
	"io"
	"log"
	"time"

	"github.com/goburrow/serial"
)

const max_PDU = 253
const max_ADU_RTU = 256
const min_ADU_RTU = 4
const max_ADU_TCP = 260

func (s *Server) ListenRTU(serialConfig *serial.Config) (err error) {
	port, err := serial.Open(serialConfig)
	if err != nil {
		log.Fatalf("failed to open %s: %v\n", serialConfig.Address, err)
	}
	s.ports = append(s.ports, port)
	go s.acceptSerialRequests(port)
	return err
}

func (s *Server) acceptSerialRequests(port io.ReadWriteCloser) {
	for {

		select {
		case <-s.closeChan:
			close(s.requestChan)
			return
		default:
			request, _, err := s.readRequests(port)

			t := time.Now()

			if err != nil {
				if err.Error() == "serial: timeout" {
					s.last = s.last[:0]
					continue // timeOut error is not an issue
				} else if err.Error() == "Unsupported Function in this Modbus-Library" {
					continue
				} else {
					log.Fatal(err) // this could be more sophisticated
				}
			}

			frame, err := NewRTUFrame(request)

			if err != nil {
				//discard all of the bytes of the requests until slaveID then save bytes[slaveId:] in s.last
				n, err := find(s.slaveId, request[1:])
				if err == nil {
					s.last = append(request[n:], s.last...)
				}
				log.Printf("bad serial frame error %v\n", err)
				continue
			}

			if frame.Address != s.slaveId {
				//Package is not for us so discard it; could check for this earlier ... ?!
				continue
			}

			s.requestChan <- &Request{port, frame, t}

		}

	}
}

func (s *Server) readRequests(reader io.Reader) ([]byte, int, error) {
	var err error
	req := make([]byte, max_ADU_RTU+3)
	expected := min_ADU_RTU
	read := 0

	if len(s.last) != 0 {
		n, err := find(s.slaveId, s.last)
		if err == nil {
			read += copy(req, s.last[n:])
		}
		s.last = s.last[:0]
	}

	expected, err = s.getRTUSizeFromHeader(req[:read])

	if err != nil {
		n, err2 := find(s.slaveId, req[1:read])
		if err2 == nil {
			s.last = req[n:read]
		}
		return nil, read, err
	}

	for read < expected {

		n, err := reader.Read(req[read:])
		read += n
		if err != nil {
			s.last = append(s.last, req[:read]...)
			return nil, read, err
		}

		expected, err = s.getRTUSizeFromHeader(req[:read])
		/*
			if err != nil {
				n, err := find(s.slaveId, s.last[1:])
				if err == nil {
					s.last = append(s.last, req[n:])
				}
				s.last = s.last[:0]
				return nil, read, err
			}
		*/
	}

	if read > expected {
		s.last = append(s.last[:0], req[expected:read]...)

		return req[:expected], expected, nil
	}

	return req[:read], read, nil
}

func (s *Server) getPDUSizeFromHeader(header []byte) (int, error) {

	fc := uint8(header[0])

	switch fc {
	case ReadCoils_fc, ReadDiscreteInput_fc, ReadHoldingRegisters_fc, ReadInputRegisters_fc, WriteSingleCoil_fc, WriteHoldingRegister_fc:
		return 5, nil

	case WriteMultipleCoils_fc, WriteHoldingRegisters_fc:
		if len(header) > 5 {
			return int(header[5]) + 6, nil
		}
		return max_PDU, nil

	case ReadExceptionStatus_fc, GetCommEventCounter_fc, GetCommEventLog_fc, ReportSlaveId_fc:
		return 1, nil

	case ReadFifoQueue_fc:
		return 3, nil

	case ReadDeviceIdentification_fc:
		return 4, nil

	case MaskWriteRegister_fc:
		return 7, nil

	case WriteFileRecord_fc:
		if len(header) < 2 {
			return max_PDU, nil
		}
		return int(header[1]) + 2, nil

	case ReadWriteMultipleRegisters_fc:
		if len(header) < 10 {
			return max_PDU, nil
		}
		return int(header[9]) + 10, nil

	default:
		return 0, errors.New("Unsupported Function in this Modbus-Library")

	}
}

//GetRTUSizeFromHeader returns the expected sized of a rtu packet with the given
//RTU header, if not enough info is in the header, then it returns the shortest possible.
func (s *Server) getRTUSizeFromHeader(header []byte) (x int, err error) {
	if len(header) < 2 {
		return max_PDU, nil
	}

	x, err = s.getPDUSizeFromHeader(header[1:])
	x += 3

	return
}

func find(x byte, data []byte) (uint, error) {

	if len(data) == 0 {
		return 0, errors.New("Empty Byte-Slice")
	}

	var n uint

	n = 0

	for n < uint(len(data)) {
		if data[n] == x {
			return n, nil
		}
		n++
	}

	return 0, errors.New("Slave Id not found in Server.last")
}
