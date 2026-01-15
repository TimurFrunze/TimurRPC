package logInterface

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/gogf/gf/v2/container/gqueue"
	"gopkg.in/natefinch/lumberjack.v2"
)

var defaultBatchSize int = 10000

type LoggerInterface interface {
	Log(msg string) (_err_ error)
}

type TestLogger struct {
}

func NewTestLogger() *TestLogger {
	return &TestLogger{}
}

func (l *TestLogger) Log(msg string) error {
	fmt.Println(msg)
	return nil
}

type Signal struct{}

type TimurLogger struct {
	batchSize  int
	logger     *lumberjack.Logger
	fileWriter *log.Logger
	q          *gqueue.Queue
	isClosed   atomic.Bool
	signals    chan Signal
	guard      sync.WaitGroup //发出结束申请后，需要等待后台协程结束后才可以关闭和清理资源
}

// func writeAll(result *TimurLogger, signal Signal, sign bool) {
// 	//没有关闭则取出全部日志并逐一写入
// 	for {
// 		logMsg := result.q.Pop()
// 		if logMsg == nil {
// 			break
// 		}
// 		if msg, ok := logMsg.(string); ok {
// 			result.fileWriter.Println(msg)
// 		}
// 		if result.q.Len() == 0 {
// 			break
// 		}
// 	}
// 	//将队列中的信号全部清空，然后再放入一个信号，防止因为通道满，无法继续添加新信号而无法持续触发
// 	for {
// 		select {
// 		case <-result.signals:
// 			continue
// 		default:
// 			func() {
// 				if sign {
// 					result.signals <- signal
// 				}
// 			}()
// 			return
// 		}
// 	}
// }

func (l *TimurLogger) processSignal() bool {
	//非阻塞函数，信号通道不满时直接返回false，只有信号通道满时返回true
	for i := 0; i < l.batchSize; i++ {
		select {
		case <-l.signals:
			continue
		default:
			return false
		}
	}
	return true
}

func (l *TimurLogger) Close() (_err_ error) {
	defer func() {
		if r := recover(); r != nil {
			_err_ = fmt.Errorf("recover from panic: %v while closing logger", r)
		}
	}()
	//先将isClosed设为false并通过信号通道激活提醒
	l.isClosed.Store(true)

	//等待后台协程处理结束
	l.guard.Wait()
	//清理资源
	close(l.signals)
	l.q.Close()
	return l.logger.Close()
}

func (l *TimurLogger) Log(msg string) (_err_ error) {
	defer func() {
		if r := recover(); r != nil {
			_err_ = fmt.Errorf("recover from panic: %v while logging", r)
		}
	}()
	//日志已经关闭，返回错误
	if l.isClosed.Load() {
		return fmt.Errorf("logger is closed")
	}
	//正常添加日志并告知后台协程
	l.q.Push(msg)
	if len(l.signals) < cap(l.signals) {
		l.signals <- Signal{}
	}
	return nil
}

func NewTimurLogger(filename string, maxSizeMB int, maxBackups int, maxAgeDays int, compress bool, localTime bool, batchSize int) (timur *TimurLogger, _err_ error) {
	defer func() {
		if r := recover(); r != nil {
			_err_ = fmt.Errorf("recover from panic: %v", r)
			timur = nil
		}
	}()
	logger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   compress,
		LocalTime:  localTime,
	}
	fileWriter := log.New(logger, "", log.LstdFlags)
	q := gqueue.New()
	if q == nil {
		return nil, fmt.Errorf("queue is nil")
	}
	if batchSize < defaultBatchSize {
		batchSize = defaultBatchSize
	}
	result := &TimurLogger{
		batchSize:  batchSize,
		logger:     logger,
		fileWriter: fileWriter,
		q:          q,
		isClosed:   atomic.Bool{},
		signals:    make(chan Signal, batchSize),
		guard:      sync.WaitGroup{},
	}
	result.isClosed.Store(false)
	result.guard.Add(1)
	result.signals <- Signal{} //添加初始信号，触发后台协程运行
	//创建后台异步写入协程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("logger background goroutine recover from panic: %v", r)
				fileWriter.Printf("logger background goroutine recover from panic: %v", r)
			}
			result.guard.Done()
		}()
		for {
			// signal := <-result.signals
			//查看是否已经关闭
			if result.isClosed.Load() {
				//关闭则进行最后一次检查和清理
				// writeAll(result, signal, false)
				break
			}
			//批量处理全部信号，但是只有在信号占满时才再移除后再加入一个信号，防止因为通道长度有限而漏处理加入的日志
			// flag := true
			for i := 0; i < result.batchSize; i++ {
				select {
				case <-result.signals:

				}
			}
		}
	}()
	return result, nil
}
