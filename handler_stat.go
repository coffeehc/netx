// stat_procotol
package coffeenet

import (
	"sync/atomic"
	"time"
)

//统计接口
type StatInfo interface {
	GetHandlerStat() HandlerStat
	GetWorkRoutine() int
}

type HandlerStat struct {
	HandlerCount_avg int64
	HandlerCount     *int64
	ProcessTime_Max  time.Duration
	ProcessTime_Min  time.Duration
	queue            chan time.Duration
}

func NewHandlerStat() *HandlerStat {
	var handlerCount = int64(0)
	return &HandlerStat{0, &handlerCount, 0, time.Hour, make(chan time.Duration, 10000)}
}
func (this *HandlerStat) StartHandlerStat() {
	go func() {
		for {
			select {
			case delay := <-this.queue:
				if this.ProcessTime_Max < delay {
					this.ProcessTime_Max = delay
				}
				if this.ProcessTime_Min > delay {
					this.ProcessTime_Min = delay
				}
			}
		}
	}()
	go func() {
		count := atomic.LoadInt64(this.HandlerCount)
		timer := time.NewTimer(0)
		for {
			timer.Reset(time.Second)
			select {
			case <-timer.C:
				newCount := atomic.LoadInt64(this.HandlerCount)
				this.HandlerCount_avg = newCount - count
				count = newCount
			}
		}
		timer.Stop()
	}()
}

func (this *HandlerStat) acceptData(size time.Duration) {
	atomic.AddInt64(this.HandlerCount, 1)
	this.queue <- size
}
