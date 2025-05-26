package dao

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"liangck.xyz/data-service/sensors-log-acceptor/configer"
	"liangck.xyz/data-service/sensors-log-acceptor/logger"
	"strconv"
	"time"
)

// 数据访问层

// ------------------ Models Definition ----------------------

// DbpEvent 事件
type DbpEvent struct {
	gorm.Model
	ID          uint
	Event       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (e DbpEvent) String() string {
	return fmt.Sprintf("{ID: %d, Event: %s, Description: %s, CreatedAt: %s, UpdatedAt: %s}",
		e.ID, e.Event, e.Description, e.CreatedAt, e.UpdatedAt)
}

// DbpField 字段定义
type DbpField struct {
	gorm.Model
	ID        uint
	Field     string
	JsonPath  string
	Type      string
	Length    int
	Name      string
	Nullable  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (f DbpField) String() string {
	return fmt.Sprintf("{ID: %d, Field: %s, JsonPath: %s, Type: %s, Length: %d, Name: %s, Nullable: %t, CreatedAt: %s, UpdatedAt: %s}",
		f.ID, f.Field, f.JsonPath, f.Type, f.Length, f.Name, f.Nullable, f.CreatedAt, f.UpdatedAt)
}

// DbpEventField 事件字段配置
type DbpEventField struct {
	gorm.Model
	ID        uint
	Event     string
	Field     string
	Nullable  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ef DbpEventField) String() string {
	return fmt.Sprintf("{ID: %d, Event: %s, Field: %s, Nullable: %t, CreatedAt: %s, UpdatedAt: %s}",
		ef.ID, ef.Event, ef.Field, ef.Nullable, ef.CreatedAt, ef.UpdatedAt)
}

// DbpFieldEnumValue 枚举类型属性的枚举值字典表
type DbpFieldEnumValue struct {
	gorm.Model
	ID        uint
	Field     string
	EnumValue string
	ValueName string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (fev DbpFieldEnumValue) String() string {
	return fmt.Sprintf("{IO: %d, Field: %s, EnumValue: %s, ValueName: %s, CreatedAt: %s, UpdatedAt: %s}",
		fev.ID, fev.Field, fev.EnumValue, fev.ValueName, fev.CreatedAt, fev.UpdatedAt)
}

// ----------------------- Database access functions -------------------------
var _db *gorm.DB

// InitDb 初始化_db对象、数据库表
func InitDb(config *configer.Config) {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	dsn := config.DBUrl
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	_db = db
	if _db.Migrator().HasTable(&DbpEvent{}) == false {
		_db.Migrator().CreateTable(&DbpEvent{})
	}
	if _db.Migrator().HasTable(&DbpField{}) == false {
		_db.Migrator().CreateTable(&DbpField{})
	}
	if _db.Migrator().HasTable(&DbpEventField{}) == false {
		_db.Migrator().CreateTable(&DbpEventField{})
	}
	if _db.Migrator().HasTable(&DbpFieldEnumValue{}) == false {
		_db.Migrator().CreateTable(&DbpFieldEnumValue{})
	}
}

// FindAllEvents find all events
// return the pointer of []DbpEvent
func FindAllEvents() *[]DbpEvent {
	var events []DbpEvent
	result := _db.Find(&events)
	if result.Error != nil {
		panic(result.Error)
	}
	logger.Logger.Info("find " + strconv.Itoa(int(result.RowsAffected)) + " rows")
	return &events
}

// FindAllFields find all fields
// return the pointer of []DbpFields
func FindAllFields() *[]DbpField {
	var fields []DbpField
	_db.Find(&fields)
	return &fields
}

// FindAllEventFieldByEvent find all EventFields by Event
// return the pointer of []DbpEventField
func FindAllEventFieldByEvent(event string) *[]DbpEventField {
	var eventFields []DbpEventField
	_db.Where("event = ?", event).Find(&eventFields)
	return &eventFields
}

// FindAllEnumValuesByField find all EnumValues by field
// return the pointer of []DbpFieldEnumValue
func FindAllEnumValuesByField(field string) *[]DbpFieldEnumValue {
	var enumValues []DbpFieldEnumValue
	_db.Where("field = ?", field).Find(&enumValues)
	return &enumValues
}
