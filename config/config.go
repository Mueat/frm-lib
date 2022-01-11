package config

import (
	"github.com/BurntSushi/toml"
	"github.com/Mueat/frm-lib/errors"
)

// ParseConfig 解析配置文件
// @param string filePath 文件位置
// @param interface{} v 解析的对象
func ParseConfig(filePath string, v interface{}) error {
	_, err := toml.DecodeFile(filePath, v)
	if err != nil {
		return errors.New(err)
	}
	return nil
}
