package network

import (
	"errors"
	"fmt"
	"log"
	"time"

	"iapps.site/common"
	"iapps.site/proto"
)

// handle server response GC_LoginMessage
func handle_GC_LoginMessage(data proto.ProtoMessage, obj interface{}) error {
	log.Printf("Invoke clientHandle_GC_LoginMessage %#v", data)
	session, ook := obj.(*common.Session)
	if !ook {
		return errors.New(fmt.Sprintf("obj not a *common.Session"))
	}
	client, ok := session.Obj().(*TcpClient)
	if !ok {
		return errors.New(fmt.Sprintf("session.Obj() not a *TcpClient"))
	}
	gclm, ok := data.(*proto.GC_LonginMessage)
	if !ok {
		return errors.New(fmt.Sprintf("data not a *proto.GC_LonginMessage"))
	}

	retCode := gclm.RetCode
	//
	switch retCode {
	// login ok
	case proto.LONGIN_RET_OK:
		client.FillAccount(gclm.RoleId)
		//client.Close()
		log.Printf("== received login ok clientId=%d, accId=%d, roleId=%d connecting ", client.id, gclm.AccId, gclm.RoleId)
		loopHearbeat()
	// in login queue
	case proto.LONGIN_RET_INQUEUE:
		log.Printf("received login in queue clientId=%d, accId=%s\n", client.id, gclm.AccId)
	// // invalid token
	// case proto.LONGIN_RET_INVALID_TOKEN:

	// // server full
	// case proto.LONGIN_RET_SERVER_FULL:
	// 	client.Close()

	// // queue full
	// case proto.LONGIN_RET_QUEUE_FULL:
	// 	client.Close()
	// error
	default:
		log.Printf("received login failed clientId=%d, accId=%s close it\n", client.id, gclm.AccId)
		client.Close()
	}

	return nil
}

func loopHearbeat() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			sendHeartBeat()
		}
	}
}

func sendHeartBeat() {
	// do nothing
}
