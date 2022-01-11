package util

import (
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/go-ldap/ldap/v3"
)

/*
ldap配置
*/
type LdapConfig struct {
	Host       string         //ip或者主机名
	Port       string         //端口
	BaseDn     string         //基础DN
	Attributes LdapAttributes //结果集
}
type LdapAttributes struct {
	UNameKey string //ldap中用户名的key
	NameKey  string //ldap中姓名的key
	EmailKey string //ldap中email的key
}

// LdapUser LDAP用户
type LdapUser struct {
	UName string
	Name  string
	Email string
}

// LdapLogin 使用LDAP登录
func LdapLogin(config *LdapConfig, username string, password string) (*LdapUser, error) {
	var (
		filter     string
		attributes []string
		conn       *ldap.Conn
		err        error
		cur        *ldap.SearchResult
	)

	filter = fmt.Sprintf("(%s=%s)", config.Attributes.UNameKey, username)
	attributes = []string{config.Attributes.UNameKey, config.Attributes.NameKey, config.Attributes.EmailKey}
	if config.Port == "636" {
		conn, err = ldap.DialTLS("tcp", config.Host+":"+config.Port, &tls.Config{InsecureSkipVerify: true})
	} else {
		conn, err = ldap.Dial("tcp", config.Host+":"+config.Port)
	}

	if err != nil {
		return nil, err
	}
	conn.SetTimeout(5 * time.Second)
	defer conn.Close()
	dn := fmt.Sprintf(config.BaseDn, username)
	sql := ldap.NewSearchRequest(
		dn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		attributes,
		nil)

	if cur, err = conn.Search(sql); err != nil {
		return nil, err
	}
	if len(cur.Entries) == 0 {
		return nil, errors.New("UserNotFound")
	}
	entry := cur.Entries[0]
	user := LdapUser{
		Name:  entry.GetAttributeValue(config.Attributes.NameKey),
		UName: entry.GetAttributeValue(config.Attributes.UNameKey),
		Email: entry.GetAttributeValue(config.Attributes.EmailKey),
	}
	err = conn.Bind(entry.DN, password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
