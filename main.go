package main

import (
	"arbitrage_go/programs"
	"flag"
)

func main() {
	arbitragePtr := flag.Bool("arb", false, "arbitrage program")
	backrunPtr := flag.Bool("br", false, "backrun program")
	flag.Parse()

	if *arbitragePtr {
		programs.Arbitrage()
	}
	if *backrunPtr {
		programs.Mempool()
	}
}
