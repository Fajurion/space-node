package util

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
)

var Port int = 0
var UDPPort int = 0
var Log = log.New(log.Writer(), "space-node ", log.Flags())

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GenerateToken(tkLength int32) string {

	s := make([]rune, tkLength)

	length := big.NewInt(int64(len(letters)))

	for i := range s {

		number, _ := rand.Int(rand.Reader, length)
		s[i] = letters[number.Int64()]
	}

	return string(s)
}

var Protocol = "http://"
var BasePath = "http://localhost:3000"

func PostRequest(url string, body map[string]interface{}) (map[string]interface{}, error) {
	return PostRaw(BasePath+url, body)
}

func PostRaw(url string, body map[string]interface{}) (map[string]interface{}, error) {

	req, _ := sonic.Marshal(body)

	reader := strings.NewReader(string(req))

	res, err := http.Post(url, "application/json", reader)
	if err != nil {
		return nil, err
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, res.Body)

	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = sonic.Unmarshal([]byte(buf.String()), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Node64(id int64) string {
	return fmt.Sprintf("%d", id)
}

func NodeTo64(id string) int64 {
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0
	}

	return i
}
