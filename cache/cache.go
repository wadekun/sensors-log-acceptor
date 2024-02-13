package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
	"liangck.xyz/data-service/sensors-log-acceptor/configer"
	"liangck.xyz/data-service/sensors-log-acceptor/dao"
	"liangck.xyz/data-service/sensors-log-acceptor/logger"
	"time"
)

// -------------------- Cache. Implemented by Redis(https://github.com/go-redis/redis)
//       and go-cache(https://github.com/patrickmn/go-cache)
//------------------------

// 设计思路：
//	方案一：
//  	采用 redis缓存 + 进程内缓存（go-cache）结合的方式（有点类似于计算机中的多级缓存）
// 		因为需要缓存结构化的元数据信息，缓存到redis时，有序列化（写入）和反序列化（读取）的过程，（在频繁的写入/读取时）会比较耗费性能，
// 		所以**较复杂的元数据信息**采用进程内缓存的方式（dbp_field、dbp_event_field），
// 		缓存**结构简单的的元数据**缓存在redis中，直接判断（dbp_event、dbp_field_enum_values均可缓存为set，直接判断）.
//	方案二：
//		采用进程内缓存（数据）+ redis缓存（数据版本号）的方式
// 		元数据全部缓存在进程内。在redis中维护版本号，数据采集节点subscribe元数据版本号，在后台管理服务中修改元数据时publish新的版本号到redis。采集节点
// 		接收到变更消息，刷新本地元数据缓存。这样每次对上报日志数据校验的时候，只需要读取本地缓存，减少对redis的依赖，而且由于埋点元数据量理论上不会很大，
//		进程内缓存完全可以扛得住。
//
// Q1：进程内缓存如何保证与数据库的一致性？
// A1：1.采用版本号的方式，在redis中每张元数据表维护一个对应的自增版本号，每次修改该表数据时自增（incr），服务内每次查询缓存前先判断redis中的版本号与服务内保存的版本号是否一致，
//          不一致则查询数据库重新缓存。这样即使服务部署多个实例，也可以保持与数据库数据的一致性。
//     2.元数据通过管理接口修改时，直接同步更新掉进程内缓存，不过这种方式在元数据管理和数据采集拆分为两个服务、服务部署多个实例时 都不可行。

const KeyDelimiter = ":"
const KeyPrefix = "DBP:META_CACHE:"
const EventName = "Event"
const FieldName = "Field"
const ValueName = "Value"
const Field = "Field"
const EventField = "EventField"
const FieldEnumValue = "FieldEnumValue"
const EventKey = KeyPrefix + EventName
const FieldKey = KeyPrefix + FieldName
const Version = "version"
const Topic = "Topic"
const Change = "Change"
const EventChangeTopic = KeyPrefix + EventName + KeyDelimiter + Change + KeyDelimiter + Topic
const FieldChangeTopic = KeyPrefix + Field + KeyDelimiter + Change + KeyDelimiter + Topic
const EventFieldChangeTopic = KeyPrefix + EventField + KeyDelimiter + Change + KeyDelimiter + Topic
const FieldEnumValueChangeTopic = KeyPrefix + FieldEnumValue + KeyDelimiter + Change + KeyDelimiter + Topic

// Redis Client
var redisClient *redis.Client
var ctx context.Context

// go-cache
var localCache *cache.Cache

// InitRedis init redis client
func InitRedis(config *configer.Config) {
	ctx = context.TODO()
	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,     // "192.168.3.193:6379",
		Password: config.RedisPassword, // "redis@123",
		DB:       config.RedisDB,       // 7,
	})
}

// InitLocalCache init local cache instance
func InitLocalCache(config *configer.Config) {
	localCache = cache.New(24*time.Hour, 24*time.Hour)
}

// Init All
func Init(config *configer.Config) {
	InitRedis(config)
	InitLocalCache(config)

	events := dao.FindAllEvents()
	cacheAllEventsLocalWithGiven(events)
	for _, event := range *events {
		cacheAllEventFieldsLocal(event.Event)
	}
	fields := dao.FindAllFields()
	cacheAllFieldLocalWithGiven(fields)
	for _, field := range *fields {
		cacheAllEnumValuesByField(field.Field)
	}
	// listen metadata change and flush local cache
	go listenEventChange()
	go listenFieldChange()
	go listenEventFieldChange()
	go listenFieldValueChange()
}

// 缓存指定数据

// 表：dbp_events
// 缓存方式：Redis、local
// 数据结构：set
// key：DBP:META_CACHE:Events
func cacheAllEvents() {
	events := dao.FindAllEvents()
	cacheAllEventsLocalWithGiven(events)
}

// 本地缓存所有事件
func cacheAllEventsLocalWithGiven(events *[]dao.DbpEvent) {
	var eventsSlice = make(map[string]int)
	for idx, event := range *events {
		eventsSlice[event.Event] = idx
	}
	localCache.Set(EventKey, &eventsSlice, cache.NoExpiration)
}

// 从本地缓存获取所有定义的事件
func getAllEventsLocal() *map[string]int {
	if x, found := localCache.Get(EventKey); found {
		events := x.(*map[string]int)
		return events
	}

	return nil
}

// EventExists 判断事件是否存在
func EventExists(event string) bool {
	eventsLocal := getAllEventsLocal()
	if eventsLocal != nil {
		if _, ok := (*eventsLocal)[event]; ok {
			return true
		}
	}

	return false
}

// 监听事件变更并刷新缓存
func listenEventChange() {
	logger.Logger.Info("Subscribe EventChangeTopic : " + EventChangeTopic)
	pubSub := redisClient.Subscribe(ctx, EventChangeTopic)
	defer pubSub.Close()
	ch := pubSub.Channel()
	for msg := range ch {
		logger.Logger.Info("receive change message: " + msg.Payload)
		cacheAllEvents()
	}
}

// ------ Redis -----
func cacheAllEventsWithGiven(events *[]dao.DbpEvent) {
	var eventsSlice = make([]string, 5)
	for _, event := range *events {
		eventsSlice = append(eventsSlice, event.Event)
	}
	redisClient.SAdd(ctx, EventKey, eventsSlice)
}

func getAllEvents() []string {
	stringSliceCmd := redisClient.SMembers(ctx, EventKey)
	if stringSliceCmd.Err() != nil {
		logger.Logger.Info(stringSliceCmd.Err().Error())
	} else {
		return stringSliceCmd.Val()
	}

	return nil
}

// 表：dbp_fields
// 缓存方式：进程内，go-cache
// 数据结构：Struct
// key：DBP:META_CACHE:Field
func cacheAllFieldsLocal() {
	fields := dao.FindAllFields()
	cacheAllFieldLocalWithGiven(fields)
}

// 本地缓存所有字段元数据
func cacheAllFieldLocalWithGiven(fields *[]dao.DbpField) {
	localCache.Set(FieldKey, fields, cache.NoExpiration)
}

// GetAllFieldLocal 从本地缓存获取所有字段元数据
func GetAllFieldLocal() *[]dao.DbpField {
	if x, found := localCache.Get(FieldKey); found && x != nil {
		fields := x.(*[]dao.DbpField)
		return fields
	}
	return nil
}

// 监听字段元数据变更
func listenFieldChange() {
	logger.Logger.Info("Subscribe FieldChangeTopic : " + FieldChangeTopic)
	pubSub := redisClient.Subscribe(ctx, FieldChangeTopic)
	defer pubSub.Close()
	ch := pubSub.Channel()
	for msg := range ch {
		logger.Logger.Info("receive topic " + FieldChangeTopic + "message: " + msg.Payload)
		cacheAllFieldsLocal()
	}
}

// 表：dbp_event_fields
// 缓存方式：进程内，go-cache
// 数据结构：Struct
// key: DBP:META_CACHE:{event}:Fields
func cacheAllEventFieldsLocal(event string) {
	fieldByEvent := dao.FindAllEventFieldByEvent(event)
	cacheAllEventFieldLocalWithGiven(event, fieldByEvent)
}

// 本地缓存所有事件字段
func cacheAllEventFieldLocalWithGiven(event string, eventField *[]dao.DbpEventField) {
	eventFieldKey := getEventFiledCacheKey(event)
	localCache.Set(eventFieldKey, eventField, cache.NoExpiration)
}

func getEventFiledCacheKey(event string) string {
	eventFieldKey := KeyPrefix + event + KeyDelimiter + FieldName
	return eventFieldKey
}

// 从本地缓存获取所有事件字段
func getEventFieldLocalByEvent(event string) *[]dao.DbpEventField {
	key := getEventFiledCacheKey(event)
	if x, found := localCache.Get(key); found {
		fields := x.(*[]dao.DbpEventField)
		return fields
	}
	return nil
}

// 监听事件字段元数据变更
func listenEventFieldChange() {
	logger.Logger.Info("Subscribe EventFieldChangeTopic : " + EventFieldChangeTopic)
	pubSub := redisClient.Subscribe(ctx, EventFieldChangeTopic)
	defer pubSub.Close()
	ch := pubSub.Channel()
	for msg := range ch {
		logger.Logger.Info("receive topic " + string(EventFieldChangeTopic) + "message: " + msg.Payload)
		cacheAllEventFieldsLocal(msg.Payload)
	}
}

// 表：dbp_field_enum_values
// 缓存方式：redis
// 数据结构：set
// key：DBP:META_CACHE:{field}:Value
func cacheAllEnumValuesByField(field string) {
	valuesByField := dao.FindAllEnumValuesByField(field)
	cacheAllEnumValuesLocalWithGiven(field, valuesByField)
}

// 本地缓存所有枚举值
func cacheAllEnumValuesLocalWithGiven(field string, enumValues *[]dao.DbpFieldEnumValue) {
	cacheKey := getFieldEnumValueCacheKey(field)
	var valueMap = make(map[string]int)
	for idx, enumValue := range *enumValues {
		valueMap[enumValue.EnumValue] = idx
	}
	localCache.Set(cacheKey, &valueMap, cache.NoExpiration)
}

// GetAllEnumValuesLocalByField 从本地缓存获取所有枚举值
func GetAllEnumValuesLocalByField(field string) *map[string]int {
	cacheKey := getFieldEnumValueCacheKey(field)
	if x, found := localCache.Get(cacheKey); found {
		fieldValuesMap := x.(*map[string]int)
		return fieldValuesMap
	}

	return nil
}

func getFieldEnumValueCacheKey(field string) string {
	cacheKey := KeyPrefix + field + KeyDelimiter + ValueName
	return cacheKey
}

// FieldEnumValueExists 判断枚举值是否存在
func FieldEnumValueExists(field string, value string) bool {
	enumValuesMap := GetAllEnumValuesLocalByField(field)
	if enumValuesMap != nil {
		if _, ok := (*enumValuesMap)[value]; ok {
			return true
		}
	}

	return false
}

// 监听事件字段元数据变更
func listenFieldValueChange() {
	logger.Logger.Info("Subscribe FieldEnumValueChangeTopic : " + FieldEnumValueChangeTopic)
	pubSub := redisClient.Subscribe(ctx, FieldEnumValueChangeTopic)
	defer pubSub.Close()
	ch := pubSub.Channel()
	for msg := range ch {
		logger.Logger.Info("receive topic " + FieldEnumValueChangeTopic + "message: " + msg.Payload)
		cacheAllEnumValuesByField(msg.Payload)
	}
}

// SendFieldChangeMessage publish change message to channel
func SendFieldChangeMessage() {
	redisClient.Publish(ctx, FieldChangeTopic, "field change")
}

func SendEventChangeMessage() {
	err := redisClient.Publish(ctx, EventChangeTopic, "event change").Err()
	if err != nil {
		logger.Logger.Error("failed to publish message to " + EventChangeTopic + " caused: " + err.Error())
	}
}

func SendEventFieldChangeMessage(event string) {
	redisClient.Publish(ctx, EventFieldChangeTopic, event)
}

func SendFieldValuesChangeMessage(field string) {
	redisClient.Publish(ctx, FieldEnumValueChangeTopic, field)
}

// redis -----------
func cacheAllEnumValuesWithGiven(field string, enumValues *[]dao.DbpFieldEnumValue) {
	cacheKey := getFieldEnumValueCacheKey(field)
	var valueSlice = make([]string, 5)
	for _, enumValue := range *enumValues {
		valueSlice = append(valueSlice, enumValue.EnumValue)
	}
	redisClient.SAdd(ctx, cacheKey, valueSlice)
}
