package otakumo

import (
	"log"
	"os"
	"sync"

	"github.com/go-redis/redis"
)

var RedisOtakumoUpdateAddress string
var RedisOtakumoUpdateChannel string

type EntityUpdateListenerFunc func(otakumoID string)

var listenerLock sync.RWMutex
var listeners = make(map[string]EntityUpdateListenerFunc)
var masterListeners = []EntityUpdateListenerFunc{}

var updateSenderLock sync.RWMutex
var updateSender *redis.Client
var initialized = false

func initRedisSubscribeClient() {
	pubsubConn := redis.NewClient(&redis.Options{
		Addr:     RedisOtakumoUpdateAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	var pubsub *redis.PubSub
	pubsub = pubsubConn.Subscribe(RedisOtakumoUpdateChannel)
	defer pubsub.Close()
	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			log.Panic(err)
		}
		if msg.Channel == RedisOtakumoUpdateChannel {
			log.Printf("Redis received %v", msg)
			go handleEntityUpdatedCommand(msg.Payload)
		}
	}
}

// GetEntityType Get otakumod entity type from otakumo ID
func GetEntityType(otakumoID string) string {
	count := 0
	length := len(otakumoID)
	for i := 0; i < length; i++ {
		if otakumoID[i] == '-' {
			count++
		}
		if count == 2 {
			return otakumoID[0:i]
		}
	}
	return otakumoID
}

func handleEntityUpdatedCommand(otakumoID string) {
	entityType := GetEntityType(otakumoID)
	listenerLock.RLock()
	callback, ok := listeners[entityType]
	_masterListeners := masterListeners
	listenerLock.RUnlock()

	if ok && callback != nil {
		callback(otakumoID)
	}

	if _masterListeners != nil && len(_masterListeners) > 0 {
		for _, callback := range _masterListeners {
			callback(otakumoID)
		}
	}

}

// InitEntityUpdateListener Init Entity Update Listener
func InitEntityUpdateListener() {
	if len(RedisOtakumoUpdateAddress) == 0 || len(RedisOtakumoUpdateChannel) == 0 {
		log.Fatal("In order to use otakumo entity update listener module, you have to set valid ENV keys REDIS_OTAKUMO_UPDATE_ADDRESS, REDIS_OTAKUMO_UPDATE_CHANNEL.")
		return
	}
	go initRedisSubscribeClient()
	initialized = true
}
func init() {
	RedisOtakumoUpdateAddress = os.Getenv("REDIS_OTAKUMO_UPDATE_ADDRESS")
	RedisOtakumoUpdateChannel = os.Getenv("REDIS_OTAKUMO_UPDATE_CHANNEL")
}

// SubscribeEntityUpdated subscribe listener to all kinds of otakumo update object
func SubscribeEntityUpdated(listener EntityUpdateListenerFunc) {
	listenerLock.Lock()
	masterListeners = append(masterListeners, listener)
	listenerLock.Unlock()
}

// SubscribeCommand subscribe listener to a specific command
func SubscribeEntityType(entityType string, listener EntityUpdateListenerFunc) {
	if !initialized {
		log.Fatal("Please init the otakumo module first.")
		return
	}
	listenerLock.Lock()
	listeners[entityType] = listener
	listenerLock.Unlock()
}

// UnsubscribeCommand subscribe listener to a specific command
func UnsubscribeEntityType(entityType string) {
	listenerLock.Lock()
	listeners[entityType] = nil
	listenerLock.Unlock()
}

// SendCommand send a command to server
func SendEntityUpdated(otakumoID string) int {
	updateSenderLock.RLock()
	_updateSender := updateSender

	updateSenderLock.RUnlock()
	if _updateSender == nil {
		updateSenderLock.Lock()
		updateSender = redis.NewClient(&redis.Options{
			Addr:     RedisOtakumoUpdateAddress,
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		_updateSender = updateSender
		updateSenderLock.Unlock()
	}

	result := _updateSender.Publish(RedisOtakumoUpdateChannel, otakumoID)
	return int(result.Val())
}
