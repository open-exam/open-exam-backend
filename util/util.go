package util

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"strings"
)

func GenerateRandomBytes(size uint32) []byte {
	token := make([]byte, size)
	rand.Read(token)
	return token
}

func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func DecodeBase64(message []byte) ([]byte, error) {
	base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(message)))

	_, err := base64.StdEncoding.Decode(base64Text, message)
	if err != nil {
		return nil, err
	}
	return base64Text, nil
}

func EncodeBase64(message []byte) []byte {
	base64Text := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(base64Text, message)
	return base64Text
}

func IsInList(item string, list *[]string) int {
	for i, val := range *list {
		if val == item {
			return i
		}
	}
	return -1
}

func GetSHA256(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func SplitAndParse(data string) []string {
	return Map(strings.Split(data, ","), func(item string) string {
		return strings.TrimSpace(item)
	})
}

func ReadCSV(buf []byte) ([][]string, error){
	r := csv.NewReader(bytes.NewReader(buf))
	return r.ReadAll()
}