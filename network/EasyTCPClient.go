/*
**20162143, KimWonPyo
*/
package main

import ("bufio"; "fmt"; "net"; "os";"os/signal";"syscall";"time";"strings";"strconv")

func main() {
    serverName := "nsl2.cau.ac.kr"
    serverPort := "41243"
    conn, err:= net.Dial("tcp", serverName+":"+serverPort)
    // if server does not open then don't need to do anything
    if err != nil{
        fmt.Println("\nBye bye")
        return
    }
    // many thing same as to UDPClient... 
    localAddr := conn.LocalAddr().(*net.TCPAddr)
    fmt.Printf("Client is running on port %d\n", localAddr.Port)
    cancelChan := make(chan os.Signal)
    five_signal := make(chan bool)
    signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
    defer conn.Close()
    go func() {
        
        for {
            
			fmt.Println("<Menu>")
			fmt.Println("1) convert text to UPPER-case")
			fmt.Println("2) get my IP address and port number")
			fmt.Println("3) get server request count")
			fmt.Println("4) get server running time")
			fmt.Println("5) exit")
			fmt.Printf("Input option : ")
			opt, _ := bufio.NewReader(os.Stdin).ReadString('\n')
            opt = strings.TrimSpace(opt)
            oopt,_ := strconv.Atoi(opt)
            if oopt == 5{
				five_signal <- true
                
			} else if oopt == 1{
                fmt.Printf("Input lowercase sentence: ")
                input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
                t := time.Now()
                conn.Write([]byte("1"+input))
                buffer := make([]byte, 1024)
                conn.Read(buffer) 
                rtt := time.Since(t)
                fmt.Printf("Reply from server : %s", string(buffer))
                time_since := rtt.Seconds()
                fmt.Printf("RTT time: %.3fms", time_since*1000)
                
            } else if oopt == 2{
                conn.Write([]byte("2"))
                buffer := make([]byte, 1024)
                conn.Read(buffer)
                slice := strings.Split(string(buffer), ":")
                fmt.Printf("Reply from server: client IP = %s, port = %s", slice[0],slice[1])
                
            } else if oopt == 3{
                conn.Write([]byte("3"))
                buffer := make([]byte, 1024)
                conn.Read(buffer)
                fmt.Printf("Reply from server: requests served : %s", string(buffer))
                
            } else if oopt == 4{
                conn.Write([]byte("4"))
                buffer := make([]byte, 1024)
                conn.Read(buffer)
                hour := strings.Split(string(buffer), "h")
                minute := strings.Split(string(buffer), "m")
               
                fmt.Printf("Reply from server: run time = ")
                if len(hour) == 1{
                    fmt.Printf("0:")
                }else{
                    fmt.Printf("%s:", hour[0])
                }
                if len(minute) == 1{
                    fmt.Printf("0:%s",strings.Split(minute[0],".")[0])
                }else{
                    fmt.Printf("%s:%s", minute[0],strings.Split(minute[1],".")[0])
                }
                
                

            
            }
            time.Sleep(50*time.Millisecond)
            fmt.Println("")
        }
        
    }()

    select{
    case <- five_signal:
        fmt.Println("Bye bye~")   
        close(cancelChan)
        close(five_signal)
    case <- cancelChan:
        fmt.Println("\nBye bye~")  
        close(cancelChan)
        close(five_signal) 
    }

    conn.Write([]byte("5"))
   
 
}

