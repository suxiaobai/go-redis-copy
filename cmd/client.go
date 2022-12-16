package cmd

import (
	"strings"

	"github.com/go-redis/redis/v8"
)

// func NewRedisClient(url string) *redis.Client {
// 	opt, err := redis.ParseURL(url)
// 	if err != nil {
// 		panic(err)
// 	}

// 	opt.PoolSize = 200
// 	opt.ReadTimeout = -1
// 	opt.WriteTimeout = -1
// 	return redis.NewClient(opt)
// }

// func NewRedisClusterClient(url string) *redis.ClusterClient {
// 	redis.UniversalClient
// 	opt, err := redis.ParseURL(url)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return redis.NewClusterClient(&redis.ClusterOptions{
// 		Addrs:    []string{opt.Addr},
// 		Username: opt.Username,
// 		Password: opt.Password,
// 		PoolSize: 200,
// 	})
// }

func NewRedisClient(addr string, username string, password string) redis.UniversalClient {
	return redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        strings.Split(addr, ","),
		Username:     username,
		Password:     password,
		PoolSize:     200,
		ReadTimeout:  -1,
		WriteTimeout: -1,
	})
}

// func CreateClient(t string, url string) interface{} {
// 	switch t {
// 	case "standone":
// 		return NewRedisClient(url)
// 	case "cluster":
// 		return NewRedisClusterClient(url)
// 	}
// 	panic("redis deploy type must be standone or cluster")
// }
