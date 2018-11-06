package mbserver

import (
	"encoding/binary"
)

//constants for function codes

const (
	ReadCoils_fc             uint8 = 1
	ReadDiscreteInput_fc     uint8 = 2
	ReadHoldingRegisters_fc  uint8 = 3
	ReadInputRegisters_fc    uint8 = 4
	WriteSingleCoil_fc       uint8 = 5
	WriteHoldingRegister_fc  uint8 = 6
	WriteMultipleCoils_fc    uint8 = 15
	WriteHoldingRegisters_fc uint8 = 16

	ReadExceptionStatus_fc        uint8 = 7
	GetCommEventCounter_fc        uint8 = 11
	GetCommEventLog_fc            uint8 = 12
	ReportSlaveId_fc              uint8 = 17
	ReadFileRecord_fc             uint8 = 20
	WriteFileRecord_fc            uint8 = 21
	MaskWriteRegister_fc          uint8 = 22
	ReadWriteMultipleRegisters_fc uint8 = 23
	ReadFifoQueue_fc              uint8 = 24
	ReadDeviceIdentification_fc   uint8 = 43
)

func functionCodeToString(fc uint8) string {
	switch fc {
	case ReadCoils_fc:
		return "readCoils"
	case ReadDiscreteInput_fc:
		return "readDiscreteInput"
	case ReadHoldingRegisters_fc:
		return "readHoldingRegisters"
	case ReadInputRegisters_fc:
		return "readInputRegisters"
	case WriteSingleCoil_fc:
		return "writeSingleCoil"
	case WriteHoldingRegister_fc:
		return "writeHoldingRegister"
	case WriteMultipleCoils_fc:
		return "writeMultipleCoils"
	case WriteHoldingRegisters_fc:
		return "writeHoldingRegisters"
	case ReadExceptionStatus_fc:
		return "readExceptionStatus"
	case GetCommEventCounter_fc:
		return "getCommEventCounter"
	case GetCommEventLog_fc:
		return "getCommEventLog"
	case ReportSlaveId_fc:
		return "reportSlaveID"
	case ReadFileRecord_fc:
		return "readFileRecord"
	case WriteFileRecord_fc:
		return "writeFileRecord"
	case MaskWriteRegister_fc:
		return "maskWriteRegister"
	case ReadWriteMultipleRegisters_fc:
		return "readWriteMultipleRegisters"
	case ReadFifoQueue_fc:
		return "readFifoQueue"
	case ReadDeviceIdentification_fc:
		return "readDeviceIdentification"

	default:
		return "Unkown Modbus Function"

	}
}

// ReadCoils function 1, reads coils from internal memory.
func ReadCoils(s *Server, frame Framer) ([]byte, *Exception) {
	register, numRegs, endRegister := RegisterAddressAndNumber(frame)
	if endRegister > len(s.Coils) {
		return []byte{}, &IllegalDataAddress
	}
	dataSize := numRegs / 8
	if (numRegs % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)
	for i, value := range s.Coils[register:endRegister] {
		if value != 0 {
			shift := uint(i) % 8
			data[1+i/8] |= byte(1 << shift)
		}
	}
	return data, &Success
}

// ReadDiscreteInputs function 2, reads discrete inputs from internal memory.
func ReadDiscreteInputs(s *Server, frame Framer) ([]byte, *Exception) {
	register, numRegs, endRegister := RegisterAddressAndNumber(frame)
	if endRegister > len(s.DiscreteInputs) {
		return []byte{}, &IllegalDataAddress
	}
	dataSize := numRegs / 8
	if (numRegs % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)
	for i, value := range s.DiscreteInputs[register:endRegister] {
		if value != 0 {
			shift := uint(i) % 8
			data[1+i/8] |= byte(1 << shift)
		}
	}
	return data, &Success
}

// ReadHoldingRegisters function 3, reads holding registers from internal memory.
func ReadHoldingRegisters(s *Server, frame Framer) ([]byte, *Exception) {
	register, numRegs, endRegister := RegisterAddressAndNumber(frame)
	if endRegister*2 > len(s.HoldingRegisters) {
		return []byte{}, &IllegalDataAddress
	}

	return append([]byte{byte(numRegs * 2)}, s.HoldingRegisters[register*2:endRegister*2]...), &Success
}

// ReadInputRegisters function 4, reads input registers from internal memory.
func ReadInputRegisters(s *Server, frame Framer) ([]byte, *Exception) {
	register, numRegs, endRegister := RegisterAddressAndNumber(frame)
	if endRegister*2 > len(s.InputRegisters) {
		return []byte{}, &IllegalDataAddress
	}
	return append([]byte{byte(numRegs * 2)}, s.InputRegisters[register*2:endRegister*2]...), &Success
}

// WriteSingleCoil function 5, write a coil to internal memory.
func WriteSingleCoil(s *Server, frame Framer) ([]byte, *Exception) {
	register, value := RegisterAddressAndValue(frame)
	// TODO Should we use 0 for off and 65,280 (FF00 in hexadecimal) for on?

	if !(len(s.Coils) > register) {
		return []byte{}, &IllegalDataAddress
	}

	if value == 0 || value == 65535 {
		if value == 0 {
			s.Coils[register] = byte(0)
		} else {
			s.Coils[register] = byte(1)
		}
		return frame.GetData()[0:4], &Success
	}

	return []byte{}, &IllegalDataValue
}

// WriteHoldingRegister function 6, write a holding register to internal memory.
func WriteHoldingRegister(s *Server, frame Framer) ([]byte, *Exception) {
	register, value := RegisterAddressAndValue(frame)

	if (register+1)*2 > len(s.HoldingRegisters) {
		return []byte{}, &IllegalDataAddress
	}

	input := []uint16{value}
	data := Uint16ToBytes(input)
	copy(s.HoldingRegisters[register*2:(register+1)*2], data)
	return frame.GetData()[0:4], &Success
}

// WriteMultipleCoils function 15, writes holding registers to internal memory.
func WriteMultipleCoils(s *Server, frame Framer) ([]byte, *Exception) {
	register, numRegs, endRegister := RegisterAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]

	if endRegister > len(s.Coils) {
		return []byte{}, &IllegalDataAddress
	}

	// TODO This is not correct, bits and bytes do not always align
	//if len(valueBytes)/2 != numRegs {
	//	return []byte{}, &IllegalDataAddress
	//}

	bitCount := 0
	for i, value := range valueBytes {
		for bitPos := uint(0); bitPos < 8; bitPos++ {
			s.Coils[register+(i*8)+int(bitPos)] = bitAtPosition(value, bitPos)
			bitCount++
			if bitCount >= numRegs {
				break
			}
		}
		if bitCount >= numRegs {
			break
		}
	}

	return frame.GetData()[0:4], &Success
}

// WriteHoldingRegisters function 16, writes holding registers to internal memory.
func WriteHoldingRegisters(s *Server, frame Framer) ([]byte, *Exception) {
	register, numRegs, endRegister := RegisterAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]
	var exception *Exception
	var data []byte

	if len(valueBytes)/2 != numRegs {
		return []byte{}, &IllegalDataAddress
	}

	if endRegister*2 > len(s.HoldingRegisters) {
		return []byte{}, &IllegalDataAddress
	}

	valuesUpdated := copy(s.HoldingRegisters[2*register:], valueBytes)
	if valuesUpdated == numRegs*2 {
		exception = &Success
		data = frame.GetData()[0:4]
	} else {
		exception = &IllegalDataAddress
	}

	return data, exception
}

// BytesToUint16 converts a big endian array of bytes to an array of unit16s
func BytesToUint16(bytes []byte) []uint16 {
	values := make([]uint16, len(bytes)/2)

	for i := range values {
		values[i] = binary.BigEndian.Uint16(bytes[i*2 : (i+1)*2])
	}
	return values
}

// Uint16ToBytes converts an array of uint16s to a big endian array of bytes
func Uint16ToBytes(values []uint16) []byte {
	bytes := make([]byte, len(values)*2)

	for i, value := range values {
		binary.BigEndian.PutUint16(bytes[i*2:(i+1)*2], value)
	}
	return bytes
}

func bitAtPosition(value uint8, pos uint) uint8 {
	return (value >> pos) & 0x01
}
