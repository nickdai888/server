//服务端解包过程
package main


import (
    "fmt"
    "net"
    "os"
    "time"
    "log"
    "protos"
     proto "github.com/golang/protobuf/proto"
)

//ss
var debugLog *log.Logger
var gLogFile *os.File

func main() {
    netListen, err := net.Listen("tcp", ":9999")
    CheckError(err)
    
    
    defer netListen.Close()
	defer uninit()

    Log("Waiting for clients")
    for {
        conn, err := netListen.Accept()
        if err != nil {
            continue
        }

        Log(conn.RemoteAddr().String(), " tcp connect success")
        go handleConnection(conn)
    }
    
}


func handleConnection(conn net.Conn) {
    //声明一个临时缓冲区，用来存储被截断的数据
    tmpBuffer := make([]byte, 0)

	
    //声明一个管道用于接收解包的数据
    readerChannel := make(chan []byte, 16)
    go reader(readerChannel,conn)


    buffer := make([]byte, 1024)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            Log(conn.RemoteAddr().String(), " connection error: ", err)
            return
        }

        tmpBuffer = Unpack(append(tmpBuffer, buffer[:n]...), readerChannel)
    }
}

 
func reader(readerChannel chan []byte, conn net.Conn) {
    for {
        select {
        	case data := <-readerChannel:
//            	Log(string(data))
//            	conn.Write(data)
				processMessage(data)
				//数据回传
            	b := Packet([]byte(data))
      			conn.Write(b)
      			Log("Data has been sent back to client!")
            case <-time.After(time.Second*600):
            	 Log("It's really weird to get Nothing!!!")
            	 conn.Close()
        }
    }
}

func processMessage(msgbuf []byte){
	   msg := &protos.Helloworld{}
	   err := proto.Unmarshal(msgbuf, msg) //unSerialize
	   CheckError(err)
	   Log("the msg is", msg)
}

func Log(v ...interface{}) {
//    fmt.Println(v...)
	debugLog.Println(v...)
}

func init(){
	logFileName := "server.log"
	logFile,err := os.Create(logFileName)
//	defer logFile.Close()
	gLogFile = logFile
	if nil != err {
		log.Fatalln("Open log file Err !")
	} 
	debugLog = log.New(logFile, "[Debug]", log.Lshortfile|log.Ltime|log.Ldate)
}

func uninit(){
	Log("uninite called")
	gLogFile.Close()
}

func CheckError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
        os.Exit(1)
    }
}