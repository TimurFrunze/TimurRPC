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

func TestWriteLog(t *testing.T) {
	var timur *TimurLogger
	var err error
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("recover from panic: %v", r)
		}
	}()
	timur, err = NewTimurLogger("testLog01.log", 50, 5, 28, true, true, 10000, 10)
	//验证日志模块可以正常创建
	if timur == nil {
		var msg string
		msg = fmt.Sprintf("NewTimurLogger failed, with timur == nil: %v", err)
		t.Error(msg)
		panic(msg)
	}
	if err != nil {
		var msg string
		msg = fmt.Sprintf("NewTimurLogger failed, with err != nil: %v", err)
		t.Error(msg)
		panic(msg)
	}
	defer func() {
		if _err_ := timur.Close(); _err_ != nil {
			var msg string
			msg = fmt.Sprintf("Close failed, with err != nil: %v", _err_)
			t.Error(msg)
			panic(msg)
		}
	}()
	//验证可以正常写入日志
	err = timur.Log("test log")
	if err != nil {
		var msg string
		msg = fmt.Sprintf("Log failed, with err != nil: %v", err)
		t.Error(msg)
		panic(msg)
	}
	err = timur.Info("test info")
	if err != nil {
		var msg string
		msg = fmt.Sprintf("Info failed, with err != nil: %v", err)
		t.Error(msg)
		panic(msg)
	}
	err = timur.Error("test error")
	if err != nil {
		var msg string
		msg = fmt.Sprintf("Error failed, with err != nil: %v", err)
		t.Error(msg)
		panic(msg)
	}
	err = timur.Debug("test debug")
	if err != nil {
		var msg string
		msg = fmt.Sprintf("Debug failed, with err != nil: %v", err)
		t.Error(msg)
		panic(msg)
	}
	err = timur.Warn("test warn")
	if err != nil {
		var msg string
		msg = fmt.Sprintf("Warn failed, with err != nil: %v", err)
		t.Error(msg)
		panic(msg)
	}
	//验证文件存在和内容正确

}
