/*
**20162143, KimWonPyo
 */
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var my_turn int
var nickname string
var oppoName string
var oppoPort string
var server_addr *net.UDPAddr
var user int
var turn int
var turn_count int
var win int
var board = Board{}
var turn_start = make(chan bool)
var x int
var y int
var do = make(chan bool)
var expire = make(chan bool)
var exit = make(chan bool)
var op_nick string
var can_play = true
var op_left = false

const (
	Row = 10
	Col = 10
)

type Board [][]int

func printBoard(b Board) {
	fmt.Print("   ")
	for j := 0; j < Col; j++ {
		fmt.Printf("%2d", j)
	}

	fmt.Println()
	fmt.Print("  ")
	for j := 0; j < 2*Col+3; j++ {
		fmt.Print("-")
	}

	fmt.Println()

	for i := 0; i < Row; i++ {
		fmt.Printf("%d |", i)
		for j := 0; j < Col; j++ {
			c := b[i][j]
			if c == 0 {
				fmt.Print(" +")
			} else if c == 1 {
				fmt.Print(" 0")
			} else if c == 2 {
				fmt.Print(" @")
			} else {
				fmt.Print(" |")
			}
		}

		fmt.Println(" |")
	}

	fmt.Print("  ")
	for j := 0; j < 2*Col+3; j++ {
		fmt.Print("-")
	}

	fmt.Println()
}

func checkWin(b Board, x, y int) int {
	lastStone := b[x][y]
	startX, startY, endX, endY := x, y, x, y

	// Check X
	for startX-1 >= 0 && b[startX-1][y] == lastStone {
		startX--
	}
	for endX+1 < Row && b[endX+1][y] == lastStone {
		endX++
	}

	if endX-startX+1 >= 5 {
		return lastStone
	}

	// Check Y
	startX, startY, endX, endY = x, y, x, y
	for startY-1 >= 0 && b[x][startY-1] == lastStone {
		startY--
	}
	for endY+1 < Row && b[x][endY+1] == lastStone {
		endY++
	}

	if endY-startY+1 >= 5 {
		return lastStone
	}

	// Check Diag 1
	startX, startY, endX, endY = x, y, x, y
	for startX-1 >= 0 && startY-1 >= 0 && b[startX-1][startY-1] == lastStone {
		startX--
		startY--
	}
	for endX+1 < Row && endY+1 < Col && b[endX+1][endY+1] == lastStone {
		endX++
		endY++
	}

	if endY-startY+1 >= 5 {
		return lastStone
	}

	// Check Diag 2
	startX, startY, endX, endY = x, y, x, y
	for startX-1 >= 0 && endY+1 < Col && b[startX-1][endY+1] == lastStone {
		startX--
		endY++
	}
	for endX+1 < Row && startY-1 >= 0 && b[endX+1][startY-1] == lastStone {
		endX++
		startY--
	}

	if endY-startY+1 >= 5 {
		return lastStone
	}

	return 0
}

func clear() {
	fmt.Printf("%s", runtime.GOOS)

	clearMap := make(map[string]func()) //Initialize it
	clearMap["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clearMap["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	value, ok := clearMap[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                             //if we defined a clearMap func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clearMap terminal screen :(")
	}
}

// this function is from server
func receive(conn net.Conn, recmsg chan string) {
	defer conn.Close()
	for {
		select {
		case msg := <-recmsg:
			// if 0 + content -> server sent to client with inforamtion (not tcp close)
			if strings.HasPrefix(msg, "0") {
				fmt.Println(msg[1:])
			} else {
				// if 1 + content -> tcp connection close and have to know other client's ip address and udp port
				oppoName = strings.Split(strings.Split(strings.Split(msg, "(")[1], ")")[0], ":")[0]
				oppoPort = strings.Split(strings.Split(strings.Split(msg, "(")[1], ")")[0], ":")[1]
				slice := strings.Split(msg[1:], ".")
				op_nick = strings.Split(slice[len(slice)-6], " ")[1]
				if strings.Contains(msg, "you play first") {
					user = 1
					my_turn = 1
				} else {
					user = 2
					my_turn = 0
				}
				fmt.Println(msg[1:])

				return
			}
		default:
			go receivefromServer(conn, recmsg)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// receive from server function
func receivefromServer(conn net.Conn, recmsg chan string) {
	buffer := make([]byte, 1024)
	count, err := conn.Read(buffer)
	if err != nil {
		return
	}
	recmsg <- string(buffer[:count])

}

// this part is when p2p omok game client send part
func send_udp(pconn net.PacketConn) {
	for {

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			continue
		}

		input = strings.TrimSpace(input)
		if strings.HasPrefix(input, "\\") {
			if strings.HasPrefix(input, "\\\\") {
				// this part is not my turn ...
				if my_turn == 0 {
					fmt.Println("It's not your turn")
					continue
				}
				// if game finish there is no need to do anything but else do below
				if can_play {
					if len(strings.Split(input, " ")) != 3 {
						fmt.Println("error, must enter \\\\ x y!")
						time.Sleep(1 * time.Second)
						continue
					}
					xx, _ := strconv.Atoi(strings.Split(input, " ")[1])
					yy, _ := strconv.Atoi(strings.Split(input, " ")[2])

					if xx < 0 || yy < 0 || xx >= Row || yy >= Col {
						fmt.Println("error, out of bound!")
						time.Sleep(1 * time.Second)
						continue
					} else if board[xx][yy] != 0 {
						fmt.Println("error, already used!")
						time.Sleep(1 * time.Second)
						continue
					}
					do <- true
					if turn == 0 {
						board[xx][yy] = 1
					} else {
						board[xx][yy] = 2
					}
					x = xx
					y = yy
					server_addr, _ := net.ResolveUDPAddr("udp", oppoName+":"+oppoPort)
					pconn.WriteTo([]byte("0"+strings.Split(input, " ")[1]+strings.Split(input, " ")[2]), server_addr)

					my_turn = 0

					turn_start <- true
					// omok game logic
				} else {
					// if can't play -> here
					fmt.Println("already_finish")
				}
			} else if strings.HasPrefix(input, "\\gg") {
				if can_play {
					fmt.Println("you lose")
					can_play = false
					if op_left {
						continue
					}
					server_addr, _ := net.ResolveUDPAddr("udp", oppoName+":"+oppoPort)
					pconn.WriteTo([]byte("2"), server_addr)
					if my_turn == 1 {
						do <- true
					} else {
						continue
					}
				}

				// if gg -> can_play = false and have to send other client with "2"

			} else if strings.HasPrefix(input, "\\exit") {
				if op_left {
					exit <- true

				} else {
					server_addr, _ := net.ResolveUDPAddr("udp", oppoName+":"+oppoPort)
					pconn.WriteTo([]byte("3"), server_addr)

					exit <- true
					do <- true
				}
				// if exit -> if op_left -> then just go exit channel (just finish)
				// 		   -> if not op_left -> have to send other client with "3" which means op_left....
			} else {
				fmt.Println("invalid command")
			}
		} else {
			server_addr, _ := net.ResolveUDPAddr("udp", oppoName+":"+oppoPort)

			pconn.WriteTo([]byte("1"+nickname+"> "+input), server_addr)
		}

	}

}

// 10seconds timer...
func timer(pconn net.PacketConn) {
	this_time := time.Now()
	// flag := 0
	for {
		select {
		case <-do:
			return
		default:
			if time.Since(this_time).Seconds() > 10 {
				fmt.Println("you lose")
				can_play = false

				server_addr, _ := net.ResolveUDPAddr("udp", oppoName+":"+oppoPort)
				pconn.WriteTo([]byte("2"), server_addr)
				return
			}
			time.Sleep(10 * time.Millisecond)

		}
	}
}

func receive_udp(pconn net.PacketConn, recmsg chan string) {
	for {

		select {
		// have to print
		case msg := <-recmsg:
			fmt.Println(msg)
		// every time start time
		case <-turn_start:
			clear()
			printBoard(board)

			win = checkWin(board, x, y)

			if win != 0 {
				if win == user {
					fmt.Println("you win")
				} else {
					fmt.Println("you lose")
				}
				can_play = false
				continue
			}

			turn_count += 1
			if turn_count == Row*Col {
				fmt.Printf("draw!\n")
				can_play = false

				continue
			}

			turn = (turn + 1) % 2
			if my_turn == 1 {
				go timer(pconn)
			}
		// when there is no need to start or print then receive
		default:
			go receive_udp_fromOp(pconn, recmsg)
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func receive_udp_fromOp(pconn net.PacketConn, recmsg chan string) {

	buffer := make([]byte, 1024)
	count, _, err := pconn.ReadFrom(buffer)

	if err != nil {
		return
	}
	// 1 -> this info is x,y position
	if strings.HasPrefix(string(buffer[:count]), "1") {
		recmsg <- string(buffer[1:count])
	} else if strings.HasPrefix(string(buffer[:count]), "0") {
		xx, _ := strconv.Atoi(string(buffer[:count][1]))
		yy, _ := strconv.Atoi(string(buffer[:count][2]))
		if turn == 0 {
			board[xx][yy] = 1
		} else {
			board[xx][yy] = 2
		}
		x = xx
		y = yy
		my_turn = 1
		turn_start <- true
		// 2 -> op_gg
	} else if strings.HasPrefix(string(buffer[:count]), "2") {

		can_play = false

		fmt.Println("you win")
		do <- true

	} else {
		// 3 -> opponent exit
		can_play = false
		op_left = true
		fmt.Println("opponent left")
		do <- true

	}
}

func main() {

	serverName := "nsl2.cau.ac.kr"
	serverPort := "51243"
	nickname = os.Args[1]
	conn, err := net.Dial("tcp", serverName+":"+serverPort)
	if err != nil {
		fmt.Println("gg~")
		return
	}
	pconn, err := net.ListenPacket("udp", ":")
	defer pconn.Close()
	if err != nil {
		fmt.Println("gg~")
		return
	}

	localUDP := pconn.LocalAddr().(*net.UDPAddr)
	localAddr := conn.LocalAddr().(*net.TCPAddr)
	cancelChan := make(chan os.Signal)

	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	conn.Write([]byte(nickname + " " + localAddr.IP.String() + " " + strconv.Itoa(localAddr.Port) + " " + strconv.Itoa(localUDP.Port)))
	recmsg := make(chan string)
	receive(conn, recmsg)

	go func() {
		x, y, turn, turn_count, win = 0, 0, -1, 0, 0
		for i := 0; i < Row; i++ {
			var tempRow []int
			for j := 0; j < Col; j++ {
				tempRow = append(tempRow, 0)
			}
			board = append(board, tempRow)
		}

		go send_udp(pconn)
		recmsg := make(chan string)
		go receive_udp(pconn, recmsg)
		turn_start <- true
	}()
	select {
	case <-cancelChan:
		if op_left == false {
			server_addr, _ := net.ResolveUDPAddr("udp", oppoName+":"+oppoPort)
			pconn.WriteTo([]byte("3"), server_addr)
			fmt.Println("Bye~")
			close(cancelChan)
		} else {
			fmt.Println("Bye~")
			close(cancelChan)
		}

	case <-exit:
		fmt.Println("Bye~")
	}

}
