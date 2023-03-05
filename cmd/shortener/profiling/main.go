//package main
//
//import (
//	"bytes"
//	"net/http"
//	"strings"
//)

//	func Fibonacci(n int) int {
//		if n <= 1 {
//			return 1
//		}
//		return Fibonacci(n-1) + Fibonacci(n-2)
//	}
//
//	func main() {
//		fmt.Println(Fibonacci(7))
//	}
//
// package main
//
// import (
//
//	"sync"
//	"sync/atomic"
//
// )
//
// type Connection struct{}
//
//	type LoadBalancer interface {
//		NextConn() *Connection
//	}
//
//	type LoadBalancerChan struct {
//		conns []*Connection
//		ch    chan *Connection
//		stop  chan struct{}
//	}
//
//	func NewLoadBalancerChan(conns []*Connection) *LoadBalancerChan {
//		return &LoadBalancerChan{conns: conns, ch: make(chan *Connection), stop: make(chan struct{})}
//	}
//
//	func (b *LoadBalancerChan) Init() {
//		go b.worker()
//	}
//
//	func (b *LoadBalancerChan) Close() {
//		b.stop <- struct{}{}
//	}
//
//	func (b *LoadBalancerChan) worker() {
//		for i := 0; ; {
//			select {
//			case b.ch <- b.conns[i]:
//				i++
//				if i == len(b.conns) {
//					i = 0
//				}
//
//			case <-b.stop:
//				return
//			}
//		}
//	}
//
//	func (b *LoadBalancerChan) NextConn() *Connection {
//		return <-b.ch
//	}
//
//	type LoadBalancerAtomic struct {
//		conns   []*Connection
//		counter uint32
//	}
//
//	func NewLoadBalancerAtomic(conns []*Connection) *LoadBalancerAtomic {
//		return &LoadBalancerAtomic{conns: conns}
//	}
//
//	func (b *LoadBalancerAtomic) NextConn() *Connection {
//		i := atomic.AddUint32(&b.counter, 1) % uint32(len(b.conns))
//		return b.conns[i]
//	}
//
//	type LoadBalancerMutex struct {
//		conns   []*Connection
//		counter int
//		mu      sync.Mutex
//	}
//
//	func NewLoadBalancerMutex(conns []*Connection) *LoadBalancerMutex {
//		return &LoadBalancerMutex{conns: conns}
//	}
//
//	func (b *LoadBalancerMutex) NextConn() *Connection {
//		b.mu.Lock()
//		defer b.mu.Unlock()
//		b.counter = (b.counter + 1) % len(b.conns)
//		return b.conns[b.counter]
//	}
package main

import (
	"bytes"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
	"strings"
)

const (
	addr    = ":8080"  // адрес сервера
	maxSize = 10000000 // будем растить слайс до 10 миллионов элементов
)

func AlignRightSimple(s string, length int, lead rune) string {
	for len(s) < length {
		s = string(lead) + s
	}
	return s
}

func AlignRightBuffer(s string, length int, lead rune) string {
	buf := bytes.Buffer{}
	for i := 0; i < length-len(s); i++ {
		buf.WriteRune(lead)
	}
	buf.WriteString(s)
	return buf.String()
}

func AlignRightRepeat(s string, length int, lead rune) string {
	if len(s) < length {
		return strings.Repeat(string(lead), length-len(s)) + s
	}
	return s
}

func foo() {
	// полезная нагрузка
	s := "123777"
	for {
		//var s []int
		//for i := 0; i < maxSize; i++ {
		//	s = append(s, i)
		//}
		AlignRightSimple(s, 1000, 'a')

		AlignRightBuffer(s, 1000, 'a')

		AlignRightRepeat(s, 1000, 'a')
	}
}

func main() {
	go foo()                       // запускаем полезную нагрузку в фоне
	http.ListenAndServe(addr, nil) // запускаем сервер
}
