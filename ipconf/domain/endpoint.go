package domain

import (
	"sync/atomic"
	"unsafe"
)

type Endpoint struct {
	IP          string       `json:"ip"`
	Port        int          `json:"port"`
	ActiveScore float64      `json:"active_score"`
	StaticScore float64      `json:"static_score"`
	Stats       *Stat        `json:"-"`
	window      *stateWindow `json:"-"`
}

func NewEndPoint(ip string, port int) *Endpoint {
	ed := &Endpoint{
		IP:   ip,
		Port: port,
	}
	ed.window = newStateWindow()
	ed.Stats = ed.window.getStat()
	go func() {
		for stat := range ed.window.statChan {
			ed.window.appendStat(stat)
			newStat := ed.window.getStat()
			atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(ed.Stats)), unsafe.Pointer(newStat)) // 这里使用了原子操作来更新Stats指针
			// atomic.SwapPointer((*unsafe.Pointer)(&ed.Stats), unsafe.Pointer(newStat))
		}
	}()
	return ed
}

// UpdateStat 更新Endpoint的统计信息
func (ed *Endpoint) UpdateStat(stat *Stat) {
	ed.window.statChan <- stat
}

// CalScore 计算Endpoint的得分
func (ed *Endpoint) CalScore(ctx *IPConfContext) {
	if ed.Stats != nil {
		ed.ActiveScore = ed.Stats.CalculateActiveSorce()
		ed.StaticScore = ed.Stats.CalculateStaticSorce()
	}
}
