package util

import (
	"crypto/rand"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
)

var Log = log.New(log.Writer(), "voice-node ", log.Flags())

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
