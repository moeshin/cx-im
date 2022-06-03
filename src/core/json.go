package core

type JsonImageHostingToken struct {
	Result bool   `json:"result"`
	Token  string `json:"_token"`
}

type JsonUpload struct {
	Result   bool   `json:"result"`
	Msg      string `json:"msg"`
	ObjectId string `json:"objectId"`
}
