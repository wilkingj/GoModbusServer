package mbserver

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/TheCount/modbus"
)

func isEqual(a interface{}, b interface{}) bool {
	expect, _ := json.Marshal(a)
	got, _ := json.Marshal(b)
	if string(expect) != string(got) {
		return false
	}
	return true
}

//works
func TestAduRegisterAndNumber(t *testing.T) {
	var frame TCPFrame
	SetDataWithRegisterAndNumber(&frame, 0, 64)

	expect := []byte{0, 0, 0, 64}
	got := frame.Data
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}
}

func TestAduSetDataWithRegisterAndNumberAndValues(t *testing.T) {
	var frame TCPFrame
	SetDataWithRegisterAndNumberAndValues(&frame, 7, 2, []uint16{3, 4})

	expect := []byte{0, 7, 0, 2, 4, 0, 3, 0, 4}
	got := frame.Data
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}
}

func TestUnsupportedFunction(t *testing.T) {
	s, _ := NewServer(255)
	var frame TCPFrame
	frame.Function = 255

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	if exception != IllegalFunction {
		t.Errorf("expected IllegalFunction (%d), got (%v)", IllegalFunction, exception)
	}
}

func TestModbus(t *testing.T) {
	// Server
	s, _ := NewServer(255)

	s.InputRegisters = make([]byte, 65535)
	s.HoldingRegisters = make([]byte, 65535)
	s.Coils = make([]byte, 65535)
	s.DiscreteInputs = make([]byte, 65535)

	err := s.ListenTCP("127.0.0.1:3333")
	if err != nil {
		t.Fatalf("failed to listen, got %v\n", err)
	}
	defer s.Close()

	// Allow the server to start and to avoid a connection refused on the client
	time.Sleep(1 * time.Millisecond)

	// Client
	handler := modbus.NewTCPClientHandler("127.0.0.1:3333")
	handler.SlaveId = 255
	// Connect manually so that multiple requests are handled in one connection session
	err = handler.Connect()
	if err != nil {
		t.Errorf("failed to connect, got %v\n", err)
		t.FailNow()
	}
	defer handler.Close()
	client := modbus.NewClient(handler)

	// Coils
	results, err := client.WriteMultipleCoils(100, 9, []byte{255, 1})
	if err != nil {
		t.Errorf("WriteMultipleCoils expected nil, got %v\n", err)
		t.FailNow()
	}

	results, err = client.ReadCoils(100, 16)
	if err != nil {
		t.Errorf("ReadCoils expected nil, got %v\n", err)
		t.FailNow()
	}
	expect := []byte{255, 1}
	got := results
	if !isEqual(expect, got) {
		t.Errorf("ReadCoils expected %v, got %v", expect, got)
	}

	// Discrete inputs
	results, err = client.ReadDiscreteInputs(0, 64)
	if err != nil {
		t.Errorf("ReadDiscreteInputs expected nil, got %v\n", err)
		t.FailNow()
	}
	// test: 2017/05/14 21:09:53 modbus: sending 00 01 00 00 00 06 ff 02 00 00 00 40
	// test: 2017/05/14 21:09:53 modbus: received 00 01 00 00 00 0b ff 02 08 00 00 00 00 00 00 00 00
	expect = []byte{0, 0, 0, 0, 0, 0, 0, 0}
	got = results
	if !isEqual(expect, got) {
		t.Errorf("ReadDiscreteInputs expected %v, got %v", expect, got)
	}

	//function Code : 16
	// Holding registers
	results, err = client.WriteMultipleRegisters(1, 2, []byte{0, 3, 0, 4})
	if err != nil {
		t.Errorf("WriteMultipleRegisters expected nil, got %v\n", err)
		t.FailNow()
	}
	// received: 00 01 00 00 00 06 ff 10 00 01 00 02
	expect = []byte{0, 2}
	got = results
	if !isEqual(expect, got) {
		t.Errorf("WriteMultipleRegisters expected %v, got %v", expect, got)
	}

	results, err = client.ReadHoldingRegisters(1, 2)
	if err != nil {
		t.Errorf("ReadHoldingRegisters expected nil, got %v\n", err)
		t.FailNow()
	}
	expect = []byte{0, 3, 0, 4}
	got = results
	if !isEqual(expect, got) {
		t.Errorf("ReadHoldingRegisters expected %v, got %v", expect, got)
	}

	// Input registers
	s.InputRegisters[2] = 0
	s.InputRegisters[3] = 0xff

	s.InputRegisters[4] = 0
	s.InputRegisters[5] = 1

	results, err = client.ReadInputRegisters(1, 2)
	if err != nil {
		t.Errorf("ReadInputRegisters expected nil, got %v\n", err)
		t.FailNow()
	}
	expect = []byte{0, 0xff, 0, 1}
	got = results
	if !isEqual(expect, got) {
		t.Errorf("ReadInputRegisters expected %v, got %v", expect, got)
	}
}

func TestReadRequests(t *testing.T) {

	serv, _ := NewServer(255)

	//create new Buff-Reader with the desired "Read" in the constructor

	//Test 1 :: Valid Request and clear Buffer
	data := []byte{255, 3, 0, 12, 0, 2, 127, 128}
	buffReader := bytes.NewBuffer(data)

	req, _, err := serv.readRequests(buffReader)

	if err != nil {
		t.Errorf("Test1 :expected nil, got %v\n", err)
	}

	if !isEqual(data, req) {
		t.Errorf("Test1 :expected %v got %v\n", data, req)
	}

	//Valid Request but split in Buffer and Read
	serv.last = []byte{255}

	data = []byte{3, 0, 12, 0, 2, 127, 128}
	buffReader.Write(data)

	req, _, err = serv.readRequests(buffReader)

	if err != nil {
		t.Errorf("Test2 :expected nil, got %v\n", err)
	}

	data = append([]byte{255}, data...)
	if !isEqual(data, req) {
		t.Errorf("Test2 :expected %v got %v\n", data, req)
	}

	//invalid buffer without Slave-Id and then valid Request
	serv.last = []byte{2, 36, 36, 99}

	data = []byte{255, 3, 0, 12, 0, 2, 127, 128}
	buffReader.Write(data)

	req, _, err = serv.readRequests(buffReader)

	if err != nil {
		t.Errorf("Test3 :expected nil, got %v\n", err)
	}

	if !isEqual(data, req) {
		t.Errorf("Test3 :expected %v got %v\n", data, req)
	}

	//invalid buffer with Slave-Id and then valid Request
	serv.last = []byte{2, 255, 36, 99}

	data = []byte{255, 3, 0, 12, 0, 2, 127, 128}
	buffReader.Write(data)

	req, _, err = serv.readRequests(buffReader)

	if err == nil {
		t.Errorf("Test3 :expected non-nil, got %v\n", err)
	}

	req, _, err = serv.readRequests(buffReader)

	if err != nil {
		t.Errorf("Test3 :expected nil, got %v\n", err)
	}

	if !isEqual(data, req) {
		t.Errorf("Test3 :expected %v got %v\n", data, req)
	}

	//invalid buffer with Slave-Id and then valid Request
	serv.last = []byte{2, 255, 36, 99, 255, 3, 0, 12, 0, 2, 127, 128}

	//data = []byte{255, 3, 0, 12, 0, 2, 127, 128}
	//buffReader.Write(data)

	req, _, err = serv.readRequests(buffReader)

	if err == nil {
		t.Errorf("Test4 :expected non-nil, got %v\n req : %v", err, req)
	}

	req, _, err = serv.readRequests(buffReader)

	if err != nil {
		t.Errorf("Test4 :expected nil, got %v\n", err)
	}

	if !isEqual(data, req) {
		t.Errorf("Test4 :expected %v got %v\n", data, req)
	}

	//Valid Request but split in Buffer and Read
	serv.last = []byte{255}

	data = []byte{3, 0, 12, 0, 2, 127, 128, 255, 3, 0, 12, 0, 2, 127, 128}
	buffReader.Write(data)

	req, n, err := serv.readRequests(buffReader)

	if err != nil {
		t.Errorf("Test5 :expected nil, got %v\n", err)
	}

	data = append([]byte{255}, data...)
	if !isEqual(data[:n], req) {
		t.Errorf("Test5 :expected %v got %v\n", data[:n], req)
	}

	if tail := []byte{255, 3, 0, 12, 0, 2, 127, 128}; !isEqual(serv.last, tail) {
		t.Errorf("Test 5: s.last expected %v got %v\n", tail, serv.last)
	}

	//test readwriteMultipleRegisters

	data = []byte{255, 23, 0, 3, 0, 6, 0, 12, 0, 3, 6, 0, 200, 0, 200, 0, 200, 127, 128}
	buffReader.Write(data)

	req, _, err = serv.readRequests(buffReader)

	if !isEqual(req, []byte{255, 3, 0, 12, 0, 2, 127, 128}) {
		t.Errorf("Test 6 : req expectet %v got %v\n", data, req)
	}

	req, _, err = serv.readRequests(buffReader)

	if !isEqual(req, data) {
		t.Errorf("Test 6 : req expectet %v got %v\n", data, req)
	}

}
