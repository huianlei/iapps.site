package network

import (
	"errors"
	"fmt"
	"log"
	"time"

	"iapps.site/common"
	"iapps.site/proto"
)

func handle_CG_LoginMessage(data proto.ProtoMessage, obj interface{}) error {
	log.Printf("serverHandle_CG_LoginMessage data=[%#v]", data)
	cglm, ok := data.(*proto.CG_LonginMessage)
	if !ok {
		log.Printf("serverHandle_CG_LoginMessage data not a *proto.CG_LonginMessage")
		return errors.New(fmt.Sprintf("data not a proto.CG_LonginMessage"))
	}

	session, ok := obj.(*common.Session)
	if !ok {
		log.Printf("serverHandle_CG_LoginMessage obj not a *session")
		return errors.New(fmt.Sprintf("obj not a Session"))
	}

	log.Printf("serverHandle_CG_LoginMessage handle logic")

	// just sleep a while, to simulate validating token from third platform (eg. facebook.com).
	time.Sleep(time.Duration(common.ValidateTokenSleep * time.Millisecond.Nanoseconds()))

	//
	validateToken := true

	// feedback login result
	roleId := int64(time.Now().Nanosecond())

	var retCode int8 = proto.LONGIN_RET_OK
	var errMsg string = ""
	if !validateToken {
		retCode = proto.LONGIN_RET_INVALID_TOKEN
		errMsg = "Oauth token invalid!"
	}

	// after login manager player
	player := common.NewPlayer(roleId, cglm.AccId)
	common.PlayerManagerIns().AddPlayerOnLogin(player)

	ServerLoginBack(cglm, session, roleId, retCode, errMsg, 0)

	return nil
}

// common login back if login not ok
func ServerLoginBack(cglm *proto.CG_LonginMessage, session *common.Session, roleId int64,
	retCode int8, errMsg string, queueIndex int32) {

	respData := proto.GC_LonginMessage{
		ProtoId:    proto.GC_LonginMessageId,
		AccId:      cglm.AccId,
		RetCode:    retCode,
		ErrMsg:     errMsg,
		RoleId:     roleId,
		QueueIndex: queueIndex,
	}

	log.Printf("ServerLoginBack accId=%s, roleId=%d send response: \n%#v\n", cglm.AccId, roleId, respData)
	handleEncodeGob(session.ReaderWriter().Writer, respData)
}

func ServerLoginBackFail(cglm *proto.CG_LonginMessage, session *common.Session, retCode int8, errMsg string) {
	ServerLoginBack(cglm, session, 0, retCode, errMsg, 0)
}
