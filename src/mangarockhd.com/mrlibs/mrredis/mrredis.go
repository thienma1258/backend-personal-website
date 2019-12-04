package mrredis

import (
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"mangarockhd.com/mrlibs/mrutils"
)

const luaScript = `
	local results = {}
	for i = 1, table.getn(KEYS) do
		results[i] = redis.call("HMGET", KEYS[i], unpack(ARGV))
	end
	return results
`

var luaSHA = make(map[int]string)
var luaLock sync.RWMutex

type HashCacheItem struct {
	CacheKey   string
	CacheField string
	Value      string
}
type HashCacheMultipleFieldsItem struct {
	CacheKey    string
	CacheFields []string
	Values      map[string]([]byte)
}

type RedisConnectionSetting struct {
	Address string
	DB      int
	Timeout time.Duration
	client  *redis.Client
}

var redisClients = make(map[int]*RedisConnectionSetting)

func RegisterRedisConnection(connection int, setting *RedisConnectionSetting) {
	conn := redis.NewClient(&redis.Options{
		Addr:     setting.Address,
		Password: "",         // no password set
		DB:       setting.DB, // use default DB
	})
	setting.client = conn
	redisClients[connection] = setting
}

func getClient(conn int) *redis.Client {
	if client, ok := redisClients[conn]; ok {
		return client.client
	}
	return nil
}

func getTimeout(conn int) time.Duration {
	if client, ok := redisClients[conn]; ok {
		return client.Timeout
	}
	return time.Minute * 10
}

func applyTimeout(conn int, client *redis.Client, cKey string) {
	timeout := getTimeout(conn)
	if timeout > 0 {
		client.Expire(cKey, timeout)
	}
}
func GetRedisClient(conn int) *redis.Client {
	return getClient(conn)
}

func SetTimeout(conn int, cKey string, value interface{}, timeout time.Duration) {
	client := getClient(conn)
	client.Set(cKey, value, timeout)
}

func HGetZ(conn int, cKey string, cField string) []byte {
	client := getClient(conn)
	data, err := client.HGet(cKey, cField).Bytes()
	if err != nil {
		return nil
	}
	return mrutils.DecompressToString(data)
}

func HGet(conn int, cKey string, cField string) []byte {
	// log.Printf("HGet conn=%d cKey=%s cField=%s", conn, cKey, cField)
	client := getClient(conn)
	data, err := client.HGet(cKey, cField).Bytes()
	if err != nil {
		return nil
	}
	return data
}
func HGetAll(conn int, cKey string) map[string]string {
	client := getClient(conn)
	data := client.HGetAll(cKey)
	return data.Val()
}

func HMGet(conn int, cKey string, cFields ...string) map[string]string {
	client := getClient(conn)
	result := client.HMGet(cKey, cFields...)
	if result.Err() != nil {
		return nil
	}
	results := make(map[string]string)
	values := result.Val()
	// result.String()
	for i := 0; i < len(cFields); i++ {
		cField := cFields[i]
		if values[i] == nil {
			results[cField] = ""
		} else {
			results[cField] = fmt.Sprintf("%s", values[i])
		}
	}
	return results
}

func MGet(conn int, cKeys []string) map[string]string {
	client := getClient(conn)
	result := client.MGet(cKeys...)
	if result.Err() != nil {
		return nil
	}
	results := make(map[string]string)
	values := result.Val()
	for i := 0; i < len(cKeys); i++ {
		ckey := cKeys[i]
		if values[i] == nil {
			results[ckey] = ""
		} else {
			results[ckey] = fmt.Sprintf("%s", values[i])
		}
	}
	return results
}

func HGetUInt32(conn int, cKey string, cField string, defaultValue uint32) uint32 {
	client := getClient(conn)
	data, err := client.HGet(cKey, cField).Bytes()
	if err != nil {
		return defaultValue
	}
	result, err := strconv.Atoi(string(data))
	if err != nil {
		return defaultValue
	}
	return uint32(result)
}

func HGetMultiple(conn int, items []*HashCacheItem) {
	client := getClient(conn)
	pipe := client.Pipeline()
	total := len(items)
	cmds := make([](*redis.StringCmd), total)
	for i := 0; i < total; i++ {
		item := items[i]
		cmds[i] = pipe.HGet(item.CacheKey, item.CacheField)
	}
	pipe.Exec()
	for i := 0; i < total; i++ {
		items[i].Value = cmds[i].Val()
	}
}

func TTLMultiple(conn int, keys []string) map[string]int {
	client := getClient(conn)
	pipe := client.Pipeline()
	total := len(keys)
	cmds := make([](*redis.DurationCmd), total)
	for i := 0; i < total; i++ {
		key := keys[i]
		cmds[i] = pipe.TTL(key)
	}
	pipe.Exec()
	results := map[string]int{}
	for i := 0; i < total; i++ {
		results[keys[i]] = int(cmds[i].Val())
	}
	return results
}

func HGetMultipleKeysWithOneField(conn int, cacheKeys []string, cacheField string) map[string]string {
	client := getClient(conn)
	pipe := client.Pipeline()
	total := len(cacheKeys)
	cmds := make([](*redis.StringCmd), total)
	for i := 0; i < total; i++ {
		cmds[i] = pipe.HGet(cacheKeys[i], cacheField)
	}
	pipe.Exec()
	results := map[string]string{}
	for i := 0; i < total; i++ {
		results[cacheKeys[i]] = cmds[i].Val()
	}
	return results
}

func HGetMultipleFields(conn int, items []*HashCacheMultipleFieldsItem) {
	client := getClient(conn)
	pipe := client.Pipeline()
	cmds := make([](*redis.SliceCmd), len(items))
	for i := 0; i < len(items); i++ {
		item := items[i]
		cmds[i] = pipe.HMGet(item.CacheKey, item.CacheFields...)
	}
	pipe.Exec()
	for i := 0; i < len(items); i++ {
		items[i].Values = make(map[string]([]byte))
		total := len(items[i].CacheFields)
		values := cmds[i].Val()
		for j := 0; j < total; j++ {
			field := items[i].CacheFields[j]
			switch v := values[j].(type) {
			case int:
				items[i].Values[field] = values[j].([]byte)
			case string:
				items[i].Values[field] = []byte(values[j].(string))
			case nil:
				items[i].Values[field] = nil
			default:
				fmt.Printf("I don't know about type %T!\n", v)
			}
		}
	}
}

func HGetMultipleFieldsLuaScript(conn int,
	cKeys []string, fields []string,
) *map[string](map[string]([]byte)) {
	client := getClient(conn)
	var err error
	luaLock.RLock()
	sha := luaSHA[conn]
	luaLock.RUnlock()

	if sha == "" {
		sha, err = client.ScriptLoad(luaScript).Result()
		if err != nil {
			log.Fatalf("Error while loading script %v\n", err)
		}
		luaLock.Lock()
		luaSHA[conn] = sha
		luaLock.Unlock()
	}

	argv := make([]interface{}, len(fields))
	for i := 0; i < len(fields); i++ {
		argv[i] = fields[i]
	}

	results, err := client.EvalSha(sha, cKeys, argv...).Result()
	if err != nil {
		log.Printf("Error while load from cache %v\n", err)
		return nil
	}

	values := results.([]interface{})
	total := len(values)
	totalFields := len(fields)
	items := make(map[string](map[string]([]byte)))

	for i := 0; i < total; i++ {
		item := make(map[string]([]byte))

		val := values[i].([]interface{})
		for j := 0; j < totalFields; j++ {
			field := fields[j]
			switch v := val[j].(type) {
			case int:
				item[field] = val[j].([]byte)
			case string:
				item[field] = []byte(val[j].(string))
			case nil:
				item[field] = nil
			default:
				fmt.Printf("I don't know about type %T!\n", v)
			}
		}
		items[cKeys[i]] = item
	}
	return &items
}

func HGetInt(conn int, cKey string, cField string, defaultValue int) int {
	client := getClient(conn)
	data, err := client.HGet(cKey, cField).Bytes()
	if err != nil {
		return defaultValue
	}
	result, err := strconv.Atoi(string(data))
	if err != nil {
		return defaultValue
	}
	return result
}

func HGetIntMultipleKeysWithOneField(conn int, cKeys []string, cField string, defaultValue int) map[string]int {
	total := len(cKeys)
	if total == 0 {
		return nil
	}
	client := getClient(conn)
	pipe := client.Pipeline()
	cmds := make([](*redis.StringCmd), total)
	for i := 0; i < total; i++ {
		cmds[i] = pipe.HGet(cKeys[i], cField)
	}
	pipe.Exec()
	results := make(map[string]int)
	for i := 0; i < total; i++ {
		result, err := strconv.Atoi(cmds[i].Val())
		if err != nil {
			results[cKeys[i]] = defaultValue
		} else {
			results[cKeys[i]] = result
		}
	}
	return results
}

func HGetString(conn int, cKey string, cField string, defaultValue string) string {
	client := getClient(conn)
	data, err := client.HGet(cKey, cField).Bytes()
	if err != nil {
		return defaultValue
	}

	return string(data)
}

func HSetUInt32(conn int, cKey string, cField string, value uint32) {
	client := getClient(conn)
	client.HSet(cKey, cField, string(strconv.Itoa(int(value))))
	applyTimeout(conn, client, cKey)
}

func HSetInt(conn int, cKey string, cField string, value int) {
	client := getClient(conn)
	client.HSet(cKey, cField, string(strconv.Itoa(value)))
	applyTimeout(conn, client, cKey)
}

func GetZ(conn int, cKey string) []byte {
	client := getClient(conn)
	data, err := client.Get(cKey).Bytes()
	if err != nil {
		return nil
	}
	return mrutils.DecompressToString(data)
}

func Get(conn int, cKey string) []byte {
	// log.Printf("Get conn=%d cKey=%s", conn, cKey)
	client := getClient(conn)
	data, err := client.Get(cKey).Bytes()
	if err != nil {
		return nil
	}
	return data
}

func GetInt(conn int, cKey string) int {
	client := getClient(conn)
	data, err := client.Get(cKey).Int64()
	if err != nil {
		return 0
	}
	return int(data)
}

func HSetZ(conn int, cKey string, cField string, value []byte) {
	client := getClient(conn)
	client.HSet(cKey, cField, string(mrutils.CompressBytes(value)))
	applyTimeout(conn, client, cKey)
}

func HMSet(conn int, cKey string, fields map[string]interface{}) {
	if len(fields) == 0 {
		debug.PrintStack()
		mrutils.Log("HMSet with empty fields")
	}
	client := getClient(conn)
	result := client.HMSet(cKey, fields)
	if result.Err() != nil {
		mrutils.Log("HMSet error %v", result.Err())
	}

	applyTimeout(conn, client, cKey)
}

func HSet(conn int, cKey string, cField string, value interface{}) {
	client := getClient(conn)
	client.HSet(cKey, cField, value)
	applyTimeout(conn, client, cKey)
}

func SAddString(conn int, cKey string, values ...string) {
	client := getClient(conn)
	data := make([]interface{}, len(values))
	for i := 0; i < len(values); i++ {
		data[i] = values[i]
	}
	client.SAdd(cKey, data...)
	applyTimeout(conn, client, cKey)
}

func SisMember(conn int, cKey string, value interface{}) bool {
	client := getClient(conn)
	isMember := client.SIsMember(cKey, value)
	return isMember.Val()
}

func MSIsMember(conn int, cKeys []string, value interface{}) map[string]bool {
	client := getClient(conn)
	pipe := client.Pipeline()
	cmds := make([]*redis.BoolCmd, len(cKeys))
	for i, cKey := range cKeys {
		cmds[i] = pipe.SIsMember(cKey, value)
	}
	_, err := pipe.Exec()
	if err != nil {
		log.Printf("MSIsMember error=%v", err)
	}
	results := make(map[string]bool)
	for i, cKey := range cKeys {
		results[cKey] = cmds[i].Val()
	}
	return results
}

func SMembers(conn int, cKey string) []string {
	client := getClient(conn)
	sets, err := client.SMembers(cKey).Result()
	if err != nil {
		log.Printf("SMembers error=%v", err)
	}
	return sets
}

func MSMembers(conn int, cKeys []string) map[string][]string {
	client := getClient(conn)
	pipe := client.Pipeline()
	cmds := make([]*redis.StringSliceCmd, len(cKeys))

	results := make(map[string][]string)
	for i, cKey := range cKeys {
		cmds[i] = pipe.SMembers(cKey)
		results[cKey] = make([]string, 0)
	}
	_, err := pipe.Exec()
	if err != nil {
		log.Printf("MSMembers pipeline exec error=%v", err)
		return nil
	}

	for i := 0; i < len(cmds); i++ {
		results[cKeys[i]] = cmds[i].Val()
	}
	return results
}

func SUnion(conn int, sets ...string) []string {
	client := getClient(conn)
	return client.SUnion(sets...).Val()
}

func Exists(conn int, cKey string) int64 {
	client := getClient(conn)
	exists := client.Exists(cKey)
	return exists.Val()
}

func SetZ(conn int, cKey string, value []byte) {
	// log.Printf("SetZ conn=%d key=%s", conn, cKey)
	client := getClient(conn)
	client.Set(cKey, string(mrutils.CompressBytes(value)), getTimeout(conn))
}

func Set(conn int, cKey string, value interface{}) {
	// log.Printf("SetZ conn=%d key=%s", conn, cKey)
	client := getClient(conn)
	client.Set(cKey, value, getTimeout(conn))
}

func MSet(conn int, pairs []interface{}) {
	client := getClient(conn)
	client.MSet(pairs...)
	keys := []string{}
	total := len(pairs)
	for i := 0; i < total; i += 2 {
		key, ok := pairs[i].(string)
		if ok {
			keys = append(keys, key)
		}
	}
	if len(keys) > 0 {
		MExpire(conn, keys, getTimeout(conn))
	}
}

func SetInt(conn int, cKey string, value int) {
	client := getClient(conn)
	client.Set(cKey, strconv.Itoa(value), getTimeout(conn))
}

func DeletePattern(conn int, pattern string) {
	client := getClient(conn)
	data := client.Keys(pattern)
	if len(data.Val()) > 0 {
		client.Del(data.Val()...)
	}
}

func Delete(conn int, cKey string) {
	client := getClient(conn)
	client.Del(cKey)
}

func MDelete(conn int, cKeys []string) {
	client := getClient(conn)
	client.Del(cKeys...)
}

func HDelete(conn int, cKey string, cFields []string) {
	client := getClient(conn)
	client.HDel(cKey, cFields...)
}

func HMDelete(conn int, cKeys []string, cFields []string) {
	client := getClient(conn)
	pipe := client.Pipeline()
	for i := 0; i < len(cKeys); i++ {
		pipe.HDel(cKeys[i], cFields...)
	}
	_, err := pipe.Exec()
	if err != nil {
		log.Printf("HMDelete error=%v", err)
	}
}

func HMSetMultipleKeys(conn int, items map[string](map[string]interface{})) {
	client := getClient(conn)
	pipe := client.Pipeline()
	for cKey, values := range items {
		pipe.HMSet(cKey, values)
	}
	_, err := pipe.Exec()
	if err != nil {
		log.Printf("HMSetMultipleKeys error=%v", err)
	}
}

func HMSetMultipleKeysToOneField(conn int, cField string, items map[string]interface{}) {
	client := getClient(conn)
	pipe := client.Pipeline()
	for cKey, cValue := range items {
		pipe.HSet(cKey, cField, cValue)
	}
	_, err := pipe.Exec()
	if err != nil {
		log.Printf("HMSetMultipleKeysToOneField error=%v", err)
	}
}

func HMGetAll(conn int, cKeys []string) map[string]map[string]string {
	client := getClient(conn)
	pipe := client.Pipeline()
	cmds := make([]*redis.StringStringMapCmd, len(cKeys))
	for i, cKey := range cKeys {
		cmds[i] = pipe.HGetAll(cKey)
	}
	_, err := pipe.Exec()
	if err != nil {
		mrutils.Log("HMGetAll pipeline exec err %v", err)
		return map[string]map[string]string{}
	}

	results := make(map[string]map[string]string)
	for i := 0; i < len(cKeys); i++ {
		results[cKeys[i]] = make(map[string]string)
	}
	for i := 0; i < len(cKeys); i++ {
		hashResults := cmds[i].Val()
		for field, values := range hashResults {
			results[cKeys[i]][field] = values
		}
	}
	return results
}

func MSetRemoveOneMember(conn int, sets []string, member interface{}) {
	client := getClient(conn)
	pipe := client.Pipeline()
	cmds := make([]*redis.IntCmd, len(sets))
	for i, set := range sets {
		cmds[i] = pipe.SRem(set, member)
	}
	_, err := pipe.Exec()
	if err != nil {
		mrutils.Log("MSetRemoveMember pipeline exec error=%v", err)
	}
}

// LGetAllElements implement LRANGE(0, -1)
// to get all elements in given list
func LGetAllElements(conn int, key string) []string {
	client := getClient(conn)
	return client.LRange(key, 0, -1).Val()
}

// LPush with timeout
func LPush(conn int, key string, values []interface{}, timeout time.Duration) {
	client := getClient(conn)
	client.LPush(key, values...)
	client.Expire(key, timeout)
}

// LSetElements remove list, then add elements to new list
// to prevent duplicate elements
func LSetElements(conn int, key string, values []interface{}, timeout time.Duration) {
	client := getClient(conn)
	pipe := client.Pipeline()
	cmds := make([]interface{}, 3)
	cmds[0] = pipe.Del(key)
	cmds[1] = pipe.LPush(key, values...)
	cmds[2] = pipe.Expire(key, timeout)
	_, err := pipe.Exec()
	if err != nil {
		mrutils.Log("LSetElements pipeline exec error=%v", err)
	}
}

// MHExists field exists in multiple hashes
func MHExists(conn int, cKeys []string, field string) map[string]bool {
	client := getClient(conn)
	pipe := client.Pipeline()
	cmds := make([]*redis.BoolCmd, len(cKeys))
	for i, cKey := range cKeys {
		cmds[i] = pipe.HExists(cKey, field)
	}
	_, err := pipe.Exec()
	if err != nil {
		mrutils.Log("MHExists pipeline exec error=%v", err)
	}
	results := make(map[string]bool)
	for i := 0; i < len(cKeys); i++ {
		results[cKeys[i]] = cmds[i].Val()
	}
	return results
}

// Rename rename keys
func Rename(conn int, key, newKey string) {
	client := getClient(conn)
	client.Rename(key, newKey)
}

// Keys get all key
func Keys(conn int, pattern string) []string {
	client := getClient(conn)
	return client.Keys(pattern).Val()
}

// Scan return keys matching given pattern
func Scan(conn int, match string) []string {
	client := getClient(conn)
	var cursor uint64
	var keys []string
	cKeys := map[string]bool{}
	keys, cursor = client.Scan(cursor, match, 1000).Val()
	total := len(keys)
	for i := 0; i < total; i++ {
		cKeys[keys[i]] = true
	}

	for cursor != 0 {
		keys, cursor = client.Scan(cursor, match, 1000).Val()
		total = len(keys)
		if total > 0 {
			for i := 0; i < total; i++ {
				cKeys[keys[i]] = true
			}
		}
	}
	total = len(cKeys)
	results := make([]string, total)

	i := 0
	for key := range cKeys {
		results[i] = key
		i++
	}

	return results
}

// MExpire set expire for cKey
func Expire(conn int, cKey string, timeout time.Duration) {
	client := getClient(conn)
	if client != nil {
		client.Expire(cKey, timeout)
	}
}

// MExpire set expire for multiple keys
func MExpire(conn int, cKeys []string, timeout time.Duration) {
	client := getClient(conn)
	pipe := client.Pipeline()
	cmds := make([]*redis.BoolCmd, len(cKeys))
	for i := 0; i < len(cKeys); i++ {
		cmds[i] = pipe.Expire(cKeys[i], timeout)
	}
	_, err := pipe.Exec()
	if err != nil {
		mrutils.Log("MExpire pipeline exec error=%v", err)
	}
}
