package monitor

import (
	"github.com/coffeehc/coffeenet"
	"github.com/coffeehc/web"
	"net/http"
)

type Monitor struct {
	statInfo coffeenet.StatInfo
}

func NewNetServerMonitor(statInfo coffeenet.StatInfo) *Monitor {
	return &Monitor{statInfo}
}

func (this *Monitor) ShowStatInfo(r *http.Request, pathValues map[string]string, reply web.Reply) {
	result := new(struct {
		HandlerStat   coffeenet.HandlerStat
		WorkGoruntine int
	})
	result.HandlerStat = this.statInfo.GetHandlerStat()
	result.WorkGoruntine = this.statInfo.GetWorkRoutine()
	reply.With(result).As(web.Transport_Json)
}
