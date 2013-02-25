package Listener

// Imports
import (
	"net"
	"log"
	"os"
)

//External Globals
var PacketsReceived = 0
var LifetimePacketsReceived = 0

//Internal Globals

type PacketType int
const (
    RFC3164Packet PacketType  = iota + 1
    AlienVaultPacket 
    MaxPacket
)

// Types
type Packet struct {
    Type PacketType
    Host string
    Message string
}


func UDPListener(port string, finish <-chan bool, packets chan<- Packet) {
    log.Print("Attached UDP Listener: ", port)
	listener("udp", port, finish , packets)
}

func UnixDatagramListener(fileName string, finish <-chan bool, packets chan<- Packet) {
	listener("unixgram", fileName, finish, packets)
    log.Print("Attached UnixDatagram Listener: ", fileName)
}

func UnixStreamListener(fileName string, finish <-chan bool, packets chan<- Packet) {
	var err error 
	if _, err := os.Stat(fileName); err == nil {
		log.Print(fileName, " exists deleting..")
		err = os.Remove(fileName)
		if err != nil {
			log.Print("Unable to delete file")
			return
		}
	} else {
        log.Print("No file found")
    }

	listen, err := net.Listen("unix", fileName)
	if err != nil {
		log.Print("Unable to Listen to file: ", err)
		return
	}

	err = os.Chmod(fileName, 0666)
	if err != nil {
		log.Print("Unable to chmod file: ", err)
		return
	}
    
    log.Print("Attached UnixStream Listener: ", fileName)
	for {
		conn, err := listen.Accept()
		if err != nil {
            log.Print("Error: ", err)
			continue
		}

		go func() {
			for  {
				buffer := make([]byte, 1024)
				bytesRead, err := conn.Read(buffer[0:])
				if err != nil {
					 break;
				} else {
                    PacketsReceived++
					 packets <- Packet{ RFC3164Packet, "127.0.0.1", string(buffer[0:bytesRead]) }
				}
			}
		}()
	}


}

func listener(prot string, url string, finish <-chan bool, packets chan<- Packet) {
	listener, err := net.ListenPacket(prot, url)
	if err != nil { 
		log.Print("ListenPacket Failure: ", err, " not listening") 
		return
	}
    
    go func() {
        buffer := make([]byte, 1024)
        for  {
            bytesRead, address, err := listener.ReadFrom(buffer[0:])
            if err != nil {
                 log.Print("Listener: Unable to Read Packet!")
            } else {
                PacketsReceived++
                packets <- Packet{ RFC3164Packet, address.String(), string(buffer[0:bytesRead]) }
            }
        }
    }()

	for  {
		select  {
		case <- finish:
			log.Print("listener: signalled to end, closing")
			return
		}
	}
}

func AVListener(url string, finish <- chan bool, packets chan<- Packet) {
    listener, err := net.Listen("tcp", url)

    if err != nil {
        log.Print("Listen Failure: ", err, " not listening for AV Logger Packets")
    } else {
        log.Print("Attached new AV Logger: ", url )
    }

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Print("Accept failed: ", err)
            continue
        } else {
            log.Print("Got Connection")
        }

        go func(connection net.Conn) {
            buffer := make([]byte, 300)
            bytesRead, err := connection.Read(buffer[0:])
            if err != nil {
                return
            }

            if string(buffer[0:bytesRead]) == "Hello I am a OSSIM Agent" {
                connection.Write([]byte("HELLO\n"))
                for {
                    buffer := make([]byte, 1024)
                    bytesRead, err := connection.Read(buffer[0:])
                    if err != nil {
                        break;
                    } else {
                        PacketsReceived++
                        packets <- Packet{ AlienVaultPacket, connection.RemoteAddr().String(), string(buffer[0:bytesRead]) }
                    }
                }
            } else {
                log.Print("Received the wrong connection message: ", string(buffer[0:bytesRead]))
            }
        }(conn)
    }
}
