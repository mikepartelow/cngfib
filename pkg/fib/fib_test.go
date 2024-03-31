package fib_test

import (
	"context"
	"mp/fib/pkg/fib"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var want []int

func init() {
	// uncomment more of these to make your tests run for longer
	want = []int{0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 377, 610, 987, 1597, 2584, 4181, 6765} //, 10946, 17711, 28657, 46368, 75025, 121393, 196418} //, 317811, 514229, 832040, 1346269, 2178309, 3524578, 5702887, 9227465, 14930352, 24157817, 39088169, 63245986, 102334155, 165580141, 267914296, 433494437, 701408733, 1134903170, 1836311903, 2971215073, 4807526976, 7778742049, 12586269025, 20365011074, 32951280099, 53316291173, 86267571272, 139583862445, 225851433717, 365435296162, 591286729879, 956722026041, 1548008755920, 2504730781961}
}

func TestFibonacci(t *testing.T) {
	for i := 0; i < len(want); i++ {
		got := fib.Iterate(context.Background(), i)
		require.Equal(t, want[i], got)
	}

	for i := 0; i < len(want); i++ {
		got := fib.Recurse(context.Background(), i)
		require.Equal(t, want[i], got)
	}

	for i := 0; i < len(want); i++ {
		got := fib.Recurse(context.Background(), i, fib.WithSimpleMemoization())
		require.Equal(t, want[i], got)
	}

	for i := 0; i < len(want); i++ {
		got := <-fib.Channel(context.Background(), i)
		require.Equal(t, want[i], got)
	}
}

func TestCancel(t *testing.T) {
	cctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	fib.Iterate(cctx, 10000000) // Iterate is fast enough that we must set num very high to test cancel
	assert.NotNil(t, cctx.Err())

	cctx, cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	fib.Recurse(cctx, 1000)
	assert.NotNil(t, cctx.Err())

	cctx, cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	fib.Recurse(cctx, 10000, fib.WithSimpleMemoization())
	assert.NotNil(t, cctx.Err())

	cctx, cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	<-fib.Channel(cctx, 1000)
	assert.NotNil(t, cctx.Err())
}

func BenchmarkIterate(b *testing.B) {
	for j := 0; j < b.N; j++ {
		for i := 0; i < len(want); i++ {
			got := fib.Iterate(context.Background(), i)
			require.Equal(b, want[i], got)
		}
	}
}

func BenchmarkRecurse(b *testing.B) {
	for j := 0; j < b.N; j++ {
		for i := 0; i < len(want); i++ {
			got := fib.Recurse(context.Background(), i)
			require.Equal(b, want[i], got)
		}
	}
}

func BenchmarkRecurseWithMemo(b *testing.B) {
	for j := 0; j < b.N; j++ {
		for i := 0; i < len(want); i++ {
			got := fib.Recurse(context.Background(), i, fib.WithSimpleMemoization())
			require.Equal(b, want[i], got)
		}
	}
}

func BenchmarkChannel(b *testing.B) {
	for j := 0; j < b.N; j++ {
		for i := 0; i < len(want); i++ {
			got := <-fib.Channel(context.Background(), i)
			require.Equal(b, want[i], got)
		}
	}
}
