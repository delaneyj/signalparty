package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/delaneyj/signalparty/alien"
	"github.com/delaneyj/signalparty/foo"
	"github.com/delaneyj/signalparty/rocket"
	"github.com/jamiealquiza/tachymeter"
	"github.com/jedib0t/go-pretty/v6/table"
)

func main() {
	flag.Parse()

	f, err := os.Create("default.pgo")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	log.Printf("warming up")

	benchmarkAlien(true)
	// for i := 0; i < 10; i++ {
	benchmarkRocket(true)
	benchmarkFoo(true)
	// }
	// benchmarkDumbdumb(true)
}

var (
	ww    = []int{1, 10, 100, 1_000}
	hh    = []int{1, 10, 100, 1_000}
	iters = 100
)

func addOne(oldValue int) int {
	return oldValue + 1
}
func pass(l int) error {
	return nil
}

func benchmarkAlien(shouldRender bool) {

	getValue := func(x any) int {
		switch x := x.(type) {
		case *alien.WriteableSignal[int]:
			return x.Value() + 1
		case *alien.ReadonlySignal[int]:
			return x.Value() + 1
		default:
			panic("unknown type")
		}
	}

	tbl := table.NewWriter()
	tbl.SetTitle("Alien Signals")
	tbl.SetOutputMirror(os.Stdout)
	tbl.AppendHeader(table.Row{"benchmark", "avg", "min", "p75", "p99", "max"})

	for _, w := range ww {
		for _, h := range hh {

			tach := tachymeter.New(&tachymeter.Config{Size: iters})

			// fmt.Sprintf("propagate: %dx%d", w, h), func(b *testing.B) {
			rs := alien.CreateReactiveSystem(func(from alien.SignalAware, err error) {
				log.Panic(err)
			})
			src := alien.Signal(rs, 1)
			for i := 0; i < w; i++ {
				var last any
				last = src
				for j := 0; j < h; j++ {
					prev := last
					last = alien.Computed(rs, func(oldValue int) int {
						return getValue(prev)
					})
				}

				alien.Effect(rs, func() error {
					getValue(last)
					return nil
				})

			}

			for i := 0; i < iters; i++ {
				start := time.Now()
				src.SetValue(src.Value() + 1)
				tach.AddTime(time.Since(start))
			}

			calc := tach.Calc()
			tbl.AppendRows([]table.Row{
				{
					fmt.Sprintf("propagate: %d * %d", w, h),
					calc.Time.Avg,
					calc.Time.Min,
					calc.Time.P75,
					calc.Time.P99,
					calc.Time.Max,
				},
			})
		}
	}

	if shouldRender {
		tbl.Render()
	}

}

func benchmarkFoo(shouldRender bool) {

	tbl := table.NewWriter()
	tbl.SetTitle("Foo Signals")
	tbl.SetOutputMirror(os.Stdout)
	tbl.AppendHeader(table.Row{"benchmark", "avg", "min", "p75", "p99", "max"})

	pass := func(l int) {
	}

	for _, w := range ww {
		for _, h := range hh {
			tach := tachymeter.New(&tachymeter.Config{Size: iters})

			// fmt.Sprintf("propagate: %dx%d", w, h), func(b *testing.B) {
			src := foo.Signal(1)
			for i := 0; i < w; i++ {
				var last foo.Dependency
				last = src
				for j := 0; j < h; j++ {
					prev := last
					last = foo.Computed1(prev, addOne)
				}

				foo.Effect1(last, pass)
			}

			for i := 0; i < iters; i++ {
				start := time.Now()
				src.SetValue(src.Value() + 1)
				tach.AddTime(time.Since(start))
			}

			calc := tach.Calc()
			tbl.AppendRows([]table.Row{
				{
					fmt.Sprintf("propagate: %d * %d", w, h),
					calc.Time.Avg,
					calc.Time.Min,
					calc.Time.P75,
					calc.Time.P99,
					calc.Time.Max,
				},
			})
		}
	}

	if shouldRender {
		tbl.Render()
	}
}

func benchmarkRocket(shouldRender bool) {

	tbl := table.NewWriter()
	tbl.SetTitle("🚀 Signals")
	tbl.SetOutputMirror(os.Stdout)
	tbl.AppendHeader(table.Row{"benchmark", "avg", "min", "p75", "p99", "max"})

	rs := rocket.NewReactiveSystem()
	for _, w := range ww {
		for _, h := range hh {
			tach := tachymeter.New(&tachymeter.Config{Size: iters})

			// fmt.Sprintf("propagate: %dx%d", w, h), func(b *testing.B) {

			src := rocket.Signal(rs, 1)
			for i := 0; i < w; i++ {
				var last rocket.Dependency
				last = src
				for j := 0; j < h; j++ {
					prev := last
					last = rocket.Computed1(rs, prev, addOne)
				}

				rocket.Effect1(rs, last, pass)
			}

			for i := 0; i < iters; i++ {
				start := time.Now()
				src.SetValue(src.Value() + 1)
				tach.AddTime(time.Since(start))
			}

			calc := tach.Calc()
			tbl.AppendRows([]table.Row{
				{
					fmt.Sprintf("propagate: %d * %d", w, h),
					calc.Time.Avg,
					calc.Time.Min,
					calc.Time.P75,
					calc.Time.P99,
					calc.Time.Max,
				},
			})
		}
	}

	if shouldRender {
		tbl.Render()
	}
}
