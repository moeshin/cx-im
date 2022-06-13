package core

import (
	"encoding/json"
)

type JsonImageHostingToken struct {
	Result bool   `json:"result"`
	Token  string `json:"_token"`
}

type JsonUpload struct {
	Result   bool   `json:"result"`
	Msg      string `json:"msg"`
	ObjectId string `json:"objectId"`
}

type JsonLogin struct {
	Status bool   `json:"status"`
	Msg    string `json:"mes"`
	//Url    string `json:"url"`
	//Type   int    `json:"type"`
}

//func JsonUnmarshal(filename string, v any) error {
//	file, err := os.Open(filename)
//	if err != nil {
//		return err
//	}
//	defer errs.Close(file)
//	data, err := ioutil.ReadAll(file)
//	if err != nil {
//		return err
//	}
//	return json.Unmarshal(data, v)
//}

func JsonMarshal(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
