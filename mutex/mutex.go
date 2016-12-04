package mutex

import (
	"strings"
	"time"

	"strconv"

	"fmt"
)

var (
	ExistenceToken = "1"
	MutexToken     = "1"

	ExistenceKey    = "existence"
	MutexNamespace  = "mutex"
	TakenMutexesKey = "taken"

	//ExistenceKeyTTL   = 24 * time.Hour
	MutexKeyTTL       = 10 * time.Minute
	StaleMutexTimeout = 15 * time.Second
)

type Mutex struct {
	redis   Redis
	name    string
	takenAt int64
	//ttl     time.Duration
}

func New(redis Redis, name string) *Mutex {
	return &Mutex{
		redis: redis,
		name:  name,
	}
}

//	LUA:
//	GETSET existence
//	EXPIRE existence
//	RPUSH TSList
//	EXPIRE TSList
const lockLUA = `return redis.call('GETSET', KEYS[1], ARGV[1]) ~= ARGV[1] and
redis.call('EXPIRE', KEYS[1], ARGV[2]) and
redis.call('RPUSH', KEYS[2], ARGV[3]) and
redis.call('EXPIRE', KEYS[2], ARGV[4]) and 1 or 0`

func (m *Mutex) Lock() (bool, error) {
	//existenceKeyTTL := fmt.Sprintf("%v", int(ExistenceKeyTTL.Seconds()))
	mutexKeyTTL := fmt.Sprintf("%v", int(MutexKeyTTL.Seconds()))
	keys := []string{m.mutexExistenceKey(), m.mutexKey()}
	args := []string{ExistenceToken, mutexKeyTTL, MutexToken, mutexKeyTTL}

	reply, err := m.redis.EvalScript(lockLUA, keys, args)
	_, err = m.redis.Bool(reply, err)

	if err != nil {
		return false, err
	}

	//if !res {
	//
	//}

	if token, err := m.redis.BRPopLPush(m.mutexKey(), m.mutexKeyWithTS()); err == nil {
		if token == MutexToken {
			return true, nil
		}
		//	TODO: else what?
	} else {
		return false, err
	}

	return false, err
}

// TODO: should unlock only if owner?
const unlockLUA = `return redis.call('RPOPLPUSH', KEYS[1], KEYS[2]) and
redis.call('EXPIRE', KEYS[2], ARGV[1])`

// Move value from TS list to mutex list and expire it
func (m *Mutex) Unlock() error {
	mutexKeyTTL := fmt.Sprintf("%v", int(MutexKeyTTL.Seconds()))
	keys := []string{m.mutexKeyWithTS(), m.mutexKey()}
	args := []string{mutexKeyTTL}

	_, err := m.redis.EvalScript(unlockLUA, keys, args)

	return err
}

func (m *Mutex) mutexExistenceKey() string {
	return namespacedKey(m.name, ExistenceKey)
}

func (m *Mutex) mutexKey() string {
	return namespacedKey(m.name)
}

func (m *Mutex) mutexKeyWithTS() string {
	if m.takenAt == 0 {
		m.takenAt = time.Now().UnixNano()
	}
	return namespacedKey(TakenMutexesKey, m.name, strconv.FormatInt(m.takenAt, 10))
}

func namespacedKey(tokens ...string) string {
	keyParts := []string{MutexNamespace}

	for _, token := range tokens {
		keyParts = append(keyParts, token)
	}
	return strings.Join(keyParts, ":")
}

//func ReleaseStaleMutexes(client Redis) error {
//	keys, err := findStaleMutexes(client)
//
//	if err != nil {
//		return err
//	}
//
//	for _, key := range keys {
//		mtx := New(client, key.name, key.takenAt)
//		err = mtx.Unlock()
//	}
//
//	return nil
//}
//
//type staleMutexEntry struct {
//	name    string
//	takenAt int64
//}

//func findStaleMutexes(client Redis) ([]staleMutexEntry, error) {
//	pattern := namespacedKey(TakenMutexesKey, "*")
//	cursor := 0
//
//	keys := []staleMutexEntry{}
//	scanBatchSize := 1000
//
//	for {
//		response, err := redis.Values(client.execute("SCAN", cursor, "MATCH", pattern, "COUNT", scanBatchSize))
//
//		if err != nil {
//			return []staleMutexEntry{}, err
//		}
//
//		cursor, _ := redis.Int(response[0], nil)
//		values, _ := redis.Strings(response[1], nil)
//
//		for _, key := range values {
//			// extract mutex name w/o namespaces and TS
//			keyAndTS := extractKeyAndTS(key)
//
//			// parse timestamp
//			mutexTakenAt, err := parseTS(keyAndTS[1])
//
//			if err != nil {
//				// TODO: notify somehow about broken mutex record?
//				continue
//			}
//
//			// if mutex is stale, collect it
//			if mutexTakenAt+StaleMutexTimeout.Nanoseconds() < int64(time.Now().UnixNano()) {
//				keys = append(keys, staleMutexEntry{name: keyAndTS[0], takenAt: mutexTakenAt})
//			}
//		}
//
//		if cursor == 0 {
//			break
//		}
//	}
//
//	return keys, nil
//}

//func extractKeyAndTS(token string) []string {
//	tokens := strings.Split(token, ":")
//	l := len(tokens)
//	return tokens[l-2 : l]
//}
//
//func parseTS(token string) (int64, error) {
//	i, err := strconv.ParseInt(token, 10, 64)
//	if err != nil {
//		return 0, err
//	}
//	return i, nil
//}
