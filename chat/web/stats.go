package web

import (
	"io"
	"log"
	"sync"
	"time"
)

type statInfo struct {
	requestReceived int
	msgPushed       int
	requestWithErr  int
}

type stat struct {
	mu         sync.Mutex
	totalInfo  statInfo
	minuteInfo statInfo
	logger     *log.Logger
}

func newStat(w io.Writer) *stat {
	return &stat{
		logger: log.New(w, "", log.LstdFlags),
	}
}

func (s *stat) addRequestReceived(cnt int, errCnt int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalInfo.requestReceived += cnt
	s.minuteInfo.requestReceived += cnt
	s.totalInfo.requestWithErr += errCnt
	s.minuteInfo.requestWithErr += errCnt
}

func (s *stat) addMessagePushed(cnt int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalInfo.msgPushed += cnt
	s.minuteInfo.msgPushed += cnt
}

func (s *stat) printStat() {
	s.mu.Lock()
	defer s.mu.Unlock()
	format := "total: requests received %d(error %d), messages pushed %d," +
		"last minute: requests received %d(error %d), messages pushed %d"
	s.logger.Printf(format,
		s.totalInfo.requestReceived, s.totalInfo.requestWithErr, s.totalInfo.msgPushed,
		s.minuteInfo.requestReceived, s.minuteInfo.requestWithErr, s.minuteInfo.msgPushed)
}

func (s *stat) resetMinuteStat() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.minuteInfo.requestReceived = 0
	s.minuteInfo.requestWithErr = 0
	s.minuteInfo.msgPushed = 0
}

func (s *stat) runStat() {
	for {
		time.Sleep(1 * time.Minute)
		s.printStat()
		s.resetMinuteStat()
	}
}
