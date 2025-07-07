package domain

import (
	"sort"
	"sync"

	"github.com/lewyhua/plato/ipconf/source"
)

type Dispatcher struct {
	candidateTable map[string]*Endpoint
	sync.RWMutex
}

var dp *Dispatcher

func Init() {
	dp = &Dispatcher{}
	dp.candidateTable = make(map[string]*Endpoint)
	go func() {
		for event := range source.EventChan() {
			switch event.Type {
			case source.AddNodeEventType:
				dp.addNode(event)
			case source.DelNodeEventType:
				dp.delNode(event)
			}
		}
	}()
}
func Dispatch(ctx *IPConfContext) []*Endpoint {
	// Step1: 获得候选endport
	eds := dp.getCandidateEndport(ctx)
	// Step2: 逐一计算得分
	for _, ed := range eds {
		ed.CalScore(ctx)
	}
	// Step3: 全局排序，动静结合的排序策略。
	sort.Slice(eds, func(i, j int) bool {
		// 优先基于活跃分数进行排序
		if eds[i].ActiveScore > eds[j].ActiveScore {
			return true
		}
		// 如果活跃分数相同，则使用静态分数排序
		if eds[i].ActiveScore == eds[j].ActiveScore {
			return eds[i].StaticScore > eds[j].StaticScore
		}
		return false
	})
	return eds
}

func (dp *Dispatcher) getCandidateEndport(ctx *IPConfContext) []*Endpoint {
	dp.RLock()
	defer dp.RUnlock()
	candidateList := make([]*Endpoint, 0, len(dp.candidateTable))
	for _, ed := range dp.candidateTable {
		candidateList = append(candidateList, ed)
	}
	return candidateList
}
func (dp *Dispatcher) delNode(event *source.Event) {
	dp.Lock()
	defer dp.Unlock()
	delete(dp.candidateTable, event.Key())
}
func (dp *Dispatcher) addNode(event *source.Event) {
	dp.Lock()
	defer dp.Unlock()
	ed := NewEndPoint(event.IP, event.Port)
	ed.UpdateStat(&Stat{
		ConnectNum:   event.ConnectNum,
		MessageBytes: event.MessageBytes,
	})
	dp.candidateTable[event.Key()] = ed
}
