package logInterface

import (
	"fmt"
	"testing"
)

//测试日志模块是否可以正常打开关闭

func TestInitClose(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("recover from panic: %v", r)
		}
	}()
	timur, err := NewTimurLogger("testLog.log", 50, 5, 28, true, true, 10000, 10)
	if timur == nil {
		var msg string
		msg = fmt.Sprintf("NewTimurLogger failed, with timur == nil: %v", err)
		panic(msg)
	}
	if err != nil {
		var msg string
		msg = fmt.Sprintf("NewTimurLogger failed, with err != nil: %v", err)
		panic(msg)
	}
	defer func() {
		err = timur.Close()
		if err != nil {
			var msg string
			msg = fmt.Sprintf("Close failed, with err != nil: %v", err)
			panic(msg)
		}
	}()
}
