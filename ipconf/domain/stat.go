package domain

import "math"

// 对于gateway网关机来说，存在不同时期加入进来的物理机，所以机器的配置是不同的，使用负载来衡量会导致偏差。
// 为更好的应对动态的机器配置变化，我们统计其剩余资源值，来衡量一个机器其是否更适合增加其负载。

// Stat 代表一个机器的剩余资源统计信息，每一秒钟统计一次
type Stat struct {
	ConnectNum   float64 // 业务上，im gateway 总体持有的长连接数量 的剩余值
	MessageBytes float64 // 业务上，im gateway 每秒收发消息的总字节数 的剩余值
}

// 因此假设网络带宽将是系统瓶颈所在，那么哪台机器富余的带宽资源多，哪台机器的负载就是最轻的。
// TODO: 如何预估他的数量级？何时使用静态值衡量
// TODO: json 的解析失效
func (s *Stat) CalculateActiveSorce() float64 {
	return getGB(s.MessageBytes)
}

func (s *Stat) Avg(num float64) {
	s.ConnectNum /= num
	s.MessageBytes /= num
}
func (s *Stat) Clone() *Stat {
	newStat := &Stat{
		MessageBytes: s.MessageBytes,
		ConnectNum:   s.ConnectNum,
	}
	return newStat
}

func (s *Stat) Add(st *Stat) {
	if st == nil {
		return
	}
	s.ConnectNum += st.ConnectNum
	s.MessageBytes += st.MessageBytes
}

func (s *Stat) Sub(st *Stat) {
	if st == nil {
		return
	}
	s.ConnectNum -= st.ConnectNum
	s.MessageBytes -= st.MessageBytes
}

// getGB 将字节数转换为GB
func getGB(m float64) float64 {
	return decimal(m / (1 << 30))
}

// decimal 保留两位小数
func decimal(value float64) float64 {
	return math.Trunc(value*1e2+0.5) * 1e-2 // 这里使用了四舍五入的方式保留两位小数
}

// min 返回三个数中的最小值
func min(a, b, c float64) float64 {
	m := func(k, j float64) float64 {
		if k > j {
			return j
		}
		return k
	}
	return m(a, m(b, c))
}
func (s *Stat) CalculateStaticSorce() float64 {
	return s.ConnectNum
}
