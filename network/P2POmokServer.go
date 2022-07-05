/*
**20162143, KimWonPyo
 */
package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Info struct {
	conn  net.Conn
	write chan string
	ip    string
	port  string
}

var arr map[string]Info = map[string]Info{}

var buffer = make([]byte, 1024)
var client_count int

func handleConn(conn net.Conn, info Info) {
	defer conn.Close()
	for {
		select {
		// first tcp connection close here
		case msg := <-info.write:
			conn.Write([]byte(msg))
			return
		default:
			time.Sleep(50 * time.Millisecond)
		}

	}
}

func main() {
	serverPort := "51243"
	client_count = 0
	listener, _ := net.Listen("tcp", ":"+serverPort)
	defer listener.Close()
	cancelChan := make(chan os.Signal)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for {

			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			buffer := make([]byte, 1024)
			count, _ := conn.Read(buffer)
			str := string(buffer[:count])
			slice := strings.Split(str, " ")
			// welcome ironman to p2p-omok server at 165.194.35.202:59999. waiting for an opponent.
			// ironman joined from 165.194.35.202:34567. UDP port 23456. 1 user connected, waiting for another
			localAddr := conn.LocalAddr().(*net.TCPAddr)
			// if there is no client -> then have to wait
			if client_count == 0 {
				client_count += 1
				arr[slice[0]] = Info{conn: conn, write: make(chan string), ip: slice[1], port: slice[3]}
				conn.Write([]byte("0welcome " + slice[0] + " to p2p-omok server at " + localAddr.IP.String() + ":" + strconv.Itoa(localAddr.Port) + ". waiting for an opponent."))
				fmt.Println(slice[0] + " joined from " + slice[1] + ":" + slice[2] + ". UDP port " + slice[3] + ". 1 user connected, waiting for another.")
			} else {
				// welcome ironman to p2p-omok server at 165.194.35.202:59999.ironman is waiting for you (165.194.35.202:23456).ironman plays first.
				client_count = 0
				// superman joined (165.194.35.202:78901). you play first.
				// if someone wait in server then have to delete wait client and tcp connection close
				for key, value := range arr {
					delete(arr, key)
					value.write <- "1 " + slice[0] + " joined (" + slice[1] + ":" + slice[3] + "). you play first."
					conn.Write([]byte("1welcome " + slice[0] + " to p2p-omok server at " + localAddr.IP.String() + ":" + strconv.Itoa(localAddr.Port) + ". " + key + " is waiting for you (" + value.ip + ":" + value.port + "). " + key + " plays first."))
					fmt.Println(slice[0] + " joined from " + slice[1] + ":" + slice[2] + ". UDP port " + slice[3] + ". 2 users connected, notifying " + key + " and " + slice[0] + ". " + key + " and " + slice[0] + " disconnected.")
					conn.Close()
					// second client here close tcp socket
				}
			}
			go handleConn(conn, arr[slice[0]])
		}
	}()

	select {
	case <-cancelChan:
		fmt.Println("Bye bye~")
		close(cancelChan)
	}
}
