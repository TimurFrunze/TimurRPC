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
var defaultCloseSignalSize int = 10

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

func tryPop(q *gqueue.Queue) (item any, ok bool) {
	select {
	case item, ok = <-q.C:
		return item, ok
	default:
		return nil, true

	}
}

type TimurLogger struct {
	batchSize       int
	closeSignalSize int
	logger          *lumberjack.Logger
	fileWriter      *log.Logger
	q               *gqueue.Queue
	isClosed        atomic.Bool
	signals         chan Signal
	closeSignal     chan Signal
	guard           sync.WaitGroup //发出结束申请后，需要等待后台协程结束后才可以关闭和清理资源
}

func writeAll(result *TimurLogger) (_ok_ bool, _err error) {
	/*
		_ok_ : 用来表示是否接收到停止信号，为false表示接收到停止信号
		_err : 用来表示是否发生了panic
		后台协程通过返回值判断是否为最后一轮处理，如果为最后一轮（返回值为false）则直接退出
	*/
	defer func() {
		if r := recover(); r != nil {
			_ok_ = false
			_err = fmt.Errorf("recover from panic: %v while writing all logs", r)
		}
	}()
	//closeSignal通道和signals通道都可以激活后台协程
	select {
	case <-result.closeSignal:
		//说明日志模块已经关闭，最终返回false
		//将全部的日志取出并写入文件
		for {
			//需要select实现非阻塞弹出
			logMsg, __ok := tryPop(result.q)
			if !__ok {
				//说明通道异常关闭，直接panic
				panic("logger queue channel is closed")
			}
			if logMsg == nil {
				break
			}
			if msg, ok := logMsg.(string); ok {
				result.fileWriter.Println(msg)
			}
		}
		return false, nil
	case sig, _sign := <-result.signals:
		if !_sign {
			//说明信号通道异常关闭，直接panic
			panic("logger signals channel is closed")
		}
		//将全部的日志取出并写入文件
		for {
			//需要select实现非阻塞弹出
			logMsg, __ok := tryPop(result.q)
			if !__ok {
				//说明通道异常关闭，直接panic
				panic("logger queue channel is closed")
			}
			if logMsg == nil {
				break
			}
			if msg, ok := logMsg.(string); ok {
				result.fileWriter.Println(msg)
			}
		}
		//将signals通道中的信号清除，并判断是否产生拥塞
		for i := 0; i < result.batchSize; i++ {
			select {
			case sig, _sign = <-result.signals:
				if !_sign {
					//后台协程还未结束，通道就被异常关闭
					panic("logger signals channel is closed")
				}
				continue
			default:
				//还未执行到batchSize次，说明信号队列没有发生拥塞
				return true, nil
			}
		}
		select {
		case result.signals <- sig:
		default:
		}
		return true, nil
	}
}

func (l *TimurLogger) Close() (_err_ error) {
	/*
		关闭流程：
		1.主协程将isClosed设为false，并添加信号到closeSignal通道，先禁止其他协程继续添加日志，再激活后台协程处理
		2.在此之后，主协程和其他协程无法继续添加日志
		3.主协程等待后台协程结束处理
	*/
	defer func() {
		if r := recover(); r != nil {
			_err_ = fmt.Errorf("recover from panic: %v while closing logger", r)
		}
	}()
	//先判断是否已经关闭
	if l.isClosed.Load() {
		return fmt.Errorf("logger is closed")
	}
	//先将isClosed设为true并通过信号通道激活提醒后台协程，isClosed给生产者看，closeSignal给后台协程看
	l.isClosed.Store(true)
	select {
	case l.closeSignal <- Signal{}:
	default:
	}
	//等待后台协程处理结束
	fmt.Println("close logger, waiting...")
	l.guard.Wait()
	fmt.Println("background goroutine closed")
	//清理资源
	close(l.signals)
	close(l.closeSignal)
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

func (l *TimurLogger) Info(msg string) error {
	return l.Log("INFO: " + msg)
}

func (l *TimurLogger) Debug(msg string) error {
	return l.Log("DEBUG: " + msg)
}

func (l *TimurLogger) Warn(msg string) error {
	return l.Log("WARN: " + msg)
}

func (l *TimurLogger) Error(msg string) error {
	return l.Log("ERROR: " + msg)
}

func NewTimurLogger(filename string, maxSizeMB int, maxBackups int, maxAgeDays int, compress bool, localTime bool, batchSize int, closeSignalSize int) (timur *TimurLogger, _err_ error) {
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
	fileWriter := log.New(logger, "", log.LstdFlags|log.Lmicroseconds)
	q := gqueue.New()
	if q == nil {
		return nil, fmt.Errorf("queue is nil")
	}
	if batchSize < defaultBatchSize {
		batchSize = defaultBatchSize
	}
	if closeSignalSize < defaultCloseSignalSize {
		closeSignalSize = defaultCloseSignalSize
	}
	result := &TimurLogger{
		batchSize:       batchSize,
		closeSignalSize: closeSignalSize,
		logger:          logger,
		fileWriter:      fileWriter,
		q:               q,
		isClosed:        atomic.Bool{},
		signals:         make(chan Signal, batchSize),
		closeSignal:     make(chan Signal, closeSignalSize),
		guard:           sync.WaitGroup{},
	}
	result.isClosed.Store(false)
	result.guard.Add(1)
	//创建后台异步写入协程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("logger background goroutine recover from panic: %v", r)
				fileWriter.Printf("logger background goroutine recover from panic: %v", r)
			}
			result.guard.Done()
			/*
				主协程close先将原子标识设为true，并向closeSignal通道中添加一个新信号，前者防止其他协程添加信号，后者通知后台协程结束
				通知主协程后台协程结束，只有在后台协程结束后，主协程中close函数才能开始关闭通道和清理资源
			*/
		}()
		for {
			ok, err := writeAll(result)
			if err != nil {
				fmt.Println(err)
			}
			if !ok {
				break
			}
		}
	}()
	return result, nil
}
