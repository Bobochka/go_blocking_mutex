package mutex_goredis_adapter

import (
	"time"

	"strconv"

	"gopkg.in/redis.v5"
)

type Redis struct {
	client *redis.Client
}

func New(server string, db string, poolSize int) Redis {
	return Redis{client: newRedisClient(server, db, poolSize)}
}

func (r Redis) BRPopLPush(source string, destination string) (string, error) {
	// TODO: mb add max wait?
	return r.client.BRPopLPush(source, destination, 0).Result()
}

func (r Redis) Bool(reply interface{}, err error) (bool, error) {
	return reply.(int64) == 1, err
}

func (r Redis) EvalScript(lua string, keys []string, args []string) (interface{}, error) {
	script := redis.NewScript(lua)

	var arguments []interface{}
	for _, v := range args {
		arguments = append(arguments, v)
	}

	cmd := script.Run(r.client, keys, arguments...)

	response, err := cmd.Result()

	if err == redis.Nil {
		return response, nil
	} else {
		return response, err
	}

}

func newRedisClient(server string, db string, poolSize int) *redis.Client {
	dbNum, _ := strconv.Atoi(db)

	return redis.NewClient(&redis.Options{
		Addr:         server,
		Password:     "",
		MaxRetries:   3,
		DB:           dbNum,
		PoolSize:     poolSize,
		PoolTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		DialTimeout:  2 * time.Second,
	})

}
