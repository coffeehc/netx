// stat_procotol
package coffeenet

import (
	"time"
	"sync/atomic"
)

//统计接口
type StatInfo interface {
	GetHanderStat() HanderStat
	GetWorkRuntine() int
}

type HanderStat struct {
	HandlerCount_avg int64
	HandlerCount     *int64
	ProcessTime_Max  time.Duration
	ProcessTime_Min  time.Duration
	queue            chan time.Duration
}

func NewHanderStat() *HanderStat {
	var handlerCount = int64(0)
	return &HanderStat{0,&handlerCount, 0, time.Hour, make(chan time.Duration, 10000)}
}
func (this *HanderStat) StartHanderStat() {
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
		for {
			select {
			case <-time.After(time.Second):
				newCount := atomic.LoadInt64(this.HandlerCount)
				this.HandlerCount_avg = newCount - count
				count = newCount
			}
		}
	}()
}

func (this *HanderStat) acceptData(size time.Duration) {
	atomic.AddInt64(this.HandlerCount,1)
	this.queue <- size
}
