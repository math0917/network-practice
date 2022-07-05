/*
**20162143, KimWonPyo
*/
package main

import ("bufio"; "fmt"; "net"; "os";"os/signal";"syscall";"time";"strings";"strconv")

func main() {
    serverName := "nsl2.cau.ac.kr"
    serverPort := "51243"
    // udp server port 51243
    pconn, _:= net.ListenPacket("udp", ":")
   
    localAddr := pconn.LocalAddr().(*net.UDPAddr)
    fmt.Println("Client is running on port : ", localAddr.Port)
    server_addr, _ := net.ResolveUDPAddr("udp", serverName+":"+serverPort)
    
    // ctrl + c listener
    cancelChan := make(chan os.Signal)
    signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
    // user 5 listener
    five_signal := make(chan bool)
    // afterall, end of the main pconn close
    defer pconn.Close()
    go func() {
        for {
            
			fmt.Println("<Menu>")
			fmt.Println("1) convert text to UPPER-case")
			fmt.Println("2) get my IP address and port number")
			fmt.Println("3) get server request count")
			fmt.Println("4) get server running time")
			fmt.Println("5) exit")
			fmt.Printf("Input option : ")
			opt, _:= bufio.NewReader(os.Stdin).ReadString('\n')
            
            opt = strings.TrimSpace(opt)
            oopt,_ := strconv.Atoi(opt)
            if oopt == 5{
                five_signal <- true
				
			} else if oopt == 1{
                fmt.Printf("Input lowercase sentence: ")
                input,_ := bufio.NewReader(os.Stdin).ReadString('\n')
               
                t := time.Now()
                pconn.WriteTo([]byte("1"+input), server_addr)
                buffer := make([]byte, 1024)
                
                pconn.ReadFrom(buffer)
                rtt := time.Since(t)
                fmt.Printf("Reply from server : %s", string(buffer))
                time_since := rtt.Seconds()
                // time_since would be float second -> change form to millisecond
                fmt.Printf("RTT time: %.3fms", time_since*1000)
                
            } else if oopt == 2{
                pconn.WriteTo([]byte("2"), server_addr)
                buffer := make([]byte, 1024)
                pconn.ReadFrom(buffer)
                slice := strings.Split(string(buffer), ":")
                // server would be sent to client with ~.~.~.~:~ so use slice
                fmt.Printf("Reply from server: client IP = %s, port = %s", slice[0],slice[1])
                
            } else if oopt == 3{
                pconn.WriteTo([]byte("3"), server_addr)
                buffer := make([]byte, 1024)
                pconn.ReadFrom(buffer)
                // just string...
                fmt.Printf("Reply from server: requests served : %s", string(buffer))
                
            } else if oopt == 4{
                pconn.WriteTo([]byte("4"), server_addr)
                buffer := make([]byte, 1024)
                pconn.ReadFrom(buffer)
                str := strings.Split(string(buffer), "s")
                
                time_since,_ := time.ParseDuration(str[0]+"s")
                // server sent to me time.since use proper transform...
                fmt.Printf("Reply from server: run time = %.0f:%.0f:%.0f", time_since.Hours(), time_since.Minutes(), time_since.Seconds() )
            } 
            // goroutine speed control
            time.Sleep(50*time.Millisecond)
            fmt.Println("")
            
        }
    }()
    
    select{
    // if chan get signal then~~
    case <- five_signal:
        fmt.Println("Bye bye~")
        close(cancelChan)
        close(five_signal)
        
    case <-cancelChan:
        fmt.Println("\nBye bye~")
        close(cancelChan)
        close(five_signal)
        
    }
}

