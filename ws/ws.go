package ws

import (
	"errors"
	"github.com/fasthttp/websocket"
	"github.com/sirupsen/logrus"
	logger2 "github.com/thss-cercis/cercis-server/logger"
	"sync"
)

var wsMapper = make(map[string]*ConnWrapper)
var lock = sync.RWMutex{}

var logFields = logrus.Fields{
	"middleware": false,
	"module":     "websocket",
}

// PutConn 加入新的 ConnWrapper
func PutConn(sessionID string, conn *websocket.Conn) *ConnWrapper {
	logger := logger2.GetLogger()
	// 关闭原来的
	lock.Lock()
	defer lock.Unlock()
	if wsMapper[sessionID] != nil {
		_ = wsMapper[sessionID].Close()
	}
	newWrapper := New(sessionID, conn)
	wsMapper[sessionID] = newWrapper

	logger.WithFields(logFields).Debugf("Add new ws conn for session %v", sessionID)
	return newWrapper
}

// DelConn 删除 ConnWrapper，找不到也返回 nil
func DelConn(sessionID string) error {
	logger := logger2.GetLogger()
	lock.Lock()
	defer lock.Unlock()
	var err error = nil
	if wsMapper[sessionID] != nil {
		err = wsMapper[sessionID].Close()
		delete(wsMapper, sessionID)
	}

	logger.WithFields(logFields).Debugf("Remove ws conn for session %v", sessionID)
	return err
}

// GetConn 获得 ConnWrapper，找不到则返回 nil
func GetConn(sessionID string) *ConnWrapper {
	logger := logger2.GetLogger()
	lock.RLock()
	defer lock.RUnlock()

	logger.WithFields(logFields).Debugf("Get ws conn for session %v", sessionID)
	return wsMapper[sessionID]
}

/*******************
 ** ConnWrapper 区域
 *******************/

type ConnWrapper struct {
	sessionID string
	conn      *websocket.Conn
	isClosed  bool
	// some else
	ch chan interface{}
}

func New(sessionID string, conn *websocket.Conn) *ConnWrapper {
	return &ConnWrapper{
		sessionID: sessionID,
		conn:      conn,
		isClosed:  false,
		ch:        make(chan interface{}, 1024),
	}
}

// Start 开始遍历消息队列
func (wrapper *ConnWrapper) Start() {
	logger := logger2.GetLogger()
	logger.WithFields(logFields).Debugf("Start looping conn for session %v", wrapper.sessionID)
	go func() {
		// defer
		defer func() {
			_ = wrapper.Close()
			lock.Lock()
			delete(wsMapper, wrapper.sessionID)
			lock.Unlock()
		}()

		for jsonObj := range wrapper.ch {
			if err := wrapper.conn.WriteJSON(jsonObj); err != nil {
				break
			}
		}
	}()
}

// Write 向消息队列中写入 json object 的指针
func (wrapper *ConnWrapper) Write(v interface{}) error {
	logger := logger2.GetLogger()
	if wrapper.isClosed {
		return errors.New("connection is closed")
	}
	wrapper.ch <- v

	logger.WithFields(logFields).Debugf("Write msg for session %v", wrapper.sessionID)
	return nil
}

// Close 关闭连接，返回 conn.Close 的错误
func (wrapper *ConnWrapper) Close() error {
	logger := logger2.GetLogger()
	wrapper.isClosed = true
	close(wrapper.ch)

	logger.WithFields(logFields).Debugf("Close conn for session %v", wrapper.sessionID)
	return wrapper.conn.Close()
}
