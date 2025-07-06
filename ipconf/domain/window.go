package domain

const (
	windowSize = 5
)

type stateWindow struct {
	stateQueue []*Stat
	statChan   chan *Stat
	sumStat    *Stat
	idx        int
}

func newStateWindow() *stateWindow {
	return &stateWindow{
		stateQueue: make([]*Stat, windowSize),
		statChan:   make(chan *Stat),
		sumStat:    &Stat{},
		idx:        0,
	}
}

func (sw *stateWindow) getStat() *Stat {
	res := sw.sumStat.Clone() // 克隆是因为sumStat是一个共享的状态，需要避免并发修改导致数据不一致
	res.Avg(windowSize)
	return res
}

// appendStat 将新的Stat添加到这个Endpoint的状态窗口中，并更新统计信息。
func (sw *stateWindow) appendStat(s *Stat) {
	// 减去即将被删除的state
	sw.sumStat.Sub(sw.stateQueue[sw.idx%windowSize])
	// 更新最新的stat
	sw.stateQueue[sw.idx%windowSize] = s
	// 计算最新的窗口和
	sw.sumStat.Add(s)
	sw.idx++
}
