package submit_unit

import (
	"fmt"
	"sync/atomic"

	"golang.org/x/sync/semaphore"
)

var defaultConcurrentSize int = 10

type TaskFunc func(any) (any, error)

type ResultType struct {
	Result any
	Error  error
}

type SubmitUnit struct {
	maxConcurrentSize int
	semaphore         *semaphore.Weighted
	isClosed          atomic.Bool
}

func (su *SubmitUnit) TrySubmit(task TaskFunc, param any) (future chan *ResultType, _err_ error) {
	//_err_ 不为空，表示在关闭后尝试提交任务或者信号量获取失败
	//检查是否关闭
	if su.isClosed.Load() {
		return nil, fmt.Errorf("submit unit is closed")
	}
	//进行信号量wait操作
	if !su.semaphore.TryAcquire(1) {
		return nil, fmt.Errorf("submit unit is full")
	} else {
		//创建future通道
		future = make(chan *ResultType, 1)
		//创建一个新协程执行任务，并返回通道
		go func() {
			defer func() {
				if r := recover(); r != nil {
					future <- &ResultType{
						nil,
						fmt.Errorf("recover from panic: %v", r),
					}
				}
			}()
			result, err := task(param)
			future <- &ResultType{
				result,
				err,
			}
		}()
		return future, nil
	}
}

// func (su *SubmitUnit) Close() (_err_ error) {
// defer func() {
// 	if r := recover(); r != nil {
// 		_err_ = fmt.Errorf("recover from panic: %v", r)
// 	}
// }()
//先将关闭标志设为true，防止生产者再添加任务

//等待后台任务全部结束（或者使用context向任务发送取消信号？暂时还不确定）

//清理全部资源
// }

func NewSubmitUnit(maxConcurrentSize int) (su *SubmitUnit, _err_ error) {
	defer func() {
		if r := recover(); r != nil {
			_err_ = fmt.Errorf("recover from panic: %v", r)
			su = nil
		}
	}()
	if maxConcurrentSize <= 0 {
		maxConcurrentSize = defaultConcurrentSize
	}
	semaphore := semaphore.NewWeighted(int64(maxConcurrentSize))
	result := &SubmitUnit{
		maxConcurrentSize: maxConcurrentSize,
		semaphore:         semaphore,
		isClosed:          atomic.Bool{},
	}
	result.isClosed.Store(false)
	return result, nil
}
