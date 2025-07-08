package gateway

import "net"

type connection struct {
	fd   int
	conn *net.TCPConn
}

func (c *connection) Close() {
	err := c.conn.Close()
	panic(err)
}

func (c *connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func NewConnection(fd int, conn *net.TCPConn) *connection {
	if conn == nil {
		return nil
	}
	return &connection{
		fd:   fd,
		conn: conn,
	}
}
