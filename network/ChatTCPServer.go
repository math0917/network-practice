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
	conn        net.Conn
	write       chan string
	quit        chan bool
	server_quit chan bool
	ip          string
	port        string
}

var arr map[string]Info = map[string]Info{}

var client_count = 0
var server_version = "1.0.0.1"

//server only needs one go routine
func handleConn(conn net.Conn, info Info, nickname string) {
	defer conn.Close()
	// after return this function have to close conn
	for {
		// something needs to send to client
		select {
		case msg := <-info.write:
			conn.Write([]byte(msg))
		// when clients said to server "I want to quit", then server need to print, inform others how to find client? compare conn...
		case <-info.quit:
			client_count -= 1
			fmt.Println(nickname + " left. There are " + strconv.Itoa(client_count) + " users now.")

			for key, value := range arr {
				if value.conn == conn {

					delete(arr, key)
				} else {

					value.write <- nickname + " left. There are " + strconv.Itoa(client_count) + " users now."
				}
			}
			return
		// when I hate professor same as to above. have to inform client you kicked
		case <-info.server_quit:
			client_count -= 1
			fmt.Println(nickname + " left. There are " + strconv.Itoa(client_count) + " users now.")
			conn.Write([]byte("5"))

			for key, value := range arr {
				if value.conn == conn {
					delete(arr, key)
				} else {
					value.write <- nickname + " left. There are " + strconv.Itoa(client_count) + " users now."
				}
			}
		// else just read from client, wait for command,
		default:
			go ListenFromClient(conn, info, nickname)
			time.Sleep(100 * time.Millisecond)
		}
	}

}

func ListenFromClient(conn net.Conn, info Info, nickname string) {
	buffer := make([]byte, 1024)
	count, err := conn.Read(buffer)
	if err != nil {
		return
	}
	if strings.Contains(string(buffer[:count]), "I hate professor") {
		info.server_quit <- true
		return
	}
	if strings.HasPrefix(string(buffer[:count]), "5") {
		info.quit <- true
		return
	} else if strings.HasPrefix(string(buffer[:count]), "6") {

		for _, value := range arr {
			if value.conn != conn {
				value.write <- nickname + "> " + string(buffer[1:count])
			}
		}

	} else if strings.HasPrefix(string(buffer[:count]), "1") {

		var strings = ""
		for key, value := range arr {
			strings += "< " + key + ", " + value.ip + ", " + value.port + ">\n"
		}
		info.write <- strings
	} else if strings.HasPrefix(string(buffer[:count]), "2") {

		slice := strings.Split(string(buffer[1:count]), " ")
		str := string(buffer[1+len(slice[0])+1 : count])
		arr[slice[0]].write <- "from :" + nickname + ">" + str
	} else if strings.HasPrefix(string(buffer[:count]), "3") {

		info.write <- server_version
	} else if strings.HasPrefix(string(buffer[:count]), "4") {

		info.write <- "4"
	} else {
		fmt.Println("invalid command")
	}

}

func main() {
	serverPort := "41243"
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
			// if more than 8 then just close, send to client with 0,message
			if client_count >= 8 {
				conn.Write([]byte("0chatting room full. cannot connect"))
				conn.Close()
				continue
			}
			conn.Write([]byte("1"))

			buffer := make([]byte, 1024)
			count, _ := conn.Read(buffer)
			nickname := string(buffer[:count])
			slice := strings.Split(conn.RemoteAddr().String(), ":")
			_, exists := arr[nickname]
			// save with map and if there exists that name then send to client with 0, message
			if exists == true {
				conn.Write([]byte("0that nickname is already used by another user. cannot connect"))
				conn.Close()
			} else {
				// can add then have to initate variance and send to client with 1, message
				// and go~
				client_count += 1
				arr[nickname] = Info{conn: conn, write: make(chan string), quit: make(chan bool), server_quit: make(chan bool), ip: slice[0], port: slice[1]}
				localAddr := conn.LocalAddr().(*net.TCPAddr)
				conn.Write([]byte("1welcome <" + nickname + "> to CAU network class that room at <" + localAddr.IP.String() + ":" + strconv.Itoa(localAddr.Port) + ">. There are <" + strconv.Itoa(client_count) + "> users connected."))
				fmt.Println("<" + nickname + "> joined from <" + slice[0] + ":" + slice[1] + ">. There are <" + strconv.Itoa(client_count) + "> users connected")
				go handleConn(conn, arr[nickname], nickname)

			}

		}
	}()

	select {
	case <-cancelChan:
		fmt.Println("gg~")
		close(cancelChan)
	}
}
