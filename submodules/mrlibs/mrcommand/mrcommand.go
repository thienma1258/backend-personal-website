package mrcommand

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

const (
	COMMAND_REBUILD_CACHE                                   = "rebuildCache"
	COMMAND_REFRESH_QUERY_VERSION                           = "refreshQueryVersion"
	COMMAND_REFRESH_NEWS                                    = "refreshNews"
	COMMAND_REFRESH_CLIENT_CONFIG                           = "refreshClientConfig"
	COMMAND_REFRESH_CLIENT_SCRIPT                           = "refreshClientScript"
	COMMAND_REFRESH_CROSS_SEARCH                            = "refreshCrossSearch"
	COMMAND_REFRESH_BLACKLIST                               = "refreshBlacklist"
	COMMAND_REFRESH_SOURCE                                  = "refreshSource"
	COMMAND_CLEAR_INTERNAL_CACHE                            = "clearCache"
	COMMAND_PRINT_USAGE                                     = "printUsage"
	COMMAND_UPDATE_IP_DB                                    = "updateIPDB"
	COMMAND_FORCE_RELOAD_IP_DB                              = "reloadIPDB"
	COMMAND_SLAVE_REFRESH_CROSS_SEARCH                      = "slaveRefreshCrossSearch"
	COMMAND_SLAVE_REFRESH_LICENSE                           = "slaveRefreshLicense"
	COMMAND_SLAVE_REFRESH_MRSOURCE_CACHE_SEARCH             = "slaveRefreshMRSourceCacheSearch"
	COMMAND_SLAVE_REFRESH_MRSOURCE_CACHE_SERIES             = "slaveRefreshMRSourceCacheSeries"
	COMMAND_SLAVE_REFRESH_CACHE_SYNC_CATALOG                = "slaveRefreshCache_SyncCatalog"
	COMMAND_SLAVE_REFRESH_CACHE_FULL_CATALOG                = "slaveRefreshCache_FullCatalog"
	COMMAND_SLAVE_REFRESH_CACHE_LATEST_UPDATE_CATALOG       = "slaveRefreshCache_LatestUpdateCatalog"
	COMMAND_SLAVE_REFRESH_CACHE_MRSOURCE_LATEST_UPDATE_FEED = "slaveRefreshCache_MRSourceLatestUpdateFeed"
	COMMAND_SLAVE_REFRESH_CACHE_WEEKLY_FEATURE              = "slaveRefreshCache_WeekyFeature"
	COMMAND_SLAVE_REFRESH_CACHE_REALTIME_FEATURE            = "slaveRefreshCache_RealtimeFeature"
	COMMAND_REFRESH_FOR_YOU_LIST                            = "refreshForYouList"
	COMMAND_REFRESH_ADS_CONFIGS                             = "refreshAdsConfigs"
	COMMAND_REFRESH_ANDROID_UPGRADE                         = "refreshAndroidUpgrade"
	COMMAND_REFRESH_BETA_VERSION                            = "refreshBetaVersion"
	COMMAND_REFRESH_STICKER_PACKS                           = "refreshStickerPacks"
	COMMAND_REFRESH_WALLPAPER_LIST                          = "refreshWallpaperList"
	COMMAND_REFRESH_ADDONS_PUSH_PAYLOADS                    = "refreshAddonsPushPayloads"
	COMMAND_REFRESH_USER_CONTENT_AUDIENCE                   = "refreshUserContentAudience"
	COMMAND_REFRESH_TEST_USER_CONTENT_AUDIENCE              = "refreshTestUserContentAudience"
	COMMAND_REBUILD_AUDIENCE_CACHE                          = "rebuildAudienceCache"
	COMMAND_REBUILD_TEST_AUDIENCE_CACHE                     = "rebuildTestAudienceCache"
	COMMAND_REFRESH_SUBSCRIPTION_CONFIG                     = "refreshSubscriptionConfig"
)

var RedisCommandAddress string
var RedisCommandChannel string

type CommandListenerFunc func()

var listenerLock sync.RWMutex
var listeners = make(map[string]CommandListenerFunc)

var commandSender *redis.Client

var initialized = false

func initRedisSubscribeClient() {
	pubsubConn := redis.NewClient(&redis.Options{
		Addr:     RedisCommandAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	var pubsub *redis.PubSub
	pubsub = pubsubConn.Subscribe(RedisCommandChannel)
	defer pubsub.Close()
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			pubsub.Ping()
			// log.Printf("redis command ping %v %v", t, err)
		}
	}()
	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			log.Panic(err)
		}
		if msg.Channel == RedisCommandChannel {
			log.Printf("Redis received %v", msg)
			go handleCommand(msg.Payload)
		}
	}
}

// Init init the module
func Init() {
	if len(RedisCommandAddress) == 0 || len(RedisCommandChannel) == 0 {
		log.Fatal("In order to use mrrcommand module, you have to set valid ENV keys REDIS_MRCOMMAND_ADDRESS, REDIS_MRCOMMAND_CHANNEL.")
		return
	}

	go initRedisSubscribeClient()
	initialized = true
}

func init() {
	RedisCommandAddress = os.Getenv("REDIS_MRCOMMAND_ADDRESS")
	RedisCommandChannel = os.Getenv("REDIS_MRCOMMAND_CHANNEL")
}

func handleCommand(command string) {
	listenerLock.RLock()
	callback, ok := listeners[command]
	listenerLock.RUnlock()
	if ok && callback != nil {
		log.Printf("handleCommand %s ", command)
		callback()
	}
}

// SubscribeCommand subscribe listener to a specific command
func SubscribeCommand(command string, listener CommandListenerFunc) {
	if !initialized {
		log.Fatal("Please init the mrcommand module first.")
		return
	}
	listenerLock.Lock()
	listeners[command] = listener
	listenerLock.Unlock()
}

// UnsubscribeCommand subscribe listener to a specific command
func UnsubscribeCommand(command string) {
	listenerLock.Lock()
	listeners[command] = nil
	listenerLock.Unlock()
}

// SendCommand send a command to server
func SendCommand(command string) int {
	if commandSender == nil {
		commandSender = redis.NewClient(&redis.Options{
			Addr:     RedisCommandAddress,
			Password: "", // no password set
			DB:       0,  // use default DB
		})
	}
	result := commandSender.Publish(RedisCommandChannel, command)
	return int(result.Val())
}
