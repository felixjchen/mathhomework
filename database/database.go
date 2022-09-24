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

func NewDBConn() DB {
	client := redis.NewClient(&redis.Options{
		Addr:        "0.0.0.0:6379",
		Password:    "", // no password set
		DB:          0,  // use default DB
		ReadTimeout: 15 * time.Second,
	})
	sugar := logging.GetSugar()
	return DB{client, sugar}
}

func (db *DB) GetCycleHashes() []string {
	val, err := db.client.Get("INDEX").Result()
	if err != nil {
		panic(err)
	}
	return strings.Split(val, " ")
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
		cycles = append(cycles, CycleFromB64(cmd.(*redis.StringCmd).Val()))
	}
	return cycles
}

func (db *DB) GetCycle(k string) uniswap.Cycle {
	val, err := db.client.Get(k).Result()
	if err != nil {
		panic(err)
	}
	return CycleFromB64(val)
}

func (db *DB) AddCycle(cycle uniswap.Cycle) {
	k := hashCycle(cycle)
	v := CycleToB64(cycle)
	_, err := db.client.Get(k).Result()
	if err != nil {
		db.sugar.Info("Adding to Redis:", cycle)
		pipe := db.client.Pipeline()
		pipe.Set(k, v, 0)
		pipe.Append("INDEX", k+" ")
		_, err := pipe.Exec()
		if err != nil {
			panic(err)
		}
	} else {
		db.sugar.Info("Already In Redis:", cycle)
	}
}

func hashCycle(c uniswap.Cycle) string {
	h, err := hashstructure.Hash(c, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	return strconv.FormatUint(uint64(h), 10)
}

// go binary encoder
func CycleToB64(m uniswap.Cycle) string {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(m)
	if err != nil {
		fmt.Println(`failed gob Encode`, err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

// go binary decoder
func CycleFromB64(str string) uniswap.Cycle {
	m := uniswap.Cycle{}
	by, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		fmt.Println(`failed base64 Decode`, err)
	}
	b := bytes.Buffer{}
	b.Write(by)
	d := gob.NewDecoder(&b)
	err = d.Decode(&m)
	if err != nil {
		fmt.Println(`failed gob Decode`, err)
	}
	return m
}
