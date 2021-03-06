package main

import (
	"flag"
	"github.com/ezeql/koding-challange/common"
	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

var r *Redis

func TestMain(m *testing.M) {
	flag.Parse()
	var err error

	r, err = OpenRedis(*redisHost, *redisPort)
	if err != nil {
		log.Fatalln("Cannot dial redis", err)
	}
	defer r.Close()
	r.Do("FLUSHALL")
	os.Exit(m.Run())
}

func TestHalfMonthProcess(t *testing.T) {
	var m common.MetricEntry

	//initial time in seconds for current bucket
	current := begginingOfBucketUnix(time.Now().UTC())

	lastSecondPreviousBucket := time.Unix(current-1, 0)

	//get unixtime start date for the previous finished bucket
	prev := begginingOfBucketUnix(lastSecondPreviousBucket)

	day := time.Unix(prev, 0)

	for i := 0; i < int(bucketLength); i++ {
		m = common.MetricEntry{"John", 1, "call-api-x", &day}
		r.processMetric(m)
		m = common.MetricEntry{"Charles", 1, "login", &day}
		r.processMetric(m)
		day = day.Add(time.Hour * 24)
	}

	dest, err := r.processBuckets()
	resultMap, err := redis.IntMap(r.Do("ZRANGE", dest, 0, -1, "WITHSCORES"))

	assert.Nil(t, err, "must be nil")
	assert.Equal(t, len(resultMap), 2, "must be ...")
	assert.EqualValues(t, resultMap["call-api-x"], bucketLength, "must be ...")
	assert.EqualValues(t, resultMap["login"], bucketLength, "must be ...")

}
