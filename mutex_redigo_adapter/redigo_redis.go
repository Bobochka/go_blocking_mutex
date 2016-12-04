package mutex_redigo_adapter

import (
	"log"
	"time"

	"fmt"

	"github.com/garyburd/redigo/redis"
)

type Redis struct {
	pool *redis.Pool
	log  bool
}

func New(server string, db string, poolSize int) Redis {
	return Redis{pool: newRedisPool(server, db, poolSize)}
}

func (r Redis) BRPopLPush(key string, key2 string) (string, error) {
	val, err := redis.String(r.execute("BRPOPLPUSH", key, key2, 0))

	// TODO: for now hardcoded
	//	if len(val) != 2 {
	//		return "", errors.New("BRPopLPush didn't get value from the list")
	//	}

	return val, err
}

func (r Redis) Bool(reply interface{}, err error) (bool, error) {
	return redis.Bool(reply, err)
}

func (r Redis) EvalScript(lua string, keys []string, args []string) (interface{}, error) {
	conn := r.pool.Get()
	defer conn.Close()

	script := redis.NewScript(len(keys), lua)

	var keysAndArgs []interface{}
	for _, v := range append(keys, args...) {
		keysAndArgs = append(keysAndArgs, v)
	}

	return script.Do(conn, keysAndArgs...)
}

// TODO: extract config
func newRedisPool(server string, db string, poolSize int) *redis.Pool {
	return redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", server, redis.DialConnectTimeout(20*time.Millisecond))

		if err != nil {
			log.Println(err.Error())
			return nil, err
		}
		_, err = c.Do("SELECT", db)
		if err != nil {
			c.Close()
			log.Println(err.Error())
			return nil, err
		}
		return c, err
	}, poolSize)
}

func (r Redis) execute(command string, args ...interface{}) (interface{}, error) {
	c := r.pool.Get()
	defer c.Close()

	r.logger(command, args)

	data, err := c.Do(command, args...)
	if err != nil {
		log.Println(err.Error())
	}
	return data, err
}

func (r Redis) logger(args ...interface{}) {
	if r.log {
		log.Println("[Redis] " + fmt.Sprintln(args...))
	}
}
