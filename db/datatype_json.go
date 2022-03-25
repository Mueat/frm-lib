package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// json类型数据
type DataTypeJson struct {
	json.RawMessage
}

// IsNil 判断是否是空
func (j *DataTypeJson) IsEmpty() bool {
	return j.RawMessage == nil || len(j.RawMessage) == 0
}

// Value return json value, implement driver.Valuer interface
func (j DataTypeJson) Value() (driver.Value, error) {
	if !j.IsEmpty() {
		return nil, nil
	}
	bytes, err := json.RawMessage(j.RawMessage).MarshalJSON()
	return string(bytes), err
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *DataTypeJson) Scan(value interface{}) error {
	if value == nil {
		*j = DataTypeJson{[]byte{}}
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := json.RawMessage{}
	var err error
	if len(bytes) < 1 {
		json.Unmarshal(nil, &result)
	} else {
		err = json.Unmarshal(bytes, &result)
	}
	*j = DataTypeJson{result}
	return err
}

// MarshalJSON to output non base64 encoded []byte
func (j DataTypeJson) MarshalJSON() ([]byte, error) {
	if !j.IsEmpty() {
		return json.RawMessage(j.RawMessage).MarshalJSON()
	}
	return json.Marshal(nil)
}

// UnmarshalJSON to deserialize []byte
func (j *DataTypeJson) UnmarshalJSON(b []byte) error {
	result := json.RawMessage{}
	err := result.UnmarshalJSON(b)
	*j = DataTypeJson{result}
	return err
}

func (j DataTypeJson) String() string {
	return string(j.RawMessage)
}

// GormDataType gorm common data type
func (j DataTypeJson) GormDataType() string {
	return "string"
}

// 将Json转换为其他struct
func (j *DataTypeJson) UnmarshalTo(v interface{}) error {
	msg := json.RawMessage(j.RawMessage)
	return json.Unmarshal(msg, v)
}
