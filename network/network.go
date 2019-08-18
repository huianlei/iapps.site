// network.go
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

const (
	KeepAlive = int64(30 * time.Second)
)

type HandleDecodeFunc func(*common.Session, *bufio.Reader) error

type HandleEncodeFunc func(*bufio.Writer, proto.ProtoMessage) error

// singletone
var server *tcpServer = nil
var onceServer sync.Once

type tcpServer struct {
	listener      net.Listener
	decodeHandler HandleDecodeFunc
	encodeHandler HandleEncodeFunc
}

func TcpServerIns() *tcpServer {
	onceServer.Do(func() {
		if server == nil {
			server = &tcpServer{
				decodeHandler: handleDecodeGob,
				encodeHandler: handleEncodeGob,
			}
		}
	})
	return server
}

// Listen starts listening on the endpoint port on all interfaces.
func (e *tcpServer) Listen() error {
	var err error
	e.listener, err = net.Listen("tcp", common.Port)
	if err != nil {
		log.Printf("Unable to listen on port %s\n", common.Port)
		return err
	}
	log.Println("Listen on", e.listener.Addr().String())
	log.Printf("Server started on Port=%s \n", common.Port)

	for {
		log.Println("Accepting a connection request...")
		conn, err := e.listener.Accept()
		if err != nil {
			log.Println("Failed accepting a connection request:", err)
			continue
		}
		log.Println("Handle incoming connection.")

		// one conn one gorutine and write message to channel
		go e.handleMessages(conn)
	}
}

// handleMessages reads the connection up to the first newline.
func (e *tcpServer) handleMessages(conn net.Conn) {
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	defer conn.Close()

	// create session for current conn
	session := common.NewSession(conn, nil, rw)
	for {
		log.Println("Decode handleMessages ....")
		err := e.decodeHandler(session, rw.Reader)
		if err != nil {
			log.Println("Decode handleMessages error , close the net.Conn later by defer ", err)
			return
		}
	}
}

// decode gob message from buffer
func handleDecodeGob(session *common.Session, r *bufio.Reader) error {

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

	//log.Printf("Handle DecodeGob protoId=%d , empty data= %#v\n", protoId, data)

	dec := gob.NewDecoder(r)
	err := dec.Decode(data)

	if err != nil {
		log.Printf("Error decoding GOB data \n%#v:\n", err)
		return err
	}

	log.Printf("Handle DecodeGob protoId=%d , data= %#v\n", protoId, data)

	switch protoId {
	case proto.CG_LonginMessageId:
		// push data to channel
		msgWrap := common.NewMsgWrap(session, data)
		QueueServiceIns().PushLoginMsg(session, msgWrap)

	default:
		// sample do other things.
		GetActionManagerIns().Handle(data, session)
	}

	return nil
}

// encode data to io buffer and send it
func handleEncodeGob(w *bufio.Writer, data proto.ProtoMessage) error {
	log.Print("Handle EncodeGob:")

	protoId := data.GetProtoId()
	buf := make([]byte, 16)
	binary.BigEndian.PutUint16(buf, protoId)
	w.Write(buf)

	enc := gob.NewEncoder(w)
	err := enc.Encode(data)
	if err != nil {
		log.Println("Error encoding GOB data:", err)
		return err
	}
	log.Printf("Encodeed ProtoMessage : \n%#v\n", data)

	e := w.Flush()
	return e
}

func BootServer() error {
	serverInstance := TcpServerIns()

	// init action handler
	GetActionManagerIns()
	// init QueueService
	QueueServiceIns()
	//
	//AccountServiceIns()

	// Start listening.
	return serverInstance.Listen()
}
