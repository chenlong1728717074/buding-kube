package dto

import (
	"fmt"
	"time"
)

// TimeFormat 定义日期时间格式常量
const TimeFormat = "2006-01-02 15:04:05"

// CustomTime 自定义时间类型
type CustomTime time.Time

// MarshalJSON 实现JSON序列化接口
func (t CustomTime) MarshalJSON() ([]byte, error) {
	stamp := time.Time(t).Format(TimeFormat)
	return []byte(fmt.Sprintf("\"%s\"", stamp)), nil
}

// UnmarshalJSON 实现JSON反序列化接口
func (t *CustomTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	// 去掉JSON字符串的引号
	str := string(data)
	if len(str) >= 2 {
		str = str[1 : len(str)-1]
	}

	// 解析自定义格式的时间
	parsedTime, err := time.Parse(TimeFormat, str)
	if err != nil {
		return err
	}

	*t = CustomTime(parsedTime)
	return nil
}

// 实现String方法
func (t CustomTime) String() string {
	return time.Time(t).Format(TimeFormat)
}

func (t CustomTime) ToTime() time.Time {
	return time.Time(t)
}
