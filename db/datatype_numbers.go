package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// 数组数字类型
// 可以使用json字符串存储或者使用逗号分隔符存储
type DataTypeNumbers struct {
	V []int64
}

// IsNil 判断是否是空
func (j *DataTypeNumbers) IsNil() bool {
	return j.V == nil || len(j.V) == 0
}

// Value return json value, implement driver.Valuer interface
func (j DataTypeNumbers) Value() (driver.Value, error) {
	if j.IsNil() {
		return nil, nil
	}
	val := ""
	for _, v := range j.V {
		if val == "" {
			val = fmt.Sprintf("%d", v)
		} else {
			val += "," + fmt.Sprintf("%d", v)
		}
	}
	return val, nil
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *DataTypeNumbers) Scan(value interface{}) error {
	if value == nil {
		*j = DataTypeNumbers{nil}
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

	result := make([]int64, 0)
	if len(bytes) < 1 {
		json.Unmarshal(nil, &result)
	} else {
		rs := strings.Split(string(bytes), ",")
		for _, r := range rs {
			if ri, er := strconv.Atoi(r); er == nil {
				result = append(result, int64(ri))
			}
		}
	}
	*j = DataTypeNumbers{result}
	return nil
}

// MarshalJSON to output non base64 encoded []byte
func (j DataTypeNumbers) MarshalJSON() ([]byte, error) {
	if !j.IsNil() {
		return json.Marshal(j.V)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON to deserialize []byte
func (j *DataTypeNumbers) UnmarshalJSON(b []byte) error {
	values := make([]int64, 0)
	err := json.Unmarshal(b, &values)
	if err != nil {
		return err
	}
	*j = DataTypeNumbers{values}
	return nil
}

// GormDataType gorm common data type
func (j DataTypeNumbers) GormDataType() string {
	return "string"
}
