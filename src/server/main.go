//服务端解包过程
package main

import (
//	"fmt"
	proto "github.com/golang/protobuf/proto"
	"log"
	"net"
	"os"
	"os/signal"
	"protos"
	"sync"
	"syscall"
	"time"
	"utils"
)

type Service struct {
	ch        chan bool
	waitGroup *sync.WaitGroup
}

func NewService() *Service {
	return &Service{
		ch:        make(chan bool),
		waitGroup: &sync.WaitGroup{},
	}
}


func main() {
	//首先创建一个TCP的监听地址
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:9999")
	if nil != err {
		log.Fatalln(err)
	}
	//监听地址和对应端口
	listener, err := net.ListenTCP("tcp", laddr)
	if nil != err {
		log.Fatalln(err)
	}
	//输出监听地址信息，开始监听
	log.Println("Listening on:", listener.Addr())

	//创建service
	service := NewService()
	//将监听对象传递给service，启动service的服务
	service.Serve(listener)

	//等待系统信号的产生，启动结束服务器的流程
	ch := make(chan os.Signal, 1)
	//func Notify(c chan<- os.Signal, sig ...os.Signal),信号发生的时候把信号写入channel
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM,syscall.SIGQUIT)
	//让主线程挂起（读取管道数据），并等待系统事件的发生
	log.Println(<-ch)
	//停止系统服务,在系统事件发生后
	service.Stop()
}

//停止系统服务
func (s *Service) Stop() {
	//关闭s.ch，会导致往channer中写入false，关闭后channel仍然是可以读的，但是不可以写
	close(s.ch)
	//等待所有的routine停止后，优雅的退出
	s.waitGroup.Wait()
	utils.Log("Exit program gracefully!")
}

//启动系统服务
func (s *Service) Serve(listener *net.TCPListener) {
	s.waitGroup.Add(1)
	go func() {
		defer s.waitGroup.Done()
		for {
			//整个select代码块的作用是：判断当前程序是否需要退出，如果需要退出则需要关闭监听并结束routine
			select {
			case <-s.ch:
				log.Println("stopping listen on:", listener.Addr())
				listener.Close()
				return //return的和break的区别在于 break只能跳出select并不能跳出for循环结束routine
			default: //这里default的作用是跳过select，运行后面的代码
			}
			//设置TCP listener的挂起时间,每次间隔1秒恢复一次
			listener.SetDeadline(time.Now().Add(3 * time.Second))
			//成功获取一个连接请求
			conn, err := listener.AcceptTCP()
			//存在多种情况返回错误
			if nil != err {
				//判断是否是超时返回，如果是的话重新设置超时时间并继续监听端口
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					utils.Log("timeout and continue")
					continue
				}
				log.Println(err)
			}
//			if raddr, ok := conn.RemoteAddr.(net.Addr); ok {
//				log.Println("server get connection from:",raddr)
//			}
			log.Println("server get connection from:",conn.RemoteAddr())
			s.ProcessConnction(conn)
		}
	}()
}

//处理connection，对数据进行读写操作
func (s *Service) ProcessConnction(conn *net.TCPConn) {
	s.waitGroup.Add(1)
	go func() {
		//确保routine销毁前关闭了连接
		defer s.waitGroup.Done()
		defer conn.Close()

		//声明一个临时缓冲区，用来存储被截断的数据
		tmpBuffer := make([]byte, 0)
		//声明一个buffer用来读取conn中的数据
		buffer := make([]byte, 1024)
		//声明一个管道用于接收解包的数据
		readerChannel := make(chan []byte, 16)
		//开辟另外的routine，在其中获取数据包
		go readerDataFromChannel(readerChannel, conn)

		for {
			//这里的逻辑参考listener
			select {
			case <-s.ch:
				log.Println("disconnection:", conn.RemoteAddr)
				return
			default:
			}
			//读取数据并处理超时错误
			conn.SetDeadline(time.Now().Add(time.Second * 60)) //60秒阻塞时间
			n, err := conn.Read(buffer)
			if nil != err {
				//判断是否是设置的超时时间到了，如果是的话继续设置超时时间并从连接中读取数据
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					log.Println("server read data time out")
					continue
				}
				//如果是其它错误直接销毁routine
				log.Println(conn.RemoteAddr().String(), " quit routine,connection error: ", err)
				return
			}
			//将获取的数据(buffer[:n])传入解包模块
			tmpBuffer = Unpack(append(tmpBuffer, buffer[:n]...), readerChannel)
		}
	}()
}


//从channel中获取解包后的数据,如果有解包成功的package，则对应的数据会写入readerChannel
func readerDataFromChannel(readerChannel chan []byte, conn net.Conn) {
	for {
		select {
		case data := <-readerChannel:
			processMessage(data)
			//数据回传
			b := Packet([]byte(data))
			conn.Write(b)
			log.Println("Data has been sent back to client!")
			//对应的时间内如果没有收到数据则进行超时处理，表明客户端和服务器的连接已经断开,心跳包确保Client每间隔一段时间会往服务器发送一个包
		case <-time.After(time.Second * 60):
			log.Println("It's really weird to get Nothing!!!")
			conn.Close()
		}
	}
}

//处理解包后的数据包
func processMessage(msgbuf []byte) {
	msg := &protos.Helloworld{}
	err := proto.Unmarshal(msgbuf, msg) //unSerialize
	utils.CheckError(err)
	utils.Log("the msg is", msg)
}




