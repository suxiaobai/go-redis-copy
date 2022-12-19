package main

import (
	"flag"
	"go-redis-copy/cmd"
	"go-redis-copy/src/pusher"
	"go-redis-copy/src/scanner"
	"log"
	"sync"
)

func main() {

	pattern := flag.String("pattern", "*", "Only transfer matching keys (default *)")
	scanCount := flag.Int("scan_count", 1000, "Redis scan count option (default: 1000) ")
	reportCount := flag.Int("report_count", 10000, "Migrate report count option (default: 10000) ")
	source := flag.String("src", "", "source redis connection uris")
	sourcePassword := flag.String("src_pass", "", "source redis password")
	destination := flag.String("dest", "", "destination redis connection uris")
	destinationPassword := flag.String("dest_pass", "", "destination redis password")
	checkRoutines := flag.Int("check_pool_size", 30, "check redis exists exector pool size")
	exportRoutines := flag.Int("pull_pool_size", 30, "get redis info exector pool size")
	pushRoutines := flag.Int("push_pool_size", 30, "load redis key exector pool size")
	skipCheck := flag.Bool("skip_check", false, "skip check and copy keys")

	flag.Parse()

	if *source == "" || *destination == "" {
		log.Fatal("<src> and <dest> must be redis address, like 127.0.0.1:6379 or cluster: 127.0.0.1:6379,127.0.0.1:6380")
	}

	log.Println("Start copying")

	redisScanner := scanner.NewScanner(
		cmd.NewRedisClient(*source, "", *sourcePassword),
		scanner.RedisScannerOpts{
			Pattern:          *pattern,
			ScanCount:        *scanCount,
			ReportCount:      *reportCount,
			PullRoutineCount: *exportRoutines,
		},
	)

	redisPusher := pusher.NewRedisPusher(cmd.NewRedisClient(*destination, "", *destinationPassword), redisScanner.GetDumpChannel())

	waitingGroup := new(sync.WaitGroup)

	redisChecker := pusher.NewRedisChecker(cmd.NewRedisClient(*destination, "", *destinationPassword), redisScanner.GetCheckChannel(), redisScanner.GetKeyChannel(), *skipCheck)
	go redisChecker.Start(*checkRoutines)

	redisPusher.Start(waitingGroup, *pushRoutines)

	redisScanner.Start()
	waitingGroup.Wait()

	log.Println("Finish copying")

}
