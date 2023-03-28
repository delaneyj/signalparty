package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/delaneyj/signalparty/pkg/reactively"
	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
)

func main() {
	log.Print("Starting turnsignal benchmark, please wait...")
	defer log.Print("Finished turnsignal benchmark")

	perfTestCfgs := []benchmarkTestConfig{
		{
			name:           "simple component",
			width:          10, // can't change for decorator tests
			staticFraction: 1,  // can't change for decorator tests
			nSources:       2,  // can't change for decorator tests
			totalLayers:    5,
			readFraction:   0.2,
			iterations:     600000,
			expectedSum:    19199968,
			expectedCount:  3480000,
		},
		{
			name:           "dynamic component",
			width:          10,
			totalLayers:    10,
			staticFraction: 0.75,
			nSources:       6,
			readFraction:   0.2,
			iterations:     15000,
			expectedSum:    302310782860,
			expectedCount:  1155000,
		},
		{
			name:           "large web app",
			width:          1000,
			totalLayers:    12,
			staticFraction: 0.95,
			nSources:       4,
			readFraction:   1,
			iterations:     7000,
			expectedSum:    29355933696000,
			expectedCount:  1463000,
		},
		{
			name:           "wide dense",
			width:          1000,
			totalLayers:    5,
			staticFraction: 1,
			nSources:       25,
			readFraction:   1,
			iterations:     3000,
			expectedSum:    1171484375000,
			expectedCount:  732000,
		},
		{
			name:           "deep",
			width:          5,
			totalLayers:    500,
			staticFraction: 1,
			nSources:       3,
			readFraction:   1,
			iterations:     500,
			expectedSum:    3.0239642676898464e241,
			expectedCount:  1246500,
		},
		{
			name:           "very dynamic",
			width:          100,
			totalLayers:    15,
			staticFraction: 0.5,
			nSources:       6,
			readFraction:   1,
			iterations:     2000,
			expectedSum:    15664996402790400,
			expectedCount:  1078000,
		},
	}

	type results struct {
		sum       int
		count     int64
		duration  time.Duration
		isDynamic [][]bool
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"framework", "size", "nSources", "read%", "static%",
		"nTimes", "test", "time",
		// "gcTime",
		"updateRate", "title",
	})

	testRepeats := 5
	for _, cfg := range perfTestCfgs {
		log.Printf("Running '%s' config", cfg.name)
		counter := new(int64)
		graph, isDynamic := benchmarkMakeGraph(&benchmarkMakeGraphConfig{
			counter:        counter,
			width:          cfg.width,
			totalLayers:    cfg.totalLayers,
			nSources:       cfg.nSources,
			staticFraction: cfg.staticFraction,
		})

		runOnce := func() int {
			return benchmarkRunGraph(&benchmarkRunGraphConfig{
				rctx:         &reactively.ReactiveContext{},
				graph:        graph,
				iteration:    cfg.iterations,
				readFraction: cfg.readFraction,
			})
		}
		// run once to warm up
		runOnce()

		bestResult := &results{
			duration: time.Hour,
		}

		for i := 0; i < testRepeats; i++ {
			log.Printf("Running '%s' config, iteration %d/%d %d%%", cfg.name, i+1, testRepeats, (i+1)*100/testRepeats)
			*counter = 0
			start := time.Now()
			sum := runOnce()
			duration := time.Since(start)
			// log.Printf("%s run: %d sum: %d, count: %d, duration: %s", cfg.name, i, sum, *counter, duration)

			if duration < bestResult.duration {
				bestResult.duration = duration
				bestResult.sum = sum
				bestResult.count = *counter
				bestResult.isDynamic = isDynamic
			}
		}

		makeTitle := func() string {
			sb := strings.Builder{}
			sb.WriteString(fmt.Sprintf("%dx%d %d sources", cfg.width, cfg.totalLayers, cfg.nSources))
			if cfg.staticFraction < 1 {
				sb.WriteString(" dynamic")
			}
			if cfg.readFraction < 1 {
				sb.WriteString(fmt.Sprintf(" read %0.2f%%", 100*cfg.readFraction))
			}
			return sb.String()
		}

		updateRate := float64(bestResult.count) / (float64(bestResult.duration) / float64(time.Millisecond))

		table.Append([]string{
			"reactively", // framework
			fmt.Sprintf("%dx%d", cfg.width, cfg.totalLayers), // size
			fmt.Sprint(cfg.nSources),                         // nSources
			fmt.Sprint(cfg.readFraction),                     // read%
			fmt.Sprint(cfg.staticFraction),                   // static%
			humanize.Comma(cfg.iterations),                   // nTimes
			cfg.name,                                         // test
			fmt.Sprint(bestResult.duration),                  // time
			// fmt.Sprint(999),                                  // gcTime
			humanize.Comma(int64(updateRate)), // updateRate
			makeTitle(),                       // title

		})
	}
	table.Render() // Send output
}

type benchmarkTestConfig struct {
	name           string  // friendly name for the test, should be unique
	width          int64   // width of dependency graph to construct
	totalLayers    int64   // depth of dependency graph to construct
	staticFraction float64 // fraction of nodes that are static */ // TODO change to dynamicFraction
	nSources       int64   // construct a graph with number of sources in each node
	readFraction   float64 // fraction of [0, 1] elements in the last layer from which to read values in each test iteration
	iterations     int64   // number of test iterations
	expectedSum    float64 // sum of all iterations, for verification
	expectedCount  int64   //  count of all iterations, for verification
}

type benchmarkGraph struct {
	sources []*reactively.Reactive[int]
	layers  [][]*reactively.Reactive[int]
}

type benchmarkMakeGraphConfig struct {
	counter                      *int64
	width, totalLayers, nSources int64
	staticFraction               float64
}

func benchmarkMakeGraph(cfg *benchmarkMakeGraphConfig) (graph *benchmarkGraph, isDynamic [][]bool) {
	rctx := &reactively.ReactiveContext{}
	sources := make([]*reactively.Reactive[int], cfg.width)
	for i := range sources {
		sources[i] = reactively.Signal(rctx, i)
	}
	graph = &benchmarkGraph{sources: sources}
	graph.layers, isDynamic = makeBenchmarkDependentRows(&benchmarkMakeDependentRowsConfig{
		rctx:           rctx,
		sources:        sources,
		numRows:        cfg.totalLayers - 1,
		counter:        cfg.counter,
		staticFraction: cfg.staticFraction,
		nSources:       cfg.nSources,
	})

	return

}

type benchmarkRunGraphConfig struct {
	rctx         *reactively.ReactiveContext
	graph        *benchmarkGraph
	iteration    int64
	readFraction float64
}

// Execute the graph by writing one of the sources and reading some or all of the leaves.
// return the sum of all leaf values
func benchmarkRunGraph(cfg *benchmarkRunGraphConfig) int {
	random := rand.New(rand.NewSource(0))
	leaves := cfg.graph.layers[len(cfg.graph.layers)-1]
	skipCount := int(math.Round(float64(len(leaves)) * (1 - cfg.readFraction)))
	readLeaves := benchmarkRemoveElems(leaves, skipCount, random)

	for i := 0; i < int(cfg.iteration); i++ {
		// writing signals
		reactively.Memo(cfg.rctx, func() bool {
			sourceDex := i % len(cfg.graph.sources)
			cfg.graph.sources[sourceDex].Write(i + sourceDex)

			return true
		})

		// reading nth leaves
		for _, leaf := range readLeaves {
			leaf.Read()
		}
	}

	sum := 0
	for _, leaf := range readLeaves {
		sum += leaf.Read()
	}
	return sum
}

func benchmarkRemoveElems[T comparable](src []T, rmCount int, rand *rand.Rand) []T {
	copyWithRemovals := make([]T, len(src))
	copy(copyWithRemovals, src)
	for i := 0; i < rmCount; i++ {
		rmDex := rand.Intn(len(copyWithRemovals))
		copyWithRemovals[rmDex] = copyWithRemovals[len(copyWithRemovals)-1]
		copyWithRemovals = copyWithRemovals[:len(copyWithRemovals)-1]
	}
	return copyWithRemovals
}

type benchmarkMakeDependentRowsConfig struct {
	rctx              *reactively.ReactiveContext
	sources           []*reactively.Reactive[int]
	numRows, nSources int64
	counter           *int64
	staticFraction    float64
}

func makeBenchmarkDependentRows(cfg *benchmarkMakeDependentRowsConfig) (row [][]*reactively.Reactive[int], isDynamic [][]bool) {
	prevRow := make([]*reactively.Reactive[int], len(cfg.sources))
	copy(prevRow, cfg.sources)

	random := rand.New(rand.NewSource(0))
	rows := make([][]*reactively.Reactive[int], cfg.numRows)
	allDynamic := make([][]bool, cfg.numRows)
	for l := int64(0); l < cfg.numRows; l++ {
		row, isDynamic := makeBenchmarkRow(&benchmarkRowConfig{
			rctx:           cfg.rctx,
			sources:        prevRow,
			counter:        cfg.counter,
			staticFraction: cfg.staticFraction,
			nSources:       cfg.nSources,
			rand:           random,
		})
		rows[l] = row
		allDynamic[l] = isDynamic
	}

	return rows, allDynamic
}

type benchmarkRowConfig struct {
	rctx           *reactively.ReactiveContext
	sources        []*reactively.Reactive[int]
	counter        *int64
	staticFraction float64
	nSources       int64
	rand           *rand.Rand
}

func makeBenchmarkRow(cfg *benchmarkRowConfig) (row []*reactively.Reactive[int], isDynamic []bool) {
	row = make([]*reactively.Reactive[int], len(cfg.sources))
	isDynamic = make([]bool, len(cfg.sources))

	for myDex := range cfg.sources {
		mySources := make([]*reactively.Reactive[int], 0, cfg.nSources)
		for sourceDex := 0; sourceDex < int(cfg.nSources); sourceDex++ {
			x := (myDex + sourceDex) % len(cfg.sources)
			y := cfg.sources[x]
			mySources = append(mySources, y)
		}

		staticNode := cfg.rand.Float64() < cfg.staticFraction
		if staticNode {
			// static node, always reference sources
			row[myDex] = reactively.Memo(cfg.rctx, func() int {
				*cfg.counter++
				// log.Printf("static %d", *cfg.counter)
				sum := 0
				for _, source := range mySources {
					sum += source.Read()
				}
				return sum
			})
		} else {
			first := mySources[0]
			tail := mySources[1:]
			row[myDex] = reactively.Memo(cfg.rctx, func() int {
				*cfg.counter++
				// log.Printf("dynamic %d", *cfg.counter)
				sum := first.Read()
				shouldDrop := sum&0x1 > 0
				dropDex := sum % len(tail)

				for i := 0; i < len(tail); i++ {
					if shouldDrop && i == dropDex {
						continue
					}
					sum += tail[i].Read()
				}
				return sum
			})
			isDynamic[myDex] = true
		}
	}

	return
}
