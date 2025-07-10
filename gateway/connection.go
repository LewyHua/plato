package gateway

import (
	"net"
	"sync/atomic"
)

var nextConnID uint64 //分配全局唯一的连接ID

type connection struct {
	id   uint64       // 进程级别的生命周期
	fd   int          // 连接的文件描述符
	conn *net.TCPConn // 连接对象
	e    *epoller     // 连接所在的epoller对象
}

func (c *connection) Close() {
	ep.tables.Delete(c.id) //从全局epool对象的tables中删除该连接映射
	if c.e != nil {
		c.e.fdToConnTable.Delete(c.fd) // 从epoll的fd映射表中删除
	}
	err := c.conn.Close()
	panic(err)
}

func (c *connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func NewConnection(fd int, conn *net.TCPConn) *connection {
	connID := atomic.AddUint64(&nextConnID, 1) // 分配一个全局唯一的连接ID
	return &connection{
		id:   connID,
		fd:   fd,
		conn: conn,
	}
}

func (c *connection) BindEpoller(e *epoller) {
	c.e = e
}
