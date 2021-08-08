package util

import "time"

// 获取秒级时间戳
func Time() int64 {
	return time.Now().Unix()
}

// 获取毫秒级时间戳
func MicroTime() int64 {
	return time.Now().UnixNano() / 1e6
}

// Strtotime strtotime()
// Strtotime("02/01/2006 15:04:05", "02/01/2016 15:04:05") == 1451747045
// Strtotime("3 04 PM", "8 41 PM") == -62167144740
func Strtotime(format, strtime string) (int64, error) {
	t, err := time.Parse(format, strtime)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// Date date()
// Date("02/01/2006 15:04:05 PM", 1524799394)
func Date(format string, timestamp int64) string {
	return time.Unix(timestamp, 0).Format(format)
}

// Sleep sleep()
func Sleep(t int64) {
	time.Sleep(time.Duration(t) * time.Second)
}

// Usleep usleep()
func Usleep(t int64) {
	time.Sleep(time.Duration(t) * time.Microsecond)
}
