package web

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStats(t *testing.T) {
	buff := &bytes.Buffer{}
	stat := newStat(buff)

	stat.addRequestReceived(2, 1)
	stat.addMessagePushed(1)

	stat.addRequestReceived(3, 0)
	stat.addMessagePushed(2)

	assert.Equal(t, stat.totalInfo.requestReceived, 5)
	assert.Equal(t, stat.totalInfo.requestWithErr, 1)
	assert.Equal(t, stat.totalInfo.msgPushed, 3)
	assert.Equal(t, stat.minuteInfo.requestReceived, 5)
	assert.Equal(t, stat.minuteInfo.requestWithErr, 1)
	assert.Equal(t, stat.minuteInfo.msgPushed, 3)

	stat.resetMinuteStat()
	assert.Equal(t, stat.totalInfo.requestReceived, 5)
	assert.Equal(t, stat.totalInfo.requestWithErr, 1)
	assert.Equal(t, stat.totalInfo.msgPushed, 3)
	assert.Equal(t, stat.minuteInfo.requestReceived, 0)
	assert.Equal(t, stat.minuteInfo.requestWithErr, 0)
	assert.Equal(t, stat.minuteInfo.msgPushed, 0)

	stat.addRequestReceived(3, 1)
	stat.addMessagePushed(2)
	assert.Equal(t, stat.totalInfo.requestReceived, 8)
	assert.Equal(t, stat.totalInfo.requestWithErr, 2)
	assert.Equal(t, stat.totalInfo.msgPushed, 5)
	assert.Equal(t, stat.minuteInfo.requestReceived, 3)
	assert.Equal(t, stat.minuteInfo.requestWithErr, 1)
	assert.Equal(t, stat.minuteInfo.msgPushed, 2)

	stat.printStat()
	logContent := buff.String()
	totalC := strings.Contains(logContent, "total: requests received 8(error 2), messages pushed 5")
	minuteC := strings.Contains(logContent, "last minute: requests received 3(error 1), messages pushed 2")
	assert.True(t, totalC)
	assert.True(t, minuteC)
}
