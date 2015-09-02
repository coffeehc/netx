// stat_procotol
package coffeenet

import "time"

type HanderStat struct {
	HandlerCount_avg int64
	HandlerCount     int64
	ProcessTime_Max  time.Duration
	ProcessTime_Min  time.Duration
	queue            chan time.Duration
}

func NewHanderStat() *HanderStat {
	return &HanderStat{0, 0, 0, 10000000, make(chan time.Duration)}
}
func (this *HanderStat) StartHanderStat() {
	go func() {
		for {
			select {
			case delay := <-this.queue:
				this.HandlerCount += 1
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
		count := this.HandlerCount
		for {
			select {
			case <-time.After(time.Second):
				this.HandlerCount_avg = this.HandlerCount - count
				count = this.HandlerCount
			}
		}
	}()
}

func (this *HanderStat) AcceptData(size time.Duration) {
	this.queue <- size
}
