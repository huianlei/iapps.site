package network

import (
	"errors"
	"log"
	"runtime"
	"sync"
	"sync/atomic"

	"time"

	"iapps.site/common"
	"iapps.site/proto"
)

//
type Service interface {
	Start() error
	Tick(delta int64) error
	Stop() error
}

// QueueService implements Service
type queueService struct {
	// login queue
	msgQueue chan *common.MsgWrap
	flag     bool
	// check time for scan and sync queue index to users
	checkTime int64
	// used by atomic
	seq    uint64
	msgMap sync.Map
	// used by atomic
	finishSeq uint64
	ticker    *time.Ticker
}

type qMsg struct {
	msgWrap   *common.MsgWrap
	seq       uint64
	accId     string
	lastIndex uint64
}

// singletone
var qsinstance *queueService = nil
var qs_once sync.Once

func QueueServiceIns() *queueService {
	qs_once.Do(func() {
		if qsinstance == nil {
			log.Println("== construct singleton of queueService")
			now := common.GetNowTimeMilli()
			ch := make(chan *common.MsgWrap, common.QueueCapacity)
			qsinstance = &queueService{
				msgQueue:  ch,
				flag:      true,
				checkTime: now,
				seq:       0,
				ticker:    time.NewTicker(time.Duration(common.TickInterval * time.Millisecond.Nanoseconds())),
			}

			qsinstance.Start()
		}
	})

	//log.Printf("get queueService instance %v", aminstance)
	return qsinstance
}

// return QueueService channel queue
func (service *queueService) MsgQueue() chan *common.MsgWrap {
	return service.msgQueue
}

// start service
func (service *queueService) Start() error {
	// start a independent goroutine handle tick
	go func() {
		log.Println("QueueService tick goroutine running...")
		lastTickTime := common.GetNowTimeMilli()
		ticker := service.ticker
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				now := common.GetNowTimeMilli()
				delta := now - lastTickTime
				lastTickTime = now
				service.Tick(delta)
			}
		}
	}()

	//
	service.initLoginGoroutines()

	return nil
}

// implements interface Service
// replaced by atomic in the future
func (service *queueService) Stop() error {
	service.flag = false
	return nil
}

// implements interface Service
func (service *queueService) Tick(delta int64) error {
	//
	//log.Printf("queueService Tick time=%d \n", common.GetNowTimeMilli())
	service.onTickCheckQueue()

	// maybe do something other

	return nil
}

// start multi goroutine to process login
func (service *queueService) initLoginGoroutines() {
	routineCount := runtime.NumCPU()

	for i := 1; i <= routineCount; i++ {
		go service.processLogin(i)
	}
}

// fetch a login message from queue and process login
func (service *queueService) processLogin(id int) {
	log.Printf("QueueService processLogin goroutine_%d running...\n", id)
	for {
		isMaxOnline := common.PlayerManagerIns().IsMaxOnline()

		for isMaxOnline {
			log.Printf("QueueService processLogin goroutine_%d isMaxOnline=%v\n", id, isMaxOnline)
			time.Sleep(1 * time.Second)
			isMaxOnline = common.PlayerManagerIns().IsMaxOnline()
		}

		// blocking here if channel is empty
		msgWrap, ok := <-service.msgQueue
		if ok {
			log.Printf("QueueService processLogin goroutine_%d got msgWrap=%#v\n", id, msgWrap)
			accId, ok := getAccId(msgWrap)
			if ok {
				value, getok := service.msgMap.Load(accId)
				if getok {
					qmsg, _ := value.(*qMsg)
					// update finished seqId
					old := service.GetFinishedSeqId()
					_currentFinishSeq := qmsg.seq
					atomic.CompareAndSwapUint64(&service.finishSeq, old, _currentFinishSeq)
					log.Printf("update finished accId=%s ,seqId old=%d, newSeqId=%d \n", accId, old, _currentFinishSeq)

					// remove it
					service.msgMap.Delete(accId)
				}
			}

			session := msgWrap.Session()
			log.Printf("got msgWrap from queue protoId=%d, sessionId=%d\n",
				msgWrap.Data().GetProtoId(), session.Id())

			// handle message
			GetActionManagerIns().Handle(msgWrap.Data(), session)
		}
	}
}

// get next seq id
func (service *queueService) NextSeqId() uint64 {
	return atomic.AddUint64(&service.seq, 1)
}

// get latest finished seq
// player queue index = obj.seq - finishedSeq
func (service *queueService) GetFinishedSeqId() uint64 {
	return atomic.LoadUint64(&service.finishSeq)
}

// push login message to channel if the channel not full, otherwise return QUEUE_FULL
// this method called in other goroutines
func (service *queueService) PushLoginMsg(session *common.Session, msgWrap *common.MsgWrap) error {
	channel := service.msgQueue
	ch_len := len(channel)
	ch_cap := cap(channel)

	// checked before
	cglm, ok := msgWrap.Data().(*proto.CG_LonginMessage)

	// channel full
	if ch_len == ch_cap {
		ServerLoginBackFail(cglm, session, proto.LONGIN_RET_QUEUE_FULL, "queue full")
		session.Conn().Close()
		return nil
	}

	if ok {
		channel <- msgWrap

		qmsg := &qMsg{msgWrap, service.NextSeqId(), cglm.AccId, 0}
		service.msgMap.LoadOrStore(qmsg.accId, qmsg)
		log.Printf("PushLoginMsg to queueSerice channel ok ch_len=%d , data= %#v\n", len(channel), msgWrap.Data())

		return nil
	}

	// not happen
	return errors.New("push queue errr not a valid message")
}

// get accId from login msg
func getAccId(msgWrap *common.MsgWrap) (string, bool) {
	if msgWrap == nil {
		return "", false
	}
	cglm, ok := msgWrap.Data().(*proto.CG_LonginMessage)
	return cglm.AccId, ok
}

// check queue status and broadcast to all clients in queue
func (service *queueService) onTickCheckQueue() {
	now := common.GetNowTimeMilli()

	// channel empty
	if len(service.msgQueue) == 0 {
		service.checkTime = now
		return
	}

	// time reached
	pastTime := (now - service.checkTime) * time.Millisecond.Nanoseconds()
	if pastTime >= common.QueueCheckInterval*time.Second.Nanoseconds() {
		log.Printf("queueService onTickCheckQueue trigger broadcastQueueIndexs \n")
		service.checkTime = now
		service.broadcastQueueIndexs()
	}
}

// iterator all objs in the queue, then broadcast the position index named (queueIndex)
// queueIndex = obj.seq - finishedSeq
func (service *queueService) broadcastQueueIndexs() error {
	//
	nowFinishedSeq := service.GetFinishedSeqId()
	service.msgMap.Range(func(k interface{}, v interface{}) bool {
		if k != nil {
			_, ok := k.(string)
			if !ok {
				return false
			}
			value, vok := v.(*qMsg)
			if !vok {
				return false
			}

			// calculate queueIndex
			// eg: player got a seq 128 when be added to the queue, now then current finished seq 100.
			// so queue index is 28 = (128-100)
			index := value.seq - nowFinishedSeq
			indexChanged := (index != value.lastIndex)

			// only feedback when position changed
			if indexChanged {
				value.lastIndex = index
				queueIndex := int32(index & 0x7FFFFFFF)
				cglm, ook := value.msgWrap.Data().(*proto.CG_LonginMessage)
				session := value.msgWrap.Session()
				roleId, retCode, errMsg := int64(0), proto.LONGIN_RET_INQUEUE, "in login queue"
				if ook {
					// feedback client position in login queue
					ServerLoginBack(cglm, session, roleId, retCode, errMsg, queueIndex)
				}
			}
		}
		return true
	})

	return nil
}
