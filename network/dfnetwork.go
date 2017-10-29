package network

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"

	"github.com/ruiqihong/paxos/log"
)

type AddrData struct {
	Addr   string
	SendCh chan []byte
}

type ConnErr struct {
	Addr string
	Err  error
}

type Connection struct {
	addr   string
	sendCh <-chan []byte
	conn   net.Conn
	bw     *bufio.Writer
}

func NewConnection(addr string, sendCh <-chan []byte) *Connection {
	conn := &Connection{
		addr:   addr,
		sendCh: sendCh}
	return conn
}

func (c *Connection) Run() {
	var err error
	for {
		if c.conn == nil {
			c.conn, err = net.Dial("tcp", c.addr)
			if err != nil {
				log.With(log.F{
					"addr": c.addr,
					"err":  err,
				}).Err("can't connect to peer")
				time.Sleep(time.Second)
				continue
			}
		}

		c.bw = bufio.NewWriter(c.conn)
		msg := <-c.sendCh

		size := uint32(len(msg))
		err = binary.Write(c.bw, binary.BigEndian, size)
		if err == nil {
			_, err = c.bw.Write(msg)
		}
		if err == nil {
			err = c.bw.Flush()
		}
		if err != nil {
			log.With(log.F{
				"addr": c.addr,
				"err":  err,
			}).Err("can't send message")
			c.conn.Close()
			c.conn = nil
		}
	}
}

type RecvData struct {
	Conn net.Conn
	Msg  []byte
}

type DFNetwork struct {
	handler     MessageHandler
	addrs       map[string]*AddrData
	recvCh      chan RecvData
	acceptErrCh chan error
	mu          sync.Mutex
}

type dfMessageHandler struct {
}

func (dfMessageHandler) OnMessage(msg []byte) error {
	return nil
}

func NewDFNetwork() *DFNetwork {
	network := new(DFNetwork)
	network.handler = dfMessageHandler{}
	network.addrs = make(map[string]*AddrData)
	network.recvCh = make(chan RecvData, 1024)
	network.acceptErrCh = make(chan error, 1)
	return network
}

func (n *DFNetwork) SetMessageHandler(handler MessageHandler) {
	n.handler = handler
}

func (n *DFNetwork) getSendChannel(addr string) chan<- []byte {
	n.mu.Lock()
	defer n.mu.Unlock()
	var sendCh chan []byte
	addrData, ok := n.addrs[addr]
	if ok {
		sendCh = addrData.SendCh
	} else {
		sendCh = make(chan []byte, 100)
		n.addrs[addr] = &AddrData{Addr: addr, SendCh: sendCh}
		conn := NewConnection(addr, sendCh)
		go conn.Run()
	}
	return sendCh
}

func (n *DFNetwork) SendMessage(addr string, message []byte) {
	select {
	case n.getSendChannel(addr) <- message:
	default:
	}
}

func (n *DFNetwork) handleConn(conn net.Conn) {
	defer conn.Close()

	var err error
	br := bufio.NewReader(conn)
	for {
		size := uint32(0)
		err = binary.Read(br, binary.BigEndian, &size)
		if err != nil {
			conn.Close()
			break
		}
		msg := make([]byte, size)
		_, err = io.ReadFull(br, msg)
		if err != nil {
			conn.Close()
			break
		}
		n.recvCh <- RecvData{Conn: conn, Msg: msg}
	}
}

func (n *DFNetwork) Run(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.With(log.F{
			"addr": addr,
			"err":  err,
		}).Err("listen failed")
		return err
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				n.acceptErrCh <- err
				break
			}
			go n.handleConn(conn)
		}
	}()

	for {
		err = n.ProcessEvent()
		if err != nil {
			log.With(log.F{
				"err": err,
			}).Err("ProcessEvent error")
			return err
		}
	}
}

func (n *DFNetwork) ProcessEvent() error {
	var err error
	select {
	case recvData := <-n.recvCh:
		err = n.handler.OnMessage(recvData.Msg)
		if err != nil {
			recvData.Conn.Close()
		}
		return nil
	case acceptErr := <-n.acceptErrCh:
		log.With(log.F{
			"err": acceptErr,
		}).Err("accept fail")
		return err
	}
	panic("not reach")
}
