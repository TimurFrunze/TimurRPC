package logInterface

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
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

var logPath string = "testLog01.log"

func TestWriteLog(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("recover from panic: %v", r)
		}
	}()
	var timur *TimurLogger
	var err error
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("recover from panic: %v", r)
		}
	}()
	timur, err = NewTimurLogger(logPath, 50, 5, 28, true, true, 10000, 10)
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
	time.Sleep(100 * time.Millisecond)
}

func TestLogContent(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("recover from panic: %v", r)
		}
	}()
	//验证上一个函数的日志内容是否正确
	content, err := os.ReadFile(logPath)
	//验证文件存在和内容正确
	if err != nil {
		var msg string
		msg = fmt.Sprintf("ReadFile failed, with err != nil: %v", err)
		t.Error(msg)
		panic(msg)
	}
	if len(content) == 0 {
		var msg string
		msg = "log is empty"
		t.Error(msg)
		panic(msg)
	} else {
		fmt.Println(string(content))
	}
}

var asyncTestPath string = "asyncTestLog.log"

var concurrentTestPath string = "concurrentTestLog.log"

var numGoroutines int = 10000

var asyncTestSize int = 1000

func TestConcurrentWrite(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("recover from panic: %v", r)
		}
	}()
	//创建日志模块
	timur, err := NewTimurLogger(concurrentTestPath, 50, 5, 28, true, true, 10000, 10)
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
	wg := sync.WaitGroup{}
	wg.Add(numGoroutines)
	//创建并发协程写入
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("index %d, recover from panic: %v", i, r)
				}
				wg.Done()
			}()
			err := timur.Log(fmt.Sprintf("async log with index %d", i))
			if err != nil {
				var msg string
				msg = fmt.Sprintf("Log failed with index %d, with err != nil: %v", i, err)
				t.Error(msg)
				panic(msg)
			}
			//睡眠随机时间后结束
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		}()
	}
	wg.Wait()
}

// 检验并发写入是否成功
func TestConcurrentResult(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("recover from panic: %v", r)
		}
	}()
	content, err := os.ReadFile(concurrentTestPath)
	if err != nil {
		var msg string
		msg = fmt.Sprintf("ReadFile failed, with err != nil: %v", err)
		t.Error(msg)
		panic(msg)
	}
	if len(content) == 0 {
		var msg string
		msg = "log is empty"
		t.Error(msg)
		panic(msg)
	} else {
		fmt.Println("concurrent log content:", string(content))
	}
}

// 测试异步写入，生产者协程写入后是否快速返回
func TestAsyncWrite(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("recover from panic: %v", r)
		}
	}()
	//创建日志模块
	timur, err := NewTimurLogger(asyncTestPath, 50, 5, 28, true, true, 10000, 10)
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
	//检验是否快速返回
}
