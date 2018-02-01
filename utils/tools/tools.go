package tools

import (
	"encoding/json"
	"strconv"
)

func JsonStrToMap(jsonStr string) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return nil, err
	}
	return m, nil
}

func IntToString(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}

func Int64ToString(n int64) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}


func StringToInt(s string) (int32, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return -1, err
	}
	return int32(i), nil
}

func StringParseInt(s string) int32 {
	i, _ := strconv.ParseInt(s, 10, 32)
	return int32(i)
}

func StringParseInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return int64(i)
}

//func MergeSlice(s1 []string, s2 []string) []string {
//	slice := make([]string, len(s1)+len(s2))
//	copy(slice, s1)
//	copy(slice[len(s1):], s2)
//	return slice
//}

func MergeSlice(s1 []string, s2 []string) []string {
	s3 := append(s1,s2...)
	return s3
}