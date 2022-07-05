/*
**20162143, KimWonPyo
*/
package main

import ("bytes"; "fmt"; "net";"time";"strconv";"os";"os/signal";"syscall")
var server_count int
var t time.Time
var buffer = make([]byte, 1024)
var client_index int
var client_count int
// server_count is for client's third command
// t is for 60 second print
// buffer is buffer
// client_index is for expected client's connect number
// client_count is for client's connecting count
// for go routine... there needs to use global variable...
func handleConn(conn net.Conn, this_turn_client int){
	for {
		count, _ := conn.Read(buffer)
		opt,_ := strconv.Atoi(string(buffer[:1]))
		
		if opt != 5{
			fmt.Printf("message from %s\n", conn.RemoteAddr().String())
			fmt.Printf("Command : %c\n", buffer[0])
			server_count += 1
		}
		
		if opt == 1{
			conn.Write(bytes.ToUpper(buffer[1:count]))
		} else if opt == 2{
			conn.Write([]byte(conn.RemoteAddr().String()))
		} else if opt == 3{
			conn.Write([]byte(strconv.Itoa(server_count)))
		} else if opt == 4{
			conn.Write([]byte(time.Since(t).String()))
		} else if opt == 5{
			break
		}
		
		
	}
	// one client's tcp finishes then close that.
	client_count -= 1
	fmt.Printf("Client %d disconnected. Number of connected clients = %d\n", this_turn_client, client_count)
	
	defer conn.Close()
}
// just sleep(60 second) and then wake, printf
func await(){
	for{
		time.Sleep(60*time.Second)
		fmt.Printf("Number of connected clients = %d\n",client_count)
	}
}
func main() {
    serverPort := "41243"
    server_count = 0
    t = time.Now()
	// there needs to do 60 second for loop
    go await()
	listener, _:= net.Listen("tcp", ":" + serverPort)
    fmt.Printf("Server is ready to receive on port %s\n", serverPort)
    defer listener.Close()
    cancelChan := make(chan os.Signal)
    signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
    
    go func(){
        for{
            
            conn, err := listener.Accept()
            if err != nil{
                continue
            }
			client_index += 1
			client_count += 1
			this_turn_client := client_index
			fmt.Printf("Client %d connected. Number of connected clients = %d\n", client_index, client_count)
            go handleConn(conn,this_turn_client)
        }
    }()
	
    select{
    case <-cancelChan:
        fmt.Println("Bye bye~")
        close(cancelChan)
    }
}

