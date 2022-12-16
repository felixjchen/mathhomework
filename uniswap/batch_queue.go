package uniswap

// func StartBatchQueue(sugar *zap.SugaredLogger, batchChan chan BatchCandidate, nonceCounter *counter.TSCounter, executeCounter *counter.TSCounter, gasEstimate *eth.EstimateFee, gasEstimateMu *sync.Mutex, balanceOf *big.Int, balanceOfMu *sync.Mutex) {
// 	// TODO USE PQ
// 	// pq := make(util.PriorityQueue, 0)

// 	q := []BatchCandidate{}
// 	mu := sync.Mutex{}

// 	newWeb3 := blockchain.GetWeb3()

// 	// Entering PQ job
// 	go func() {
// 		for b := range batchChan {
// 			mu.Lock()
// 			q = append(q, b)
// 			mu.Unlock()
// 		}
// 	}()

// 	// Update PQ job

// 	// Batching top n job
// 	go func() {
// 		for {
// 			time.Sleep(time.Second)
// 			mu.Lock()
// 			if len(q) < config.MAX_BATCH_SIZE {
// 				continue
// 			}
// 			sort.Slice(q, func(i, j int) bool {
// 				return new(big.Int).Sub(q[i].NetProfit, q[j].NetProfit).Sign() == 1
// 			})
// 			for i := 1; i < config.MAX_BATCH_SIZE; i++ {
// 				start, end := 0, i+1
// 				batch := []Cycle{}
// 				usedPairsMap := make(map[Pair]bool)
// 				relaventPairs := []Pair{}
// 				for _, candidate := range q[start:end] {
// 					isUnique := true
// 					for _, pair := range candidate.Cycle.Edges {
// 						_, exists := usedPairsMap[pair]
// 						if exists {
// 							isUnique = false
// 						}
// 					}
// 					if isUnique {
// 						for _, pair := range candidate.Cycle.Edges {
// 							usedPairsMap[pair] = true
// 							relaventPairs = append(relaventPairs, pair)
// 						}
// 						batch = append(batch, candidate.Cycle)
// 					}
// 				}
// 				if len(batch) <= 1 {
// 					continue
// 				}
// 				pairToReserves := GetReservesForPairs(relaventPairs)
// 				amountIn_ := []*big.Int{}
// 				targets_ := [][]common.Address{}
// 				amountsOut_ := [][][2]*big.Int{}
// 				arbProfit_ := new(big.Int)

// 				for _, cycle := range batch {
// 					E0, E1 := GetE0E1ForCycle(cycle, pairToReserves)
// 					if new(big.Int).Sub(E0, E1).Sign() == -1 {
// 						amountIn := GetOptimalAmountIn(E0, E1)
// 						if amountIn.Sign() == 1 {
// 							amountsOut := GetAmountsOutCycle(pairToReserves, amountIn, cycle)
// 							arbProfit := big.NewInt(0).Sub(amountsOut[len(amountsOut)-1], amountIn)
// 							if arbProfit.Sign() == 1 {
// 								arbProfit_.Add(arbProfit_, arbProfit)
// 								targets := GetCycleTargets(cycle)
// 								cycleAmountsOut := GetCycleAmountsOut(cycle, amountsOut)
// 								amountIn_ = append(amountIn_, amountIn)
// 								targets_ = append(targets_, targets)
// 								amountsOut_ = append(amountsOut_, cycleAmountsOut)
// 							}
// 						}
// 					}
// 				}

// 				// run bundle
// 				executor, err := newWeb3.Eth.NewContract(config.BUNDLE_EXECTOR_ABI, config.Get().BUNDLE_EXECUTOR_ADDRESS.Hex())
// 				if err != nil {
// 					panic(err)
// 				}
// 				data, err := executor.EncodeABI("hi2", amountIn_, targets_, amountsOut_)
// 				if err != nil {
// 					panic(err)
// 				}
// 				call := &types.CallMsg{
// 					From: newWeb3.Eth.Address(),
// 					To:   executor.Address(),
// 					Data: data,
// 					Gas:  types.NewCallMsgBigInt(big.NewInt(types.MAX_GAS_LIMIT)),
// 				}
// 				gasLimit, err := newWeb3.Eth.EstimateGas(call)

// 				if err != nil {
// 					sugar.Error("ERROR IN BATCH Q")
// 					sugar.Error(err)
// 				} else {
// 					gasEstimateMu.Lock()
// 					maxGasWei := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasEstimate.MaxFeePerGas)
// 					netProfit := new(big.Int).Sub(arbProfit_, maxGasWei)
// 					sugar.Info("BATCH of len ", len(batch), " Estimated Profit ", arbProfit_, " SUB GAS ", netProfit)
// 					gasEstimateMu.Unlock()

// 					if netProfit.Sign() == 1 {
// 						gasTipCap := gasEstimate.MaxPriorityFeePerGas
// 						gasFeeCap := gasEstimate.MaxFeePerGas

// 						nonceCounter.Lock()
// 						nonce := nonceCounter.Get()
// 						nonceCounter.Unlock()
// 						hash, err := newWeb3.Eth.SendRawEIP1559TransactionWithNonce(
// 							nonce,
// 							executor.Address(),
// 							new(big.Int),
// 							gasLimit,
// 							gasTipCap,
// 							gasFeeCap,
// 							data,
// 						)
// 						if err != nil {
// 							fmt.Println("PANIC", err)
// 						} else {
// 							nonceCounter.Lock()
// 							nonceCounter.Inc()
// 							nonceCounter.Unlock()
// 							executeCounter.TSInc()
// 							sugar.Info("BATCH tx hash: ", hash)
// 						}
// 					}
// 				}
// 			}

// 			mu.Unlock()
// 		}
// 	}()
// }
