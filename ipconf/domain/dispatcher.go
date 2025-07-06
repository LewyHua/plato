package domain

import (
	"sort"
	"sync"
)

type Dispatcher struct {
	candidateTable map[string]*Endpoint
	sync.RWMutex
}

var dp *Dispatcher

func Init() {
	dp := &Dispatcher{
		candidateTable: make(map[string]*Endpoint),
	}
	go func() {
		for event := range source.EventChan() {
			switch event.Type {
			case source.AddNodeEvent:
				dp.addNode(event)
			case source.DelNodeEvent:
				dp.delNode(event)
			}
		}
	}()
}

// Dispatch 返回按照得分排序的候选节点列表
func Dispatch(ctx *IPConfContext) []*Endpoint {
	// 1. 获取候选节点列表
	eds := dp.getCandidateEndpoints(ctx)
	// 2. 对候选节点进行算分
	for _, ed := range eds {
		ed.CalScore(ctx)
	}
	// 3. 按照得分进行排序: 先按照ActiceScore降序排序，如果相同则按照StaticScore降序排序
	sort.Slice(eds, func(i, j int) bool {
		if eds[i].ActiveScore == eds[j].ActiveScore {
			return eds[i].StaticScore > eds[j].StaticScore
		}
		return eds[i].ActiveScore > eds[j].ActiveScore
	})
	return eds
}

func (d *Dispatcher) getCandidateEndpoints(ctx *IPConfContext) []*Endpoint {
	d.RLock()
	defer d.RUnlock()

	candidateList := make([]*Endpoint, 0, len(dp.candidateTable))
	for _, ed := range dp.candidateTable {
		candidateList = append(candidateList, ed)
	}
	return candidateList
}

func (d *Dispatcher) addNode(event *source.Event) {
	d.Lock()
	defer d.Unlock()

	ed := NewEndPoint(event.IP, event.Port)
	ed.UpdateStat(&Stat{
		ConnectNum:   event.ConnectNum,
		MessageBytes: event.MessageBytes,
	})

	d.candidateTable[event.Endpoint.IP] = event.Endpoint
}

func (dp *Dispatcher) delNode(event *source.Event) {
	dp.Lock()
	defer dp.Unlock()
	delete(dp.candidateTable, event.Key())
}
