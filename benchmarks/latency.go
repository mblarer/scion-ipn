package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"testing"

	"github.com/mblarer/scion-ipn"
	"github.com/mblarer/scion-ipn/filter"
	"github.com/mblarer/scion-ipn/internal"
	"github.com/mblarer/scion-ipn/segment"
	"github.com/scionproto/scion/go/lib/addr"
)

type doublepipe struct {
	io.Reader
	io.Writer
}

func main() {
	k, hops, enum := argsOrExit()

	f, _ := os.Create("cpu.prof")
	defer f.Close()
	_ = pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	result := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r1, w1 := io.Pipe()
			r2, w2 := io.Pipe()
			p1 := doublepipe{r1, w2}
			p2 := doublepipe{r2, w1}

			srcIA, _ := addr.IAFromString("1-ffaa:0:1")
			core1, _ := addr.IAFromString("1-ffaa:0:1000")
			core2, _ := addr.IAFromString("2-ffaa:0:1")
			dstIA, _ := addr.IAFromString("2-ffaa:0:1000")

			segments := make([]segment.Segment, 0)
			segments = append(segments, internal.CreateSegments(k, hops, srcIA, core1)...)
			segments = append(segments, internal.CreateSegments(k, hops, core1, core2)...)
			segments = append(segments, internal.CreateSegments(k, hops, core2, dstIA)...)

			var cfilter, sfilter segment.Filter
			switch enum {
			case "n":
				cfilter = filter.FromFilters()
				sfilter = filter.FromFilters()
			case "c":
				cfilter = filter.SrcDstPathEnumerator(srcIA, dstIA)
				sfilter = filter.FromFilters()
			case "s":
				cfilter = filter.FromFilters()
				sfilter = filter.SrcDstPathEnumerator(srcIA, dstIA)
			}

			client := ipn.Initiator{
				SrcIA:    srcIA,
				DstIA:    dstIA,
				Segments: segments,
				Filter:   cfilter,
			}
			server := ipn.Responder{
				Filter: sfilter,
			}

			go func() { _, _ = server.NegotiateOver(p1) }()
			_, _ = client.NegotiateOver(p2)
		}
	})

	f, _ = os.Create("mem.prof")
	defer f.Close()
	runtime.GC()
	_ = pprof.WriteHeapProfile(f)

	fmt.Println(result.N, int64(result.T))
}

func argsOrExit() (int, int, string) {
	if len(os.Args) != 4 {
		usageAndExit()
	}
	k, err := strconv.Atoi(os.Args[1])
	if err != nil {
		usageAndExit()
	}
	hops, err := strconv.Atoi(os.Args[2])
	if err != nil {
		usageAndExit()
	}
	enum := os.Args[3]
	if enum != "n" && enum != "c" && enum != "s" {
		usageAndExit()
	}
	return k, hops, enum
}

func usageAndExit() {
	fmt.Println("wrong command line arguments:", os.Args[0], "k:int hops:int n|c|s")
	os.Exit(1)
}
