// common.go
package common

import (
	"bufio"
	"hash/fnv"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"iapps.site/proto"
)

// =======================================================================
// constant config start
// you can modify these const
// =======================================================================
// PlayerManager const config
const (
	// max online player count
	MaxOnline int32 = 20000
)

// QueueService const config
const (
	// tick milliseconds
	TickInterval = int64(100)
	// queue service channel capacity
	QueueCapacity = int(10000)
	// check interval to broadcast position in queue. in seconds
	QueueCheckInterval = int64(2)
)

// TcpServer const
const (
	// Listen port
	Port = ":9001"
	// simulate validate token sleep time in milliseconds
	ValidateTokenSleep = int64(100)
)

// =======================================================================
// constant config end
// =======================================================================

// sessionId start from 1 , do not modify
var _suid uint64 = uint64(0)

// global guid, do not modify
var _global_guid = uint64(0)

type Session struct {
	conn net.Conn
	obj  interface{}
	rw   *bufio.ReadWriter
	id   uint64
}

func NewSession(conn net.Conn, obj interface{}, rw *bufio.ReadWriter) *Session {
	id := atomic.AddUint64(&_suid, 1)
	return &Session{
		conn, obj, rw, id,
	}
}

func (s *Session) Conn() net.Conn {
	return s.conn
}

func (s *Session) Obj() interface{} {
	return s.obj
}

func (s *Session) Id() uint64 {
	return s.id
}

func (s *Session) ReaderWriter() *bufio.ReadWriter {
	return s.rw
}

type MsgWrap struct {
	session *Session
	data    proto.ProtoMessage
}

func NewMsgWrap(session *Session, data proto.ProtoMessage) *MsgWrap {
	return &MsgWrap{
		session, data,
	}
}

func (m *MsgWrap) Session() *Session {
	return m.session
}

func (m *MsgWrap) Data() proto.ProtoMessage {
	return m.data
}

// -----------------------------------------------------------
// utils
func GetNowTimeMilli() int64 {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	return now
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func NextGuid() uint64 {
	return atomic.AddUint64(&_global_guid, 1)
}

// -----------------------------------------------------------
// just a sample
type Player struct {
	roleId int64
	accId  string
}

func (p *Player) RoleId() int64 {
	return p.roleId
}

func (p *Player) AccId() string {
	return p.accId
}

func NewPlayer(roleId int64, accId string) *Player {
	return &Player{roleId, accId}
}

//

type playerManager struct {
	playermap   sync.Map
	onlineCount int32
}

// singletone
var pm_instance *playerManager = nil
var pm_once sync.Once

func PlayerManagerIns() *playerManager {
	pm_once.Do(func() {
		if pm_instance == nil {
			log.Println("== construct singleton of playerManager")
			pm_instance = &playerManager{}
		}
	})

	return pm_instance
}

// sample
func (m *playerManager) AddPlayerOnLogin(p *Player) bool {
	if p == nil || p.roleId <= 0 {
		return false
	}
	// ok false means store success
	_, ok := m.playermap.LoadOrStore(p.roleId, p)
	if !ok {
		online_count := atomic.AddInt32(&m.onlineCount, 1)
		log.Printf("add player success on login roleId=%d, online_count=%d \n", p.roleId, online_count)
	}
	return !ok
}

// sample
func (m *playerManager) DelPlayerOnLogout(p *Player) {
	if p == nil || p.roleId <= 0 {
		return
	}

	_, findOk := m.playermap.Load(p.roleId)
	if findOk {
		m.playermap.Delete(p.roleId)
		online_count := atomic.AddInt32(&m.onlineCount, -1)
		// not happen
		if online_count < 0 {
			atomic.CompareAndSwapInt32(&m.onlineCount, online_count, 0)
		}
		log.Printf("delete player success on logout roleId=%d, online_count=%d \n", p.roleId, online_count)
	}
}

func (m *playerManager) OnlineCount() int32 {
	return atomic.LoadInt32(&m.onlineCount)
}

func (m *playerManager) IsMaxOnline() bool {
	c := m.OnlineCount()

	isMax := (c >= MaxOnline)
	if isMax {
		//
	}

	return isMax
}
