package gateway

import (
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/lewyhua/plato/common/config"
	"golang.org/x/sys/unix"
)

// 全局对象
var ep *ePool    // epoll池
var tcpNum int32 // 当前服务允许接入的最大tcp连接数

type ePool struct {
	eChan        chan *connection // worker accept到conn之后丢入channel，交给后序epoll池处理
	tables       sync.Map
	eSize        int                              // epoll池的数量，就是多reactor模型里的reactor数量
	done         chan struct{}                    // 关闭epoll池的信号，用户后序优雅结束
	listener     *net.TCPListener                 // 网关监听socket 链接
	callBackFunc func(c *connection, ep *epoller) // 处理非连接socket的回调函数，通常是处理业务逻辑
}

func initEpoll(listener *net.TCPListener, runProcFunc func(c *connection, ep *epoller)) {
	setLimit()
	ep = newEPool(listener, runProcFunc)
	ep.createAcceptProcess()
	ep.startEPool()
}

func newEPool(listener *net.TCPListener, callBackFunc func(c *connection, ep *epoller)) *ePool {
	return &ePool{
		eChan:        make(chan *connection, config.GetGatewayEpollerChanNum()),
		done:         make(chan struct{}),
		eSize:        config.GetGatewayEpollerNum(),
		tables:       sync.Map{},
		listener:     listener,
		callBackFunc: callBackFunc,
	}
}

func (e *ePool) createAcceptProcess() {
	go func() {
		for {
			conn, err := e.listener.AcceptTCP()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					log.Printf("accept timeout: %v", err)
					continue
				}
				if errors.Is(err, net.ErrClosed) {
					log.Printf("listener closed, exiting accept loop")
					return
				}
				log.Printf("fatal accept error: %v", err)
				return
			}

			// 限流熔断
			if !checkConnRate() {
				log.Printf("connection from %v rejected by rate limiter", conn.RemoteAddr())
				conn.Close()
				continue
			}

			// 设置连接的配置，目前为手动打开keepalive
			setTCPConfig(conn)

			select {
			case e.eChan <- NewConnection(socketFD(conn), conn):
			default:
				log.Printf("task queue full, closing connection from %v", conn.RemoteAddr())
				conn.Close()
			}
		}
	}()
}

func (e *ePool) startEPool() {
	for i := 0; i < e.eSize; i++ {
		go e.startEProc()
	}
}

// 轮询器池 处理器
func (e *ePool) startEProc() {
	ep, err := newEpoller()
	if err != nil {
		panic(err)
	}
	// 监听连接创建事件
	go func() {
		for {
			select {
			case <-e.done:
				return
			case conn := <-e.eChan:
				addTcpNum()
				fmt.Printf("tcpNum:%d\n", tcpNum)
				if err := ep.add(conn); err != nil {
					fmt.Printf("failed to add connection to epoll tree %v\n", err)
					conn.Close() //登录未成功直接关闭连接
					continue
				}
				fmt.Printf("EpollerPool new connection[%v] tcpSize:%d\n", conn.RemoteAddr(), tcpNum)
			}
		}
	}()
	// 轮询器在这里轮询等待发来的, 当有wait发生时则调用回调函数去处理
	for {
		select {
		case <-e.done:
			return
		default:
			connections, err := ep.wait(200) // 200ms 一次轮询避免 忙轮询

			if err != nil && err != syscall.EINTR {
				fmt.Printf("failed to epoll wait %v\n", err)
				continue
			}
			for _, conn := range connections {
				if conn == nil {
					break
				}
				e.callBackFunc(conn, ep)
			}
		}
	}
}

// epoller 对象 轮询器
type epoller struct {
	fd int
}

func newEpoller() (*epoller, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &epoller{
		fd: fd,
	}, nil
}

// TODO: 默认水平触发模式,可采用非阻塞FD,优化边沿触发模式
func (e *epoller) add(conn *connection) error {
	// Extract file descriptor associated with the connection
	fd := conn.fd
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.EPOLLIN | unix.EPOLLHUP, Fd: int32(fd)})
	if err != nil {
		return err
	}
	ep.tables.Store(fd, conn)
	return nil
}
func (e *epoller) remove(conn *connection) error {
	subTcpNum()
	fd := conn.fd
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	ep.tables.Delete(fd)
	return nil
}
func (e *epoller) wait(msec int) ([]*connection, error) {
	events := make([]unix.EpollEvent, config.GetGatewayEpollWaitQueueSize())
	n, err := unix.EpollWait(e.fd, events, msec)
	if err != nil {
		return nil, err
	}
	var connections []*connection
	for i := 0; i < n; i++ {
		//log.Printf("event:%+v\n", events[i])
		if conn, ok := ep.tables.Load(int(events[i].Fd)); ok {
			connections = append(connections, conn.(*connection))
		}
	}
	return connections, nil
}
func socketFD(conn *net.TCPConn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(*conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

// 设置go 进程打开文件数的限制
func setLimit() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	log.Printf("set cur limit: %d", rLimit.Cur)
}

func addTcpNum() {
	atomic.AddInt32(&tcpNum, 1)
}

func getTcpNum() int32 {
	return atomic.LoadInt32(&tcpNum)
}
func subTcpNum() {
	atomic.AddInt32(&tcpNum, -1)
}

// checkTcp 检查当前tcp连接数是否超过限制
func checkConnRate() bool {
	num := getTcpNum()
	maxTCPNum := config.GetGatewayMaxTCPNum()
	return num <= maxTCPNum
}

// setTCPConfig 设置tcp连接的配置，当前为手动打开keepalive，以为开启长连接
func setTCPConfig(c *net.TCPConn) {
	_ = c.SetKeepAlive(true)
}
