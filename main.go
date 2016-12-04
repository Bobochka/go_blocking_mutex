package main

import (
	"log"
	"mutex/mutex"
	"mutex/mutex_goredis_adapter"
)

func main() {
	//redigo := mutex_redigo_adapter.New("127.0.0.1:6379", "0", 3)
	//mtx1 := mutex.New(redigo, "asdf")

	goredis := mutex_goredis_adapter.New("127.0.0.1:6379", "1", 3)
	mtx2 := mutex.New(goredis, "asdf")

	//mtx1.Lock()
	//log.Println("Mutexified")
	//err := mtx1.Unlock()
	//
	//if err != nil {
	//	log.Println(err.Error())
	//}

	mtx2.Lock()
	log.Println("Mutexified 2")
	err := mtx2.Unlock()

	if err != nil {
		log.Println(err.Error())
	}
}
