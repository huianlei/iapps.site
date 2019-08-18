package network

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"iapps.site/proto"
)

const (
	// for server mode
	ActionTypeForServer = uint8(1)
	// for client mode
	ActionTypeForClient = uint8(2)
)

// ---------------------------

// ---------------------------

type ActionHandler func(proto.ProtoMessage, interface{}) error

type actionManager struct {
	actions map[uint16]ActionHandler
}

// action work type default server
var ActionType uint8 = ActionTypeForServer

// singletone
var aminstance *actionManager = nil
var am_once sync.Once

func GetActionManagerIns() *actionManager {
	am_once.Do(func() {
		if aminstance == nil {
			log.Println("== construct singleton of actionManager")
			aminstance = &actionManager{
				actions: make(map[uint16]ActionHandler, 4),
			}
			switch ActionType {
			case ActionTypeForServer:
				aminstance.intServer()
			case ActionTypeForClient:
				aminstance.initClient()
			}
		}
	})

	log.Printf("get actionManager instance %v", aminstance)
	return aminstance
}

// ==============================================================================

func (am *actionManager) initClient() {
	log.Println("Registing action handlers for client mode")
	am.register(proto.GC_LonginMessageId, handle_GC_LoginMessage)
}

func (am *actionManager) intServer() {
	log.Println("Registing action handlers for server mode")
	am.register(proto.CG_LonginMessageId, handle_CG_LoginMessage)
}

// register ActionHandler by messageId
func (manager *actionManager) register(messageId uint16, handler ActionHandler) {
	if _, ok := manager.actions[messageId]; !ok {
		manager.actions[messageId] = handler
		log.Printf("register action handler messageId=%d, ok=%v \n", messageId, ok)
	}
}

// do handle
func (manager *actionManager) Handle(data proto.ProtoMessage, obj interface{}) error {
	if obj == nil {
		return errors.New("Invoke action Handle,but obj can not be nil")
	}
	if data == nil {
		return errors.New("Invoke action Handle, but data can not be nil")
	}

	messageId := data.GetProtoId()
	log.Printf("Handle messageId=%d,[%s] \n", messageId, data)
	handler, ok := manager.actions[messageId]
	if !ok {
		errMsg := fmt.Sprintf("Invoke action Handle but handler not found, messageId=%d", messageId)
		log.Println(errMsg)
		return errors.New(errMsg)
	}

	//log.Printf("Before call handler messageId=%d, %s\n", messageId, handler)
	err := handler(data, obj)

	if err != nil {
		//log.Printf("After call handler messageId=%d, error %#v\n", messageId, err)
	}

	return err
}
