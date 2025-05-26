package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/Jeffail/gabs"
	"io/ioutil"
	"liangck.xyz/data-service/sensors-log-acceptor/cache"
	"liangck.xyz/data-service/sensors-log-acceptor/dao"
	"liangck.xyz/data-service/sensors-log-acceptor/kafka"
	"liangck.xyz/data-service/sensors-log-acceptor/logger"
	. "liangck.xyz/data-service/sensors-log-acceptor/model"
	"net/url"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"
)

const EventJsonPath string = "event"
const Event = "event"
const TypeEnum = "enum"
const TypeFloat = "float"
const TypeInt = "int"
const TypeBool = "bool"
const TypeString = "string"
const ReceiveTime = "receive_time"

// Handle 处理埋点数据请求
// todo: refactor
func Handle(jsonData []byte) (bool, error) {
	log, err := parseRequestData(string(jsonData))
	if err != nil {
		return false, err
	}

	// android ios 上传的是数组
	if log.Gzip != "" {
		decodeString, err := base64.StdEncoding.DecodeString(log.DataList)
		if err != nil {
			return false, err
		}

		decompress, err2 := GzipDecompress(decodeString)
		var logs []interface{}
		if err2 != nil {
			return false, err2
		}

		err3 := json.Unmarshal(decompress, &logs)
		if err3 != nil {
			return false, err3
		}

		for _, logData := range logs {
			marshaled, err := json.Marshal(logData)
			if err == nil {
				ok, err := ParseAndValidLogData(marshaled)
				if !ok && err != nil {
					logger.Logger.Error("valid field error: " + err.Error())
				} else {
					logger.Logger.Info("valid field successful: " + string(marshaled))
				}
			}
		}

		return true, nil
	}

	// js 上传的是单条数据
	decodeString, err := base64.StdEncoding.DecodeString(log.Data)
	if err == nil {
		return ParseAndValidLogData(decodeString)
	}

	return true, nil

}

// parse request data by construct a url. get data and ext
func parseRequestData(requestBody string) (*Log, error) {
	urlPrefix := "http://localhost:8080/test?"
	dummyUrl := urlPrefix + requestBody
	u, err := url.Parse(dummyUrl)
	if err == nil {
		data := u.Query().Get("data")
		//ext := u.Query().Get("ext")
		_gzip := u.Query().Get("gzip")
		dataList := u.Query().Get("data_list")
		crc := u.Query().Get("crc")
		log := &Log{Gzip: _gzip, DataList: dataList, Data: data, Crc: crc}
		return log, nil
	}

	return nil, err
}

// GzipDecompress decompress gzip data
func GzipDecompress(in []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		var out []byte
		return out, err
	}
	defer reader.Close()

	return ioutil.ReadAll(reader)
}

// validEvent
// valid event and add event to logger data
func validEvent(jsonParsed *gabs.Container, data *map[string]interface{}) (bool, error) {
	event, ok := jsonParsed.Path(EventJsonPath).Data().(string)
	if !ok {
		return false, errors.New("event field not found")
	}
	if !cache.EventExists(event) {
		return false, errors.New("Unknown event :" + event + "")
	}

	(*data)[Event] = event

	return true, nil
}

// validField
// 验证指定字段，如果验证通过则把该字段数据放入data字典中
func validField(jsonParsed *gabs.Container, data *map[string]interface{}, field dao.DbpField) *ValidResult {
	fieldValue := jsonParsed.Path(field.JsonPath).Data()
	if fieldValue == nil {
		// 1.非空校验
		if !field.Nullable {
			return &ValidResult{OK: false, Err: "field [" + field.Field + "] can not be null", ErrType: ValueCannotBeNull}
		}

		return &ValidResult{OK: true, ErrType: None}
	}
	fieldType := reflect.TypeOf(fieldValue)

	// 验证json格式字段
	if field.Type == "json" {
		return validJsonField(fieldValue, field, data)
	}

	// 如果是枚举，查询该字段配置的枚举值，判断上报值是否在枚举值中
	if field.Type == TypeEnum {
		exists := cache.FieldEnumValueExists(field.Field, fieldValue.(string))
		if !exists {
			return &ValidResult{OK: false, Err: "field: " + field.Name + " value: " + fieldValue.(string) + " not exists!", ErrType: ValueNotExist}
		}

		(*data)[field.Field] = fieldValue.(string)

		return &ValidResult{OK: true, ErrType: None}
	}

	// 其他类型：int、long、float、string、bool
	if fieldType.Name() != field.Type {
		// 2.类型校验
		// 数值类型（float、float64、int）如果不一样可以转换，其他情况报类型不匹配
		if !(fieldType.Name() == "float64" && (field.Type == "float" || field.Type == "int" || field.Type == "long")) {
			return &ValidResult{OK: false, Err: "Field: [" + field.Field + "] type mismatch. json value type is " + fieldType.Name() + " and dest type is " + field.Type, ErrType: TypeMisMatch}
		}
	}

	// 3. 取值 并进行 长度校验
	return extractValueAndValidLength(fieldValue, field, data)
}

func validJsonField(fieldValue interface{}, field dao.DbpField, data *map[string]interface{}) *ValidResult {
	// json 反序列化后，字段也为json会反序列化后为Map类型
	if reflect.TypeOf(fieldValue).Kind() == reflect.Map {
		strVal, err := json.Marshal(fieldValue)
		if err != nil {
			return &ValidResult{OK: false, Err: "field: " + field.Field + " invalid json string ", ErrType: InvalidFormat}
		}

		// 验证长度
		if utf8.RuneCountInString(string(strVal)) > field.Length {
			return &ValidResult{OK: false, Err: "field: " + field.Field + " value length large than " + strconv.Itoa(field.Length), ErrType: ValueTooLong}
		}

		(*data)[field.Field] = string(strVal)

		return &ValidResult{OK: true, ErrType: None}

	}

	if reflect.TypeOf(fieldValue).Kind() == reflect.String { // 加了转义符的还是会是string
		bytesVal := []byte(fieldValue.(string))
		if !json.Valid(bytesVal) {
			return &ValidResult{OK: false, Err: "field: " + field.Field + " invalid json string ", ErrType: InvalidFormat}
		}

		// 验证长度
		if utf8.RuneCountInString(string(bytesVal)) > field.Length {
			return &ValidResult{OK: false, Err: "field: " + field.Field + " value length large than " + strconv.Itoa(field.Length), ErrType: ValueTooLong}
		}

		(*data)[field.Field] = string(bytesVal)

		return &ValidResult{OK: true, ErrType: None}
	}
	// 应该不会有其他类型了
	return &ValidResult{OK: false, Err: "field: " + field.Field + " invalid json string ", ErrType: InvalidFormat}
}

// 针对基本类型（int、float、long、string、bool），转换为对应类型并存入值map中
func extractValueAndValidLength(fieldValue interface{}, field dao.DbpField, data *map[string]interface{}) *ValidResult {
	switch field.Type {
	case "float":
		(*data)[field.Field] = fieldValue.(float64)
	case "int":
		(*data)[field.Field] = int(fieldValue.(float64))
	case "long": // golang 里的long是int64
		(*data)[field.Field] = int64(fieldValue.(float64))
	case "string":
		strVal := fieldValue.(string)
		// 如果字符串长度验证不通过，返回校验结果
		if vr := validStringValueLength(strVal, field); !vr.OK {
			return vr
		}
		(*data)[field.Field] = strVal
	case "bool":
		(*data)[field.Field] = fieldValue.(bool)
	}

	return &ValidResult{OK: true, ErrType: None}
}

// validStringValueLength 验证字段值长度
func validStringValueLength(value string, field dao.DbpField) *ValidResult {
	// 如果字段定义为非空，传空字符串也不可以
	if value == "" && !field.Nullable {
		return &ValidResult{OK: false, Err: "field [" + field.Field + "] can not be null", ErrType: ValueCannotBeNull}
	}

	if utf8.RuneCountInString(value) > field.Length {
		return &ValidResult{OK: false, Err: "field: " + field.Field + " value length large than " + strconv.Itoa(field.Length), ErrType: ValueTooLong}
	}

	return &ValidResult{OK: true}
}

// ParseAndValidLogData valid logger data
func ParseAndValidLogData(data []byte) (bool, error) {
	logger.Logger.Info("parsed origin data: " + string(data))
	jsonParsed, err := gabs.ParseJSON(data)
	if err != nil {
		logger.Logger.Info("gabs parse json has error: " + err.Error())
		kafka.WriteErrorMsg(&ReportError{Err: "parse logger failed", ErrType: ParsedFailed, Data: string(data)})
		return false, errors.New(err.Error())
	}
	validDataMap := make(map[string]interface{})
	ok, err := validEvent(jsonParsed, &validDataMap)
	if !ok {
		kafka.WriteErrorMsg(&ReportError{
			Err:     err.Error(),
			ErrType: EventUndefined,
			Data:    string(data),
		})
		return false, errors.New(err.Error())
	}
	// 查询元数据中所有的字段，依次进行验证
	fields := cache.GetAllFieldLocal()
	for _, field := range *fields {
		validResult := validField(jsonParsed, &validDataMap, field)
		if !validResult.OK { // 字段验证失败
			kafka.WriteErrorMsg(&ReportError{Err: validResult.Err, ErrType: validResult.ErrType, Data: string(data)})
			return false, errors.New(validResult.Err)
		}
	}
	FillReceiveTimeField(&validDataMap)
	// 发送验证后的数据
	kafka.WriteLogMsg(&validDataMap)
	return true, nil
}

// FillReceiveTimeField 填充服务端接收时间
func FillReceiveTimeField(dataMap *map[string]interface{}) {
	(*dataMap)[ReceiveTime] = time.Now().UnixMilli()
}
