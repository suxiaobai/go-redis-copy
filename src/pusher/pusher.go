package pusher

import (
	"context"
	"errors"
	"go-redis-copy/src/scanner"
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisPusher struct {
	client       redis.UniversalClient
	checkChannel <-chan string
	dumpChannel  <-chan scanner.KeyDump
}

func NewRedisPusher(client redis.UniversalClient, dumpChannel <-chan scanner.KeyDump) *RedisPusher {
	return &RedisPusher{
		client:      client,
		dumpChannel: dumpChannel,
	}
}

func (p *RedisPusher) Start(wg *sync.WaitGroup, number int) {
	wg.Add(number)
	for i := 0; i < number; i++ {
		go p.pushRoutine(wg)
	}

}

func (p *RedisPusher) pushRoutine(wg *sync.WaitGroup) {
	for dump := range p.dumpChannel {
		err := p.loadKey(dump)
		if err != nil {
			log.Fatal(err)
		}

		if dump.Ttl > 0 {
			err = p.loadKeyTtl(dump)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	wg.Done()
}

func (p *RedisPusher) loadKey(dump scanner.KeyDump) error {

	switch dump.Type {
	case "string":
		return p.client.Set(ctx, dump.Key, dump.Value, 0).Err()
	case "set":
		return p.client.SAdd(ctx, dump.Key, dump.Value).Err()
	case "list":
		return p.client.LPush(ctx, dump.Key, dump.Value).Err()
	case "hash":
		return p.client.HMSet(ctx, dump.Key, dump.Value).Err()
	case "zset":
		var zv []*redis.Z
		for _, z := range dump.Value.([]redis.Z) {
			newZ := new(redis.Z)
			*newZ = z
			zv = append(zv, newZ)
		}
		return p.client.ZAdd(ctx, dump.Key, zv...).Err()
	default:
		return errors.New("not support redis type")
	}
}

func (p *RedisPusher) loadKeyTtl(dump scanner.KeyDump) error {
	err := p.client.Expire(ctx, dump.Key, dump.Ttl).Err()
	if err != nil {
		log.Fatal(err)
	}

	return err
}
