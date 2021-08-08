package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash"
	"io/ioutil"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Md5 md5()
func Md5(str string) string {
	hash := md5.New()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}

// Md5File md5_file()
func Md5File(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := md5.New()
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Sha1 sha1()
func Sha1(str string) string {
	hash := sha1.New()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}

// Sha1File sha1_file()
func Sha1File(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha1.New()
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// HMAC hmac加密
// @param hash.Hash  h  加密的Hash方法
// @param string d 要加密的字符串
// @param string k 加密密钥
func HMAC(h func() hash.Hash, d string, k string) string {
	hm := hmac.New(h, []byte(k))
	hm.Write([]byte(d))

	//return hex.EncodeToString(hm.Sum(nil))
	return Base64URLEncode(string(hm.Sum(nil)))
}

// EncryptPassword 加密密码
// @param string password 要加密的密码
// @return string 加密后密码
// @return error 错误
func EncryptPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	encryptedPassword := string(hash)
	return encryptedPassword, nil
}

// 验证密码
// @param string encryptedPassword 加密后的密码
// @param string password 明文密码
// @return bool
func DecryptPassword(encryptedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encryptedPassword), []byte(password))
	return err == nil
}

// Base64MapEncode 将map转化为base64
func Base64MapEncode(data map[string]interface{}) string {
	jsonByte, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return Base64URLEncode(string(jsonByte))
}

// Base64MapDecode 将base64
func Base64MapDecode(encryptedString string) (map[string]interface{}, error) {
	b, err := Base64URLDecode(encryptedString)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]interface{})

	decoder := json.NewDecoder(strings.NewReader(string(b)))
	decoder.UseNumber()
	err = decoder.Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// EncryptJWT JWT加密
// @param map[string]interface{} data 要加密的数据
// @param string encryptKey 加密密钥
// @return string 生成的JWT字符串
func EncryptJWT(data map[string]interface{}, encryptKey string) string {
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}
	partOne := Base64MapEncode(header)
	partTwo := Base64MapEncode(data)
	partThree := HMAC(sha256.New, partOne+"."+partTwo, encryptKey)

	return partOne + "." + partTwo + "." + partThree
}

// 校验JWT
// @param string encryptedString 加密的JWT字符串
// @param string encryptKey 加密密钥
// @param interface{} v 解密后的数据
func DecryptJWT(encryptedString string, encryptKey string, v interface{}) error {
	strs := strings.Split(encryptedString, ".")
	if len(strs) != 3 {
		return errors.New("TokenError")
	}

	header, err := Base64MapDecode(strs[0])
	if err != nil {
		return err
	}
	if header["alg"].(string) != "HS256" || header["typ"].(string) != "JWT" {
		return errors.New("TokenHeaderAlgError")
	}

	b, err := Base64URLDecode(strs[1])
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(strings.NewReader(string(b)))
	err = decoder.Decode(v)
	if err != nil {
		return err
	}

	sign := HMAC(sha256.New, strs[0]+"."+strs[1], encryptKey)
	if sign != strs[2] {
		return errors.New("TokenSignError")
	}

	return nil
}

//pkcs7Padding 填充模式
func pkcs7Padding(cipherText []byte, blockSize int) []byte {
	//取余计算长度,判断加密的文本是不是blockSize的倍数,如果不是的话把多余的长度计算出来,用于补齐长度
	padding := blockSize - len(cipherText)%blockSize
	//补齐
	//Repeat: 把切片[]byte{byte(padding)}复制padding个然后合并成新的字节切片返回
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	rt := append(cipherText, padText...)
	return rt
}

//实现加密
func AesEncrypt(originData []byte, key []byte, iv []byte) ([]byte, error) {
	//创建加密算法的实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//获取块的大小
	blockSize := block.BlockSize()

	//对数据进行填充,让数据的长度满足加密需求
	originData = pkcs7Padding(originData, blockSize)

	//采用aes加密方式中的CBC加密模式
	blockMode := cipher.NewCBCEncrypter(block, iv)
	crypted := make([]byte, len(originData))

	//执行加密
	blockMode.CryptBlocks(crypted, originData)

	//返回
	return crypted, nil
}

//将加密的结果进行base64编码
func EnAesCode2Base64(pwd []byte, secret []byte, iv []byte) (string, error) {
	//进行aes加密
	result, err := AesEncrypt(pwd, secret, iv)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(result), err
}

//填充的反向操作,删除填充的字符串
func pkcs7UnPadding(originData []byte) ([]byte, error) {
	//获取数据长度
	length := len(originData)
	if length <= 0 {
		return nil, errors.New("加密字符串长度不符合要求")
	}
	//获取填充字符串的长度
	unPadding := int(originData[length-1])
	//截取切片,删除填充的字节,并且返回明文
	return originData[:(length - unPadding)], nil
}

//实现解密
func AesDeCrypt(cypted []byte, key []byte, iv []byte) ([]byte, error) {
	//创建加密算法的实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//创建加密实例
	blockMode := cipher.NewCBCDecrypter(block, iv)

	originData := make([]byte, len(cypted))

	//该函数可也用来加密也可也用来解密
	blockMode.CryptBlocks(originData, cypted)
	//取出填充的字符串
	originData, err = pkcs7UnPadding(originData)
	if err != nil {
		return nil, err
	}
	return originData, nil
}

//base64解码
func DeAesCode2Base64(cyptedStr string, key []byte, iv []byte) ([]byte, error) {
	//解码base64字符串
	cyptedByte, err := base64.StdEncoding.DecodeString(cyptedStr)
	if err != nil {
		return nil, err
	}
	//执行aes解密
	return AesDeCrypt(cyptedByte, key, iv)
}
