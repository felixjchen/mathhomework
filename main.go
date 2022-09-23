package main

import (
	"arbitrage_go/programs"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	arbitrage3Ptr := flag.Bool("arb3", false, "arbitrage3 program")
	arbitrage2Ptr := flag.Bool("arb2", false, "arbitrage2 program")
	arbitragePtr := flag.Bool("arb", false, "arbitrage program")
	backrunPtr := flag.Bool("br", false, "backrun program")
	flag.Parse()

	if *arbitragePtr {
		go programs.ArbitrageMain()
	}
	if *arbitrage2Ptr {
		go programs.Arbitrage2Main()
	}
	if *arbitrage3Ptr {
		go programs.Arbitrage3Main()
	}
	if *backrunPtr {
		go programs.Mempool()
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println(sig)
		done <- true
	}()

	fmt.Println("awaiting signal")
	<-done
	fmt.Println("exiting")
}
