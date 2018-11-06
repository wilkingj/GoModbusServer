package mbserver

import (
	"io"
	"log"
	"net"
	"strings"
	"time"
)

func (s *Server) accept(listen net.Listener) error {
	for {
		select {
		case <-s.closeChan:
			close(s.requestChan)
			return nil
		default:

			conn, err := listen.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return nil
				}
				log.Printf("Unable to accept connections: %#v\n", err)
				return err
			}

			go func(conn net.Conn) {
				defer conn.Close()

				for {

					select {
					case <-s.closeChan:
						return
					default:

						packet := make([]byte, 512)
						bytesRead, err := conn.Read(packet)
						t := time.Now()
						if err != nil {
							if err != io.EOF {
								log.Printf("read error %v\n", err)
							}
							return
						}
						// Set the length of the packet to the number of read bytes.
						packet = packet[:bytesRead]

						frame, err := NewTCPFrame(packet)
						if err != nil {
							log.Printf("bad packet error %v\n", err)
							return
						}

						request := &Request{conn, frame, t}

						s.requestChan <- request
					}
				}
			}(conn)
		}
	}

}

// ListenTCP starts the Modbus server listening on "address:port".
func (s *Server) ListenTCP(addressPort string) (err error) {
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		log.Printf("Failed to Listen: %v\n", err)
		return err
	}
	s.listeners = append(s.listeners, listen)
	go s.accept(listen)
	return err
}
