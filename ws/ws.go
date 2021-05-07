package ws

import (
	"errors"
	mapset "github.com/deckarep/golang-set"
	"github.com/fasthttp/websocket"
	"github.com/sirupsen/logrus"
	logger2 "github.com/thss-cercis/cercis-server/logger"
	"sync"
	"time"
)

var sessionMapper = make(map[string]*ConnWrapper)
var userMapper = make(map[int64]mapset.Set)
var rwMutex = sync.RWMutex{}

var logFields = logrus.Fields{
	"module":     "websocket",
	"middleware": false,
}

// PutConn 加入新的 ConnWrapper
func PutConn(sessionID string, userID int64, conn *websocket.Conn) *ConnWrapper {
	logger := logger2.GetLogger()
	// 关闭原来的
	rwMutex.Lock()
	defer rwMutex.Unlock()
	if sessionMapper[sessionID] != nil {
		if !sessionMapper[sessionID].isClosed {
			_ = sessionMapper[sessionID].Close()
		}
	}
	newWrapper := New(sessionID, userID, conn)
	// add sessionMapper
	sessionMapper[sessionID] = newWrapper
	// add userMapper
	if userMapper[userID] == nil {
		userMapper[userID] = mapset.NewSet()
		userMapper[userID].Add(newWrapper)
	} else {
		userMapper[userID].Add(newWrapper)
	}

	logger.WithFields(logFields).Debugf("Add new ws conn for session %v", sessionID)
	return newWrapper
}

// DelConn 删除 ConnWrapper，找不到也返回 nil
func DelConn(sessionID string) error {
	logger := logger2.GetLogger()
	rwMutex.Lock()
	defer rwMutex.Unlock()
	conn := sessionMapper[sessionID]
	if conn != nil {
		_ = conn.Close()
		// 更改 sessionMapper
		delete(sessionMapper, sessionID)
		// 更改 userMapper
		userMapper[conn.UserID].Remove(conn)
		if userMapper[conn.UserID].Cardinality() == 0 {
			delete(userMapper, conn.UserID)
		}
	}

	logger.WithFields(logFields).Debugf("Remove ws conn for session %v", sessionID)
	return nil
}

// GetConn 获得 ConnWrapper，找不到则返回 nil
func GetConn(sessionID string) *ConnWrapper {
	logger := logger2.GetLogger()
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	logger.WithFields(logFields).Debugf("Get ws conn for session %v", sessionID)
	return sessionMapper[sessionID]
}

// GetConnByUserID 获得某个 user 的全部 conn
func GetConnByUserID(userID int64) []*ConnWrapper {
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	if userMapper[userID] == nil {
		return nil
	}
	ret := make([]*ConnWrapper, 0)
	for ele := range userMapper[userID].Iter() {
		elem, ok := ele.(*ConnWrapper)
		if ok {
			ret = append(ret, elem)
		}
	}
	return ret
}

/*******************
 ** ConnWrapper 区域
 *******************/

type ConnWrapper struct {
	SessionID  string
	UserID     int64
	conn       *websocket.Conn
	isClosed   bool
	closeMutex sync.Mutex
	ch         chan interface{}
}

func New(sessionID string, userID int64, conn *websocket.Conn) *ConnWrapper {
	return &ConnWrapper{
		SessionID: sessionID,
		UserID:    userID,
		conn:      conn,
		isClosed:  false,
		ch:        make(chan interface{}, 1024),
	}
}

// Start 开始遍历消息队列
func (wrapper *ConnWrapper) Start() {
	logger := logger2.GetLogger()
	logger.WithFields(logFields).Debugf("Start looping conn for session %v", wrapper.SessionID)

	go func() {
		t := time.NewTicker(15 * time.Second)
		for range t.C {
			_ = wrapper.Write(&struct {
				Type int64  `json:"type"`
				Msg  string `json:"msg"`
				Time int64  `json:"time"`
			}{
				Type: 1,
				Msg:  "nmsl",
				Time: time.Now().Unix(),
			})
		}
	}()

	go func() {
		for {
			if _, _, err := wrapper.conn.NextReader(); err != nil {
				break
			}
		}
	}()

	ticker := time.NewTicker(30 * time.Second)
LabelFor:
	for !wrapper.isClosed {
		select {
		case <-ticker.C:
			if err := wrapper.conn.WriteControl(websocket.PingMessage, []byte("heartbeat"), time.Now().Add(5*time.Second)); err != nil {
				_ = wrapper.Close()
				_ = DelConn(wrapper.SessionID)
				break LabelFor
			}
			logger.WithFields(logFields).Tracef("Heartbeat sent to session %v", wrapper.SessionID)
			break
		case jsonObj := <-wrapper.ch:
			if jsonObj == nil {
				break LabelFor
			}
			if err := wrapper.conn.WriteJSON(jsonObj); err != nil {
				_ = wrapper.Close()
				_ = DelConn(wrapper.SessionID)
				break LabelFor
			}
			logger.WithFields(logFields).Tracef("Json obj %v sent to session %v", jsonObj, wrapper.SessionID)
		}
	}

}

// Write 向消息队列中写入 json object 的指针
func (wrapper *ConnWrapper) Write(v interface{}) error {
	wrapper.closeMutex.Lock()
	defer wrapper.closeMutex.Unlock()
	if wrapper.isClosed {
		return errors.New("connection is closed")
	}
	wrapper.ch <- v

	logger := logger2.GetLogger()
	logger.WithFields(logFields).Debugf("Write msg for session %v", wrapper.SessionID)
	return nil
}

// Close 关闭连接，返回 conn.Close 的错误
func (wrapper *ConnWrapper) Close() error {
	wrapper.closeMutex.Lock()
	defer wrapper.closeMutex.Unlock()
	if wrapper.isClosed {
		return errors.New("ConnWrapper is already closed")
	}
	wrapper.isClosed = true
	close(wrapper.ch)

	logger := logger2.GetLogger()
	logger.WithFields(logFields).Debugf("Close conn for session %v", wrapper.SessionID)
	return wrapper.conn.Close()
}
