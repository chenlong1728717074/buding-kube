package utils

import "encoding/json"

func Struct2Map(s interface{}) map[string]string {
	j, _ := json.Marshal(s)
	m := make(map[string]string)
	json.Unmarshal(j, &m)
	return m
}

func Map2Struct(m map[string]string, s interface{}) error {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonBytes, s)
	if err != nil {
		return err
	}
	return nil
}
