package pusher

import (
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
)

type RedisChecker struct {
	client       redis.UniversalClient
	checkChannel <-chan string
	keyChannel   chan<- string
}

func NewRedisChecker(client redis.UniversalClient, cch <-chan string, kch chan<- string) *RedisChecker {
	return &RedisChecker{
		client:       client,
		checkChannel: cch,
		keyChannel:   kch,
	}
}

func (c *RedisChecker) Start(number int) {
	wg := new(sync.WaitGroup)
	wg.Add(number)
	for i := 0; i < number; i++ {
		go c.checkRoutine(wg)
	}
	wg.Wait()
	close(c.keyChannel)
}

func (c *RedisChecker) checkRoutine(wg *sync.WaitGroup) {
	for key := range c.checkChannel {
		n, err := c.client.Exists(ctx, key).Result()
		if err != nil {
			log.Fatal(err)
		}

		if n == int64(0) {
			c.keyChannel <- key
		}
	}
	wg.Done()
}
