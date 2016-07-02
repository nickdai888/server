//
//func responseCommand() {
//	for {
//		var msg string
//		fmt.Scan(&msg)
//
//		if msg == "C1" {
//			os.Exit(1)
//		} else if msg == "C2" {
//
//		} else if msg == "C3" {
//
//		} else {
//
//		}
//
//		//		b := Packet([]byte (msg))
//		//        conn.Write(b)
//	}
//}
////
//func handleConnection(conn net.Conn) {
//	//声明一个临时缓冲区，用来存储被截断的数据
//	tmpBuffer := make([]byte, 0)
//
//	//声明一个管道用于接收解包的数据
//	readerChannel := make(chan []byte, 16)
//	go readerDataFromChannel(readerChannel, conn)
//
//	buffer := make([]byte, 1024)
//	for {
//		n, err := conn.Read(buffer)
//		if err != nil {
//			Log(conn.RemoteAddr().String(), " connection error: ", err)
//			return
//		}
//
//		tmpBuffer = Unpack(append(tmpBuffer, buffer[:n]...), readerChannel)
//	}
//}
