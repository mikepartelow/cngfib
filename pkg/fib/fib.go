package fib

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	ServiceName    = "fibonnaci"
	ServiceVersion = "0.1.0"

	instrumentationName    = "mp/fib"
	instrumentationVersion = "0.1.0"
)

type Memoizer func(uint, uint)

// Memo returns a memoized result for a fibonacci number, a bool that is true if a result was returned, and a func to memoize a new result.
type Memo func(uint) (uint, bool, Memoizer)

// WithMemoization returns a Memo func that allows fib.Recurse to memoize results. Not safe for concurrency.
func WithSimpleMemoization() Memo {
	results := make(map[uint]uint)
	return func(i uint) (uint, bool, Memoizer) {
		v, ok := results[i]
		return v, ok, func(i, result uint) {
			results[i] = result
		}
	}
}

// Iterate uses iteration to compute Fibonacci numbers. Use this approach on leetcode and interviews.
func Iterate(ctx context.Context, num uint) uint {
	if num <= 1 {
		return num
	}
	var n_minus_two, n_minus_one, n uint
	n, n_minus_one = 1, 1

	for i := uint(2); i < num; i++ {
		select {
		case <-ctx.Done():
			return 0
		default:
			n_minus_two, n_minus_one = n_minus_one, n
			n = n_minus_two + n_minus_one
		}
	}

	return n
}

// Recuse uses recursion to compute Fibonnaci numbers. Pass in an optional Memo func for extra performance/fun.
func Recurse(ctx context.Context, num uint, memos ...Memo) (result uint) {
	tracer := otel.Tracer(instrumentationName,
		trace.WithInstrumentationVersion(instrumentationVersion),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	cctx, sp := tracer.Start(ctx,
		fmt.Sprintf("Recursive Fibonacci(%d)", num),
		trace.WithAttributes(attribute.Int("n", int(num))))
	defer sp.End()

	select {
	case <-cctx.Done():
		sp.SetAttributes(attribute.Bool("canceled", true))
		return 0
	default:
	}

	defer func() { sp.SetAttributes(attribute.Int("result", int(result))) }()

	for _, memo := range memos {
		r, ok, memoize := memo(num)
		if ok {
			sp.SetAttributes(attribute.Bool("from memo", true))
			result = r
			return
		}
		defer func() { memoize(num, result) }()
	}

	if num <= 1 {
		result = num
		return
	}
	result = Recurse(cctx, num-2, memos...) + Recurse(cctx, num-1, memos...)
	return
}

// Channel uses goroutines and channles to compute Fibonacci numbers.
func Channel(ctx context.Context, num uint) chan uint {
	ch := make(chan uint)

	go func() {
		tracer := otel.Tracer(instrumentationName,
			trace.WithInstrumentationVersion(instrumentationVersion),
			trace.WithSchemaURL(semconv.SchemaURL),
		)
		cctx, sp := tracer.Start(ctx,
			fmt.Sprintf("Channel Fibonacci(%d)", num),
			trace.WithAttributes(attribute.Int("n", int(num))))
		defer sp.End()

		select {
		case <-cctx.Done():
			sp.SetAttributes(attribute.Bool("canceled", true))
			close(ch)
			return
		default:
		}

		result := num

		if num > 1 {
			a := Channel(cctx, num-1)
			b := Channel(cctx, num-2)

			result = <-a + <-b
		}

		sp.SetAttributes(attribute.Int("result", int(result)))

		ch <- result
	}()

	return ch
}
