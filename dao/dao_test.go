package dao

import (
	"encoding/json"
	"testing"
)

func TestJson(t *testing.T) {
	dataMap := make(map[string]interface{})
	dataMap["event"] = "page_view"
	dataMap["lib"] = "js"
	dataMap["lib_version"] = "1.18.17"
	dataMap["os"] = "ios"
	dataMap["lib_method"] = "code"
	dataMap["screen_height"] = 1080

	marshal, err := json.Marshal(dataMap)
	if err == nil {
		println(string(marshal))
	} else {
		print("error: ", err.Error())
	}
}
