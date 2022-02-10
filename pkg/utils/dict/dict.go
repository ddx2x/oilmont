package utils

import (
	"strings"
)

// Set "path":"a.b.c"
// data = {"a":{"b":{"c":123}}}
// Set(data,"a.b.c",123)
func Set(data map[string]interface{}, path string, value interface{}) {
	head, remain := shift(path)
	_, exist := data[head]
	if !exist {
		data[head] = make(map[string]interface{})
	}
	if remain == "" {
		data[head] = value
		return
	}
	Set(data[head].(map[string]interface{}), remain, value)
}

// Get data = {"a":{"b":{"c":123}}}
// Get(data,"a.b.c") = 123
func Get(data map[string]interface{}, path string) (value interface{}) {
	head, remain := shift(path)
	_, exist := data[head]
	if exist {
		if remain == "" {
			return data[head]
		}
		switch data[head].(type) {
		case map[string]interface{}:
			return Get(data[head].(map[string]interface{}), remain)
		case map[interface{}]interface{}:
			_data := make(map[string]interface{})
			for k, v := range data[head].(map[interface{}]interface{}) {
				_data[k.(string)] = v
			}
			return Get(_data, remain)
		}
	}
	return nil
}

// Delete data = {"a":{"b":{"c":123}}}
// Delete(data,"a.b.c") = {"a":{"b":""}}
func Delete(data map[string]interface{}, path string) {
	head, remain := shift(path)
	_, exist := data[head]
	if exist {
		if remain == "" {
			delete(data, head)
			return
		}
		switch data[head].(type) {
		case map[string]interface{}:
			Delete(data[head].(map[string]interface{}), remain)
			return
		}
	}
	return
}

func shift(path string) (head string, remain string) {
	slice := strings.Split(path, ".")
	if len(slice) < 1 {
		return "", ""
	}
	if len(slice) < 2 {
		remain = ""
		head = slice[0]
		return
	}
	return slice[0], strings.Join(slice[1:], ".")
}
