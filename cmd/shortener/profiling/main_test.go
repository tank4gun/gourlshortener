package main

import "testing"

func BenchmarkFibonacci(b *testing.B) {
	b.Run("Fibonacci", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Fibonacci(20)
		}
	})
}

//
//func BenchmarkLoadBalancer(b *testing.B) {
//	conns := make([]*Connection, 100)
//	lbChan := NewLoadBalancerChan(conns)
//	lbChan.Init()
//	lbAtomic := NewLoadBalancerAtomic(conns)
//	lbMutex := NewLoadBalancerMutex(conns)
//	b.ResetTimer()
//
//	b.Run("LoadBalancerChan", func(b *testing.B) {
//		for i := 0; i < 1000; i++ {
//			lbChan.NextConn()
//		}
//	})
//	b.Run("LoadBalancerAtomic", func(b *testing.B) {
//		for i := 0; i < 1000; i++ {
//			lbAtomic.NextConn()
//		}
//	})
//	b.Run("LoadBalancerMutex", func(b *testing.B) {
//		for i := 0; i < 1000; i++ {
//			lbMutex.NextConn()
//		}
//	})
//}
