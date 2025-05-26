package cache

import (
	"testing"
)

// can not set struct
// https://github.com/go-redis/redis/issues/1504
func TestCache(t *testing.T) {
	//InitRedis()
	//dao.InitDb()
	//InitLocalCache()
	////events := dao.FindAllEvents()
	////cacheAllEventsWithGiven(events)
	//cacheAllEvents()
	////event = dao.DbpEvent{ID: 1, Event: "page_view", Description: "page view", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	//allEvents := getAllEventsLocal()
	//if allEvents != nil {
	//	println("---------- events: --------------")
	//	for event, _ := range *allEvents {
	//		print(event, ", ")
	//	}
	//}
	//println("page_view exists: ", EventExists("page_view"))
	//println()
	//
	//go listenEventChange()
	//
	//InitLocalCache()
	//fields := dao.FindAllFields()
	//cacheAllFieldLocalWithGiven(fields)
	//allField := GetAllFieldLocal()
	//if allField != nil {
	//	println("---------- field: --------------")
	//	for _, f := range *allField {
	//		fmt.Println(f)
	//	}
	//}
	//println()
	//
	//var eventMap = make(map[string]int)
	//eventMap["page_view"] = 1
	//eventMap["button_click"] = 2
	//
	//i, ok := eventMap["page_view"]
	//if ok {
	//	print(i)
	//}
}
