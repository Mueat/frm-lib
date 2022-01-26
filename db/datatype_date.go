package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// 日期类型
type DataTypeDate struct {
	*time.Time
}

// IsNil 判断是否是空
func (j *DataTypeDate) IsNil() bool {
	return j.Time == nil
}

func (j DataTypeDate) Value() (driver.Value, error) {
	if j.IsNil() {
		return nil, nil
	}
	return *j.Time, nil
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *DataTypeDate) Scan(v interface{}) error {
	if v == nil {
		*j = DataTypeDate{nil}
		return nil
	}
	if vt, ok := v.(time.Time); ok {
		*j = DataTypeDate{&vt}
		return nil
	}
	v1, ok := v.([]byte)
	if ok {
		var data time.Time
		err := json.Unmarshal(v1, &data)
		if err != nil {
			return err
		}
		*j = DataTypeDate{&data}
		return nil
	}
	return fmt.Errorf("can not convert %v to Date", v)

}

// MarshalJSON 转换成json
func (j DataTypeDate) MarshalJSON() ([]byte, error) {
	if !j.IsNil() {
		const timeFormat = "2006-01-02"
		if !j.Time.IsZero() {
			b := fmt.Sprintf("\"%s\"", j.Time.Format(timeFormat))
			return []byte(b), nil
		}
	}
	return json.Marshal(nil)
}

// UnmarshalJSON JSON字串转为String
func (j *DataTypeDate) UnmarshalJSON(bytes []byte) error {
	err := (*time.Time)(j.Time).UnmarshalJSON(bytes)
	if err != nil {
		return err
	}
	return nil
}

// GormDataType gorm common data type
func (j DataTypeDate) GormDataType() string {
	return "date"
}
