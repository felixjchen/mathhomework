package database

import (
	"arbitrage_go/logging"
	"arbitrage_go/uniswap"
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/mitchellh/hashstructure/v2"
	"go.uber.org/zap"
)

type DB struct {
	client *redis.Client
	sugar  *zap.SugaredLogger
}

const CYCLE_INDEX = "CYCLEINDEX"
const PAIR_INDEX = "PAIRINDEX"

func NewDBConn() DB {
	client := redis.NewClient(&redis.Options{
		Addr:        "0.0.0.0:6379",
		Password:    "", // no password set
		DB:          0,  // use default DB
		ReadTimeout: 15 * time.Second,
	})
	sugar := logging.GetSugar("db")
	return DB{client, sugar}
}

func (db *DB) GetCycleHashes() []string {
	val, err := db.client.Get(CYCLE_INDEX).Result()
	if err != nil {
		panic(err)
	}
	cycleHashes := strings.Split(val, " ")
	return cycleHashes[:len(cycleHashes)-1]
}

func (db *DB) GetPairs() []uniswap.Pair {
	val, err := db.client.Get(PAIR_INDEX).Result()
	if err != nil {
		panic(err)
	}
	pairHashes := strings.Split(val, " ")
	pairHashes = pairHashes[:len(pairHashes)-1]
	pairs := []uniswap.Pair{}
	for _, pairHash := range pairHashes {
		pairB64, err := db.client.Get(pairHash).Result()
		if err != nil {
			panic(err)
		}
		pairs = append(pairs, StructFromB64(pairB64, uniswap.Pair{}))
	}
	return pairs
}

func (db *DB) GetCycles(hashes []string) []uniswap.Cycle {
	pipe := db.client.Pipeline()
	for _, hash := range hashes {
		pipe.Get(hash)
	}
	cmds, err := pipe.Exec()
	if err != nil {
		panic(err)
	}
	cycles := []uniswap.Cycle{}
	for _, cmd := range cmds {
		cycles = append(cycles, StructFromB64(cmd.(*redis.StringCmd).Val(), uniswap.Cycle{}))
	}
	return cycles
}

func (db *DB) GetCycle(k string) uniswap.Cycle {
	val, err := db.client.Get(k).Result()
	if err != nil {
		panic(err)
	}
	return StructFromB64(val, uniswap.Cycle{})
}

func (db *DB) AddCycle(cycle uniswap.Cycle) {
	cycleHash := hashStruct(cycle)
	_, err := db.client.Get(cycleHash).Result()
	if err != nil {
		db.sugar.Info("Adding to Redis:", cycle)
		pipe := db.client.Pipeline()
		cycleB64 := structToB64(cycle)
		pipe.Set(cycleHash, cycleB64, 0)
		pipe.Append(CYCLE_INDEX, cycleHash+" ")

		for _, pair := range cycle.Edges {
			pairHash := hashStruct(pair)
			_, err := db.client.Get(pairHash).Result()
			if err != nil {
				pairB64 := structToB64(pair)
				pipe.Set(pairHash, pairB64, 0)
				pipe.Append(PAIR_INDEX, pairHash+" ")
			}
		}

		_, err := pipe.Exec()
		if err != nil {
			panic(err)
		}
	} else {
		db.sugar.Info("Already In Redis:", cycle)
	}
}

func hashStruct[T any](t T) string {
	h, err := hashstructure.Hash(t, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	return strconv.FormatUint(uint64(h), 10)
}

// go binary encoder
func structToB64[T any](m T) string {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(m)
	if err != nil {
		fmt.Println(`failed gob Encode`, err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

// go binary decoder
func StructFromB64[T any](str string, empty T) T {
	by, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		fmt.Println(`failed base64 Decode`, err)
	}
	b := bytes.Buffer{}
	b.Write(by)
	d := gob.NewDecoder(&b)
	err = d.Decode(&empty)
	if err != nil {
		fmt.Println(`failed gob Decode`, err)
	}
	return empty
}
