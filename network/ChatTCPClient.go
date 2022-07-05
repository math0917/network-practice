/*
**20162143, KimWonPyo
 */
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// rtt is time sending and receiving
var rtt time.Time
var exit_signal = make(chan bool)
var five_signal = make(chan bool)

func receive(conn net.Conn, recmsg chan string) {
	for {
		select {
		// after reading , if there is need to print to client, then here
		case msg := <-recmsg:
			// when client send I hate professor, then server send to me you have to finish : "5"
			if strings.HasPrefix(msg, "5") {
				five_signal <- true
				return
			} else {
				fmt.Println(msg)
			}
			// else just reading what server send to me
		default:
			go receiveFromServer(conn, recmsg)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// else just reading what server send to me
func receiveFromServer(conn net.Conn, recmsg chan string) {
	buffer := make([]byte, 1024)
	count, err := conn.Read(buffer)
	if err != nil {
		return
	}
	if strings.HasPrefix(string(buffer[:count]), "4") {
		real := time.Since(rtt).Seconds() * 1000
		msg := strconv.FormatFloat(real, 'f', 3, 32)
		recmsg <- msg + "ms"
	} else {
		recmsg <- string(buffer[:count])
	}
}

func send(conn net.Conn) {
	// client's stdin just wait to send
	for {
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return
		}

		input = strings.TrimSpace(input)
		// one byte with "1","2",...,"6" command part, and dm with message part
		if strings.HasPrefix(input, "\\") {
			if strings.HasPrefix(input, "\\list") {
				conn.Write([]byte("1"))
			} else if strings.HasPrefix(input, "\\dm") {
				conn.Write([]byte("2" + input[4:]))
			} else if strings.HasPrefix(input, "\\ver") {
				conn.Write([]byte("3"))
			} else if strings.HasPrefix(input, "\\rtt") {
				rtt = time.Now()
				conn.Write([]byte("4"))
			} else if strings.HasPrefix(input, "\\exit") {
				exit_signal <- true
				return
			} else {
				fmt.Println("invalid command")
			}
		} else {
			conn.Write([]byte("6" + input))
		}

	}

}

func main() {

	serverName := "nsl2.cau.ac.kr"
	serverPort := "41243"
	var nickname = os.Args[1]
	conn, err := net.Dial("tcp", serverName+":"+serverPort)
	// if server does not open then don't need to do anything
	if err != nil {
		fmt.Println("gg~")
		return
	}

	// localAddr := conn.LocalAddr().(*net.TCPAddr)
	cancelChan := make(chan os.Signal)

	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	defer conn.Close()
	go func() {
		buffer := make([]byte, 1024)
		count, _ := conn.Read(buffer)
		opt, _ := strconv.Atoi(string(buffer[:1]))
		if opt == 0 {
			fmt.Println(string(buffer[1:count]))
			five_signal <- true
		}
		conn.Write([]byte(nickname))
		buffer = make([]byte, 1024)
		count, _ = conn.Read(buffer)
		opt, _ = strconv.Atoi(string(buffer[:1]))
		if opt == 0 {
			fmt.Println(string(buffer[1:count]))
			exit_signal <- true
		} else {
			fmt.Println(string(buffer[1:count]))
		}
		recmsg := make(chan string)
		go send(conn)
		receive(conn, recmsg)
	}()

	select {
	case <-five_signal:
		fmt.Println("gg~")
		close(cancelChan)
		close(five_signal)
		close(exit_signal)
	case <-exit_signal:
		fmt.Println("gg~")
		close(cancelChan)
		close(five_signal)
		close(exit_signal)
		conn.Write([]byte("5"))
	case <-cancelChan:
		fmt.Println("gg~")
		close(cancelChan)
		close(five_signal)
		close(exit_signal)
		conn.Write([]byte("5"))

	}

}
