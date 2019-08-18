// tcpclient.go
package network

import (
	"bufio"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"iapps.site/common"
	"iapps.site/proto"
)

type TcpClient struct {
	conn      net.Conn
	rw        *bufio.ReadWriter
	connected bool
	account   *Account
	ch        chan *common.MsgWrap
	latch     sync.WaitGroup
	heartbeat int64
	id        uint64
}

// account info
type Account struct {
	AccId     string
	AccName   string
	RoleId    int64
	AuthToken string
}

func (client *TcpClient) connect(addr string) error {
	//log.Printf("client_%d Dial %s", client.id, addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	client.rw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	client.conn = conn
	client.connected = true

	return nil
}

// send login message to server
func (client *TcpClient) loginServer() error {
	data := proto.CG_LonginMessage{
		ProtoId: proto.CG_LonginMessageId,
		AccId:   client.account.AccId,
		AccName: client.account.AccName,
	}

	log.Println("client send login to server accId=", client.account.AccId)
	return handleEncodeGob(client.rw.Writer, data)
}

func (client *TcpClient) isConnected() bool {
	return client.connected
}

//
func (client *TcpClient) isLoingOk() bool {
	return client.account.RoleId > 0
}

func (client *TcpClient) FillAccount(roleId int64) {
	client.account.RoleId = roleId
}

func (client *TcpClient) Id() uint64 {
	return client.id
}

func (client *TcpClient) InitAccount(accId string, accName string) {
	account := &Account{
		AccId:   accId,
		AccName: accName,
	}
	client.account = account
}

func (client *TcpClient) Close() error {
	err := client.conn.Close()
	if err == nil {
		client.rw = nil
		client.connected = false
		client.latch.Done()
	}
	return err

}

//
func (client *TcpClient) Start() error {
	//
	if !client.isConnected() {
		return errors.New("client not connected yet!")
	}
	if client.account == nil {
		return errors.New("Please call client.InitAccount(string,string) first!")
	}

	log.Printf("Client start clientId=%d\n", client.id)
	session := common.NewSession(client.conn, &client, client.rw)

	if !client.isLoingOk() {
		client.latch.Add(1)
		err := client.loginServer()
		if err != nil {
			log.Println("Login server error:", err)
			client.latch.Done()
		}
	}

	// handle decode message
	go func() {
		for {
			err := handleClientDecodeGob(session, client.rw.Reader)
			if err != nil {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// blocking current client
	client.latch.Wait()
	return nil
}

// decode gob message from buffer for client
func handleClientDecodeGob(session *common.Session, r *bufio.Reader) error {
	buf := make([]byte, 16)
	n, _err := io.ReadFull(r, buf)
	if _err != nil {
		return _err
	}
	if n < 16 {
		log.Printf("not a complete protoId, wait for next read \n")
		return nil
	}
	protoId := binary.BigEndian.Uint16(buf)
	if protoId <= 0 {
		return errors.New("something err occured")
	}
	data, e := proto.GetProtoManagerIns().NewProto(protoId)
	if e != nil {
		log.Printf("Error decoding protoId=%d not found \n%#v\n", protoId, e)
		return e
	}

	//log.Printf("Client Handle DecodeGob protoId=%d , empty data= %#v\n", protoId, data)

	dec := gob.NewDecoder(r)
	err := dec.Decode(data)

	if err != nil {
		log.Printf("Client Error decoding GOB data \n%#v:\n", err)
		return err
	}

	log.Printf("Client Decodeed ProtoMessage data: \n== %#v\n", data)

	GetActionManagerIns().Handle(data, session)

	return nil
}

// create a TcpCient
// call InitAccount
// call Start
func NewTcpClient(addr string, id uint64) (*TcpClient, error) {
	var latch sync.WaitGroup
	var heartBeat int64 = time.Now().UnixNano() / int64(time.Millisecond)
	msgChannel := make(chan *common.MsgWrap, 20)

	client := &TcpClient{nil, nil, false, nil, msgChannel, latch, heartBeat, id}

	err := client.connect(addr)
	if err != nil {
		return nil, err
	}

	return client, nil
}
