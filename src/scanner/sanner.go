package scanner

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type KeyDump struct {
	Key   string
	Type  string
	Value interface{}
	Ttl   time.Duration
}

type RedisScannerOpts struct {
	Pattern          string
	ScanCount        int
	ReportCount      int
	PullRoutineCount int
}

type RedisScanner struct {
	client       redis.UniversalClient
	options      RedisScannerOpts
	checkChannel chan string
	keyChannel   chan string
	dumpChannel  chan KeyDump
}

func NewScanner(client redis.UniversalClient, options RedisScannerOpts) *RedisScanner {
	return &RedisScanner{
		client:       client,
		options:      options,
		dumpChannel:  make(chan KeyDump, 10000),
		keyChannel:   make(chan string, 10000),
		checkChannel: make(chan string, 10000),
	}
}

func (s *RedisScanner) Start() {
	wgPull := new(sync.WaitGroup)
	wgPull.Add(s.options.PullRoutineCount)

	go s.scanRoutine()
	for i := 0; i < s.options.PullRoutineCount; i++ {
		go s.exportRoutine(wgPull)
	}

	wgPull.Wait()
	close(s.dumpChannel)
}

func (s *RedisScanner) GetCheckChannel() <-chan string {
	return s.checkChannel
}

func (s *RedisScanner) GetKeyChannel() chan<- string {
	return s.keyChannel
}

func (s *RedisScanner) GetDumpChannel() <-chan KeyDump {
	return s.dumpChannel
}

func (s *RedisScanner) scanRoutine() {

	count := 0
	iter := s.client.Scan(ctx, 0, s.options.Pattern, int64(s.options.ScanCount)).Iterator()

	for iter.Next(ctx) {
		s.checkChannel <- iter.Val()
		count++

		if count%s.options.ReportCount == 0 {
			log.Printf("Already attempt => %d", count)
		}
	}

	if err := iter.Err(); err != nil {
		panic(err)
	}

	log.Printf("Already attempt => %d, finished", count)

	// var cursor uint64
	// for {
	// 	var keys []string
	// 	var err error

	// 	keys, cursor, err = s.client.Scan(ctx, cursor, s.options.Pattern, 0).Result()
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	for _, key := range keys {
	// 		s.keyChannel <- key
	// 	}

	// 	if cursor == 0 { // no more keys
	// 		break
	// 	}
	// }

	close(s.checkChannel)
}

func (s *RedisScanner) dumpKey(t string, key string) (interface{}, error) {

	switch t {
	case "string":
		return s.client.Get(ctx, key).Result()
	case "set":
		return s.client.SMembers(ctx, key).Result()
	case "list":
		return s.client.LRange(ctx, key, 0, -1).Result()
	case "hash":
		newH := make(map[string]string)
		iter := s.client.HScan(ctx, key, 0, "*", 10000).Iterator()
		count := 0
		var k string
		for iter.Next(ctx) {
			if count == 0 {
				k = iter.Val()
				count++
			} else {
				newH[k] = iter.Val()
				count = 0
			}
		}
		if err := iter.Err(); err != nil {
			return newH, err
		}
		return newH, nil
	case "zset":
		return s.client.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
			Min: "-inf",
			Max: "+inf",
		}).Result()
	}
	return nil, errors.New("not support redis type")
}

func (s *RedisScanner) exportRoutine(wg *sync.WaitGroup) {
	for key := range s.keyChannel {

		type_, err := s.client.Type(ctx, key).Result()
		if err != nil {
			panic(err)
		}

		ttl, err := s.client.TTL(ctx, key).Result()
		if err != nil {
			panic(err)
		}

		value, err := s.dumpKey(type_, key)
		if err != nil {
			log.Printf("redis type not support, key: %s, type: %s", key, type_)
			continue
		}

		s.dumpChannel <- KeyDump{
			Key:   key,
			Type:  type_,
			Ttl:   ttl,
			Value: value,
		}
	}

	wg.Done()
}
