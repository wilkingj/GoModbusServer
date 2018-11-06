package mbserver

import (
	"testing"
)

// Function 1
func TestReadCoils(t *testing.T) {
	s, _ := NewServer(255)

	s.Coils = make([]byte, 65535)

	// Set the coil values
	s.Coils[10] = 1
	s.Coils[11] = 1
	s.Coils[17] = 1
	s.Coils[18] = 1

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 6
	frame.Device = 255
	frame.Function = 1
	SetDataWithRegisterAndNumber(&frame, 10, 9)

	var req Request
	req.frame = &frame
	response := s.handle(&req)

	exception := GetException(response)
	if exception != Success {
		t.Errorf("expected Success, got %v", exception.String())
		t.FailNow()
	}
	// 2 bytes, 0b1000011, 0b00000001
	expect := []byte{2, 131, 1}
	got := response.GetData()
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}
}

//all of these Tests need to get adjusted !!!!!!!!!!!!!!!!!!!!!

// Function 2
func TestReadDiscreteInputs(t *testing.T) {
	s, _ := NewServer(255)
	s.DiscreteInputs = make([]byte, 500)
	// Set the discrete input values
	s.DiscreteInputs[0] = 1
	s.DiscreteInputs[1] = 0

	s.DiscreteInputs[6] = 0
	s.DiscreteInputs[7] = 1

	s.DiscreteInputs[8] = 1
	s.DiscreteInputs[9] = 1

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 6
	frame.Device = 255
	frame.Function = 2
	SetDataWithRegisterAndNumber(&frame, 0, 10)

	var req Request
	req.frame = &frame
	response := s.handle(&req)

	exception := GetException(response)
	if exception != Success {
		t.Errorf("expected Success, got %v", exception.String())
		t.FailNow()
	}
	expect := []byte{2, 129, 3}
	got := response.GetData()
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}
}

// Function 3
func TestReadHoldingRegisters(t *testing.T) {
	s, _ := NewServer(255)
	s.HoldingRegisters = make([]byte, 500)
	s.HoldingRegisters[200] = 0
	s.HoldingRegisters[201] = 1

	s.HoldingRegisters[202] = 0
	s.HoldingRegisters[203] = 2

	s.HoldingRegisters[204] = 0xff
	s.HoldingRegisters[205] = 0xff

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 6
	frame.Device = 255
	frame.Function = 3
	SetDataWithRegisterAndNumber(&frame, 100, 3)

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	if exception != Success {
		t.Errorf("expected Success, got %v", exception.String())
		t.FailNow()
	}
	expect := []byte{6, 0, 1, 0, 2, 255, 255}
	got := response.GetData()
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}
}

// Function 4
func TestReadInputRegisters(t *testing.T) {
	s, _ := NewServer(255)
	s.InputRegisters = make([]byte, 500)
	s.InputRegisters[400] = 0
	s.InputRegisters[401] = 1
	s.InputRegisters[402] = 0
	s.InputRegisters[403] = 2
	s.InputRegisters[404] = 0xff
	s.InputRegisters[405] = 0xff

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 6
	frame.Device = 255
	frame.Function = 4
	SetDataWithRegisterAndNumber(&frame, 200, 3)

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	if exception != Success {
		t.Errorf("expected Success, got %v", exception.String())
		t.FailNow()
	}
	expect := []byte{6, 0, 1, 0, 2, 255, 255}
	got := response.GetData()
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}
}

// Function 5
func TestWriteSingleCoil(t *testing.T) {
	s, _ := NewServer(255)
	s.Coils = make([]byte, 100000)

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 12
	frame.Device = 255
	frame.Function = 5
	//SetDataWithRegisterAndNumber(&frame, 65535, 1024)

	SetDataWithRegisterAndNumber(&frame, 65535, 0xffff)

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	if exception != Success {
		t.Errorf("expected Success, got %v", exception.String())
		t.FailNow()
	}
	expect := 1
	got := s.Coils[65535]
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v\n", expect, got)
	}
}

// Function 6
func TestWriteHoldingRegister(t *testing.T) {
	s, _ := NewServer(255)
	s.HoldingRegisters = make([]byte, 100)

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 12
	frame.Device = 255
	frame.Function = 6
	SetDataWithRegisterAndNumber(&frame, 5, 6)

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	if exception != Success {
		t.Errorf("expected Success, got %v", exception.String())
		t.FailNow()
	}
	expect := 6
	got := s.HoldingRegisters[11]
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v\n", expect, got)
	}
}

// Function 15
func TestWriteMultipleCoils(t *testing.T) {
	s, _ := NewServer(255)
	s.Coils = make([]byte, 500)

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 12
	frame.Device = 255
	frame.Function = 15
	SetDataWithRegisterAndNumberAndBytes(&frame, 1, 2, []byte{3})

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	if exception != Success {
		t.Errorf("expected Success, got %v", exception.String())
		t.FailNow()
	}
	expect := []byte{1, 1}
	got := s.Coils[1:3]
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v\n", expect, got)
	}
}

// Function 16
func TestWriteHoldingRegisters(t *testing.T) {
	s, _ := NewServer(255)
	s.HoldingRegisters = make([]byte, 500)

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 12
	frame.Device = 255
	frame.Function = 16
	SetDataWithRegisterAndNumberAndValues(&frame, 1, 2, []uint16{3, 4})

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	if exception != Success {
		t.Errorf("expected Success, got %v", exception.String())
		t.FailNow()
	}
	expect := []byte{0, 3, 0, 4}
	got := s.HoldingRegisters[2:6]
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v\n", expect, got)
	}
}

func TestBytesToUint16(t *testing.T) {
	bytes := []byte{1, 2, 3, 4}
	got := BytesToUint16(bytes)
	expect := []uint16{258, 772}
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v\n", expect, got)
	}
}

func TestUint16ToBytes(t *testing.T) {
	values := []uint16{1, 2, 3}
	got := Uint16ToBytes(values)
	expect := []byte{0, 1, 0, 2, 0, 3}
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v\n", expect, got)
	}
}

func TestOutOfBounds(t *testing.T) {
	s, _ := NewServer(255)

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 6
	frame.Device = 255

	var req Request
	req.frame = &frame

	// bits
	SetDataWithRegisterAndNumber(&frame, 65535, 2)

	frame.Function = 1
	response := s.handle(&req)
	exception := GetException(response)
	if exception != IllegalDataAddress {
		t.Errorf("expected IllegalDataAddress, got %v", exception.String())
	}

	frame.Function = 2
	response = s.handle(&req)
	exception = GetException(response)
	if exception != IllegalDataAddress {
		t.Errorf("expected IllegalDataAddress, got %v", exception.String())
	}

	SetDataWithRegisterAndNumberAndBytes(&frame, 65535, 2, []byte{3})
	frame.Function = 15
	response = s.handle(&req)
	exception = GetException(response)
	if exception != IllegalDataAddress {
		t.Errorf("expected IllegalDataAddress, got %v", exception.String())
	}

	// registers
	SetDataWithRegisterAndNumber(&frame, 65535, 2)

	frame.Function = 3
	response = s.handle(&req)
	exception = GetException(response)
	if exception != IllegalDataAddress {
		t.Errorf("expected IllegalDataAddress, got %v", exception.String())
	}

	frame.Function = 4
	response = s.handle(&req)
	exception = GetException(response)
	if exception != IllegalDataAddress {
		t.Errorf("expected IllegalDataAddress, got %v", exception.String())
	}

	SetDataWithRegisterAndNumberAndValues(&frame, 65535, 2, []uint16{0, 0})
	frame.Function = 16
	response = s.handle(&req)
	exception = GetException(response)
	if exception != IllegalDataAddress {
		t.Errorf("expected IllegalDataAddress, got %v", exception.String())
	}
}
