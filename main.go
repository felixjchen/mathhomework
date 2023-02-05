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
	arbitragePtr := flag.Bool("arb", false, "arbitrage program")
	arbitragePornPtr := flag.Bool("arb_porn", false, "arbitrage porn program")
	cyclePtr := flag.Bool("cycle", false, "cycle detection program")
	backrunPtr := flag.Bool("br", false, "backrun program")
	flag.Parse()

	if *arbitragePtr {
		go programs.ArbitrageMain()
	}
	if *arbitragePornPtr {
		go programs.ArbitragePornMain()
	}
	if *cyclePtr {
		go programs.FindCycles()
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

	fmt.Println("awaiting cancel signal")
	<-done
	fmt.Println("exiting")
}
