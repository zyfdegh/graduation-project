package common

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/compose/mejson"
	"gopkg.in/gomail.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	COMMON_ERROR_INVALIDATE   = "E12002"
	COMMON_ERROR_UNAUTHORIZED = "E12004"
	COMMON_ERROR_UNKNOWN      = "E12001"
	COMMON_ERROR_INTERNAL     = "E12003"
)

type UserParam struct {
	Email           string
	Password        string
	ConfirmPassword string
	Alias           string
	Address         string
	Company         string
	Phonenum        string
	InfoSource      string
}

func ConvertToBson(object interface{}) (document bson.M) {
	b, _ := json.Marshal(object)
	reader := strings.NewReader(string(b))
	decoder := json.NewDecoder(reader)
	decoder.Decode(&document)
	document, _ = mejson.Unmarshal(document)
	return document
}

func GetCurrentTime() (t string) {
	t = time.Now().Format(time.RFC3339)
	return
}

func IsFirstNodeInZK() bool {
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Warnln("get host name error!", err)
		return false
	}

	path, err := UTIL.ZkClient.GetFirstUserMgmtPath()
	if err != nil {
		logrus.Warnln("get usermgmt node from zookeeper error!", err)
		return false
	}

	return strings.HasPrefix(path, hostname)

}

func HashString(password string) string {
	encry := md5.Sum([]byte(password))
	return hex.EncodeToString(encry[:])
}

func GenerateActiveCode() string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	strlen := 10
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func SendMail(host string, username string, passwd string, to string, subject string, body string) {
	m := gomail.NewMessage()
	m.SetHeader("From", username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewPlainDialer(host, 25, username, passwd)

	err := d.DialAndSend(m)
	if err != nil {
		logrus.Warnln("send active user email error %v", err)
	}

	return
}

func DesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	origData = pkcs5Padding(origData, block.BlockSize())

	blockMode := cipher.NewCBCEncrypter(block, key)
	crypted := make([]byte, len(origData))
	if len(origData)%blockMode.BlockSize() != 0 {
		return nil, errors.New("failed to encrypt due to invalid encrypt message")
	}
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func DesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, key)
	origData := make([]byte, len(crypted))
	if len(crypted)%blockMode.BlockSize() != 0 {
		return nil, errors.New("failed to decrypt due to invalid decrypt message")
	}
	blockMode.CryptBlocks(origData, crypted)
	origData = pkcs5UnPadding(origData)
	return origData, nil
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func GetWaitTime(execTime time.Time) int64 {
	one_day := 24 * 60 * 60
	currenTime := time.Now()

	execInt := execTime.Unix()
	currentInt := currenTime.Unix()

	var waitTime int64
	if currentInt <= execInt {
		waitTime = execInt - currentInt
	} else {
		waitTime = (execInt + int64(one_day)) % currentInt
	}

	return waitTime
}

//default expire time is 6 hours
func GenerateExpireTime(expire int64) float64 {
	t := time.Now().Unix()

	t += expire

	return float64(t)
}
