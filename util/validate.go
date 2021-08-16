package util

import (
	"net"
	"regexp"
)

// 判断是否是手机号码
func IsMobile(mobile string) bool {
	regular := "^(1[3|4|5|6|7|8|9][0-9]\\d{4,8})$"
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobile)

}

// 判断是否是Email
func IsEmail(email string) bool {
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// 是否是网址
func IsURL(url string) bool {
	pattern := `^((http|https):\/\/)?(([A-Za-z0-9]+-[A-Za-z0-9]+|[A-Za-z0-9]+)\.)+([A-Za-z]+)[/\?\:]?.*$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(url)
}

// 是否是QQ
func IsQQ(qq string) bool {
	pattern := `^[1-9][0-9]{4,10}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(qq)
}

// 是否是IPV4
func IsIPV4(ip string) bool {
	matched, err := regexp.MatchString("((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})(\\.((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})){3}", ip)
	if err == nil && matched {
		return true
	}
	return false
}

// 是否是IP地址(包含IPV4和IPV6)
func IsIP(ip string) bool {
	address := net.ParseIP(ip)
	return address != nil
}
