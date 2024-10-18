package model

import (
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type LogicWsProxySession struct {
	fromWsConn    *websocket.Conn
	toWsConn      *websocket.Conn
	recvIdxRegMsg int
}

func NewLogicWSProxySession(from *websocket.Conn, to *websocket.Conn) (*LogicWsProxySession, error) {
	return &LogicWsProxySession{
		fromWsConn:    from,
		toWsConn:      to,
		recvIdxRegMsg: 1, //default 1, if receive reg message, set to 0
	}, nil
}

//Close 关闭
func (sws *LogicWsProxySession) Close() {
	// toWsConn.Close()
	// fromWsConn.Close()
}

func (sws *LogicWsProxySession) Start(quitChan chan bool) {
	go sws.proxyUpWsMsg(quitChan)
	go sws.proxyDownWsMsg(quitChan)
}

//proxy from lowlevel to uplevel ws
func (sws *LogicWsProxySession) proxyUpWsMsg(exitCh chan bool) {
	logrus.Println("setup proxyUpWsMsg")
	//tells other go routine quit
	defer setQuit(exitCh)
	for {
		//read websocket msg
		mt, wsData, err := sws.fromWsConn.ReadMessage()
		if err != nil {
			logrus.WithError(err).Error("reading device message failed")
			return
		}

		if sws.recvIdxRegMsg > 0 && sws.recvIdxRegMsg < 5 {
			if sws.recvIdxRegMsg == 1 && strings.HasPrefix(string(wsData), "{\"cmd\":\"reg\",") {
				sws.recvIdxRegMsg = 0
			} else {
				sws.recvIdxRegMsg++
			}
		}

		err = sws.toWsConn.WriteMessage(mt, wsData)
		logrus.Debugln("proxy message to uplevel:", string(wsData))
		if err != nil {
			logrus.WithError(err).Error("proxy message to uplevel:")
			return
		}
	}
}

//proxy from uplevel to lowlevel ws
func (sws *LogicWsProxySession) proxyDownWsMsg(exitCh chan bool) {
	logrus.Println("setup proxyDownWsMsg")
	//tells other go routine quit
	defer setQuit(exitCh)
	for {
		//read websocket msg
		mt, wsData, err := sws.toWsConn.ReadMessage()
		if err != nil {
			logrus.WithError(err).Error("read uplevel message failed")
			return
		}
		if string(wsData) == "==ping==" {
			logrus.Debugln("==ping==")
			continue
		}

		err = sws.fromWsConn.WriteMessage(mt, wsData)
		logrus.Debugln("proxy message to device:", string(wsData))
		if err != nil {
			logrus.WithError(err).Error("proxy message to device")
			return
		}
	}
}

func (sws *LogicWsProxySession) CheckReg(quitChan chan bool, milliseconds int) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
	if sws.recvIdxRegMsg == 0 {
		logrus.Println(milliseconds, "  milliseconds later, reg command received.")
	} else {
		logrus.Error(milliseconds, " milliseconds later, no reg command received,", sws.recvIdxRegMsg, " msg packages")
		setQuit(quitChan)
	}
}

func setQuit(ch chan bool) {
	ch <- true
}
