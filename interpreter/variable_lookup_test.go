package interpreter

import "testing"

func BenchmarkSafeValueEqual_Int(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		safeValueEqual(42, 42)
	}
}

func BenchmarkSafeValueEqual_String(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		safeValueEqual("hello", "hello")
	}
}
