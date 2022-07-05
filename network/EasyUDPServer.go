/*
**20162143, KimWonPyo
*/
package main

import ("bytes"; "fmt"; "net";"time";"strconv";"os";"os/signal";"syscall")

func main() {
    serverPort := "51243"
    server_count := 0
    t := time.Now()
    pconn, err:= net.ListenPacket("udp", ":"+serverPort)
    if err != nil {
        return
    }
    fmt.Printf("Server is ready to receive on port %s\n", serverPort)
    // server ctrl + c listener
    cancelChan := make(chan os.Signal)
    signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
    buffer := make([]byte, 1024)
    // want to finish
    defer pconn.Close()
    go func(){
        for {
            count, r_addr, err:= pconn.ReadFrom(buffer)
            if err != nil{
                continue
            }
            fmt.Printf("UDP message from %s\n", r_addr.String())
            fmt.Printf("Command : %c\n", buffer[0])
            // client sent to server with operation number in buffer[0]
            opt,_ := strconv.Atoi(string(buffer[:1]))
            server_count += 1
            if opt == 1{
                pconn.WriteTo(bytes.ToUpper(buffer[1:count]), r_addr)
            } else if opt == 2{
                pconn.WriteTo([]byte(r_addr.String()), r_addr)
            } else if opt == 3{
                pconn.WriteTo([]byte(strconv.Itoa(server_count)), r_addr)
            } else if opt == 4{
                pconn.WriteTo([]byte(time.Since(t).String()),r_addr)
            }
        
        }
        
    }()

    select{
    // if ctrl + c occurs
    case <- cancelChan:
        fmt.Println("Bye bye~")
        close(cancelChan)
    }
    
    
    
    
    


}

