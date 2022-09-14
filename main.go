package main

import (
	"arbitrage_go/programs"
	"flag"
)

func main() {
	arbitrage2Ptr := flag.Bool("arb2", false, "arbitrage2 program")
	arbitragePtr := flag.Bool("arb", false, "arbitrage program")
	backrunPtr := flag.Bool("br", false, "backrun program")
	flag.Parse()

	if *arbitragePtr {
		programs.ArbitrageMain()
	}
	if *arbitrage2Ptr {
		programs.Arbitrage2Main()
	}
	if *backrunPtr {
		programs.Mempool()
	}
}
