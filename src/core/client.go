package core

import (
	"bytes"
	"cx-im/src/config"
	"cx-im/src/model"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
)

var (
	regexpCourses = regexp.MustCompile(`<a href="https://mooc1\.chaoxing\.com/visit/stucoursemiddle\?courseid=(\d+?)&clazzid=(\d+)&cpi=\d+["&]`)
	regexpImToken = regexp.MustCompile(`loginByToken\('(\d+?)', '([^']+?)'\);`)
)

type CxClient struct {
	Username string
	Password string
	Fid      string
	Uid      string
	Logged   bool
	Client   *http.Client
	Log      *LogE
}

func NewClient(username, password, fid string, logE *LogE) (*CxClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	if logE == nil {
		logE = &LogE{
			Logger: log.Default(),
		}
	}
	return &CxClient{
		username,
		password,
		fid,
		"",
		false,
		&http.Client{
			Transport: HttpTransport,
			Jar:       jar,
		},
		logE,
	}, nil
}

func NewClientFromConfig(cfg *config.Config, logE *LogE) (*CxClient, error) {
	username := config.GodCI(cfg, config.Username, "")
	if username == "" {
		return nil, errors.New("账号不存在")
	}
	password := config.GodCI(cfg, config.Password, "")
	if password == "" {
		return nil, errors.New("密码不存在")
	}
	fid := config.GodCI(cfg, config.Fid, "")
	return NewClient(username, password, fid, logE)
}

func (c *CxClient) Login() error {
	log.Printf("正在登录账号：%s", c.Username)
	var req *http.Request
	var err error
	if c.Fid == "" {
		query := url.Values{
			"uname": {c.Username},
			"code":  {c.Password},
		}
		req, err = http.NewRequest(
			"GET",
			"https://passport2-api.chaoxing.com/v11/loginregister?"+query.Encode(),
			nil,
		)
	} else {
		// TODO login by fid
	}
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer c.Log.CloseResponse(resp)
	err = testResponseStatus(resp)
	if err != nil {
		return err
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Domain == ".chaoxing.com" {
			switch cookie.Name {
			case "fid":
				c.Fid = cookie.Value
			case "_uid":
				c.Uid = cookie.Value
			}
		}
	}
	c.Logged = true
	log.Println("成功登录账号")
	return nil
}

func (c *CxClient) GetCourses(courses *config.Object) error {
	log.Println("获取课程数据汇总……")
	resp, err := c.Client.Get("https://mooc2-ans.chaoxing.com/visit/courses/list?rss=1&catalogId=0&searchname=")
	if err != nil {
		return err
	}
	defer c.Log.CloseResponse(resp)
	err = testResponseStatus(resp)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	matches := regexpCourses.FindAllStringSubmatch(string(data), -1)
	for _, match := range matches {
		courseId := match[1]
		classId := match[2]
		c.Log.ErrPrint(c.GetCourseDetail(courses, courseId, classId))
	}
	return nil
}

func (c *CxClient) GetCourseDetail(courses *config.Object, courseId string, classId string) error {
	query := url.Values{
		"fid":      {c.Fid},
		"courseId": {courseId},
		"classId":  {classId},
	}
	resp, err := c.Client.Get("https://mobilelearn.chaoxing.com/v2/apis/class/getClassDetail?" + query.Encode())
	if err != nil {
		return err
	}
	defer c.Log.CloseResponse(resp)
	err = testResponseStatus(resp)
	if err != nil {
		return err
	}
	data, err := parseCxClientJson(resp)
	if err != nil {
		return err
	}
	chatId := AnyToString(data["chatid"])
	courseName := AnyToString(AnyToJObject(AnyToJArray(AnyToJObject(data["course"])["data"]).Get(0))["name"])
	className := AnyToString(data["name"])
	c.Log.Printf("发现课程：《%s》『%s』(%s, %s) %s", courseName, className, courseId, classId, chatId)

	course := config.GocObjI(courses, chatId)
	course.Set(config.ChatId, chatId)
	course.Set(config.CourseId, courseId)
	course.Set(config.ClassId, classId)
	course.Set(config.CourseName, courseName)
	course.Set(config.ClassName, className)
	return nil
}

func (c *CxClient) GetImToken() (string, string, error) {
	resp, err := c.Client.Get("https://im.chaoxing.com/webim/me")
	if err != nil {
		return "", "", err
	}
	defer c.Log.CloseResponse(resp)
	err = testResponseStatus(resp)
	if err != nil {
		return "", "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	match := regexpImToken.FindStringSubmatch(string(data))
	if match == nil {
		return "", "", errors.New("没有匹配 regexpImToken")
	}
	return match[1], match[2], nil
}

func (c *CxClient) GetActiveDetail(activeId string) (JObject, error) {
	query := url.Values{
		"activeId": []string{activeId},
	}
	resp, err := c.Client.Get("https://mobilelearn.chaoxing.com/v2/apis/active/getPPTActiveInfo?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer c.Log.CloseResponse(resp)
	err = testResponseStatus(resp)
	if err != nil {
		return nil, err
	}
	return parseCxClientJson(resp)
}

func (c *CxClient) PreSign(activeId string) error {
	query := url.Values{
		"activePrimaryId": []string{activeId},
	}
	resp, err := c.Client.Get("https://mobilelearn.chaoxing.com/newsign/preSign?" + query.Encode())
	if err != nil {
		return err
	}
	defer c.Log.CloseResponse(resp)
	return testResponseStatus(resp)
}

func (c *CxClient) Sign(activeId string, signOptions *model.SignOptions) (string, error) {
	query := url.Values{
		"activeId":  []string{activeId},
		"appType":   []string{"15"},
		"ifTiJiao":  []string{"1"},
		"address":   []string{signOptions.Address},
		"longitude": []string{signOptions.Longitude},
		"latitude":  []string{signOptions.Latitude},
		"clientip":  []string{signOptions.Ip},
		"objectId":  []string{signOptions.ImageId},
	}
	resp, err := c.Client.Get("https://mobilelearn.chaoxing.com/pptSign/stuSignajax?" + query.Encode())
	if err != nil {
		return "", err
	}
	defer c.Log.CloseResponse(resp)
	err = testResponseStatus(resp)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *CxClient) GetImageHostingToken() (string, error) {
	resp, err := c.Client.Get("https://pan-yz.chaoxing.com/api/token/uservalid")
	defer c.Log.CloseResponse(resp)
	err = testResponseStatus(resp)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var v JsonImageHostingToken
	err = json.Unmarshal(data, &v)
	if err != nil {
		return "", err
	}
	if !v.Result {
		return "", errors.New("获取 Token 失败")
	}
	return v.Token, nil
}

func (c *CxClient) buildUploadImageBody(filename string) (string, io.Reader, error) {
	token, err := c.GetImageHostingToken()
	if err != nil {
		return "", nil, err
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err != nil {
		return "", nil, err
	}
	defer c.Log.ErrClose(writer)
	err = writer.WriteField("puid", c.Uid)
	if err != nil {
		return "", nil, err
	}
	err = writer.WriteField("_token", token)
	if err != nil {
		return "", nil, err
	}
	fw, err := writer.CreateFormFile("file", "image"+path.Ext(filename))
	if err != nil {
		return "", nil, err
	}
	file, err := os.Open(filename)
	if err != nil {
		return "", nil, err
	}
	defer c.Log.ErrClose(file)
	_, err = io.Copy(fw, file)
	if err != nil {
		return "", nil, err
	}
	return writer.FormDataContentType(), body, nil
}

func (c *CxClient) UploadImage(filename string) (string, error) {
	contentType, body, err := c.buildUploadImageBody(filename)
	if err != nil {
		return "", err
	}
	resp, err := c.Client.Post("https://pan-yz.chaoxing.com/upload", contentType, body)
	if err != nil {
		return "", err
	}
	defer c.Log.CloseResponse(resp)
	err = testResponseStatus(resp)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var v JsonUpload
	err = json.Unmarshal(data, &v)
	if err != nil {
		return "", err
	}
	if !v.Result {
		return "", errors.New(v.Msg)
	}
	return v.ObjectId, nil
}

func testResponseStatus(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("响应状态非 200 OK： %s", resp.Status)
}

type JsonCxClient struct {
	Result   int     `json:"result"`
	Msg      string  `json:"msg"`
	Data     JObject `json:"data"`
	ErrorMsg *string `json:"errorMsg"`
}

func parseCxClientJson(resp *http.Response) (JObject, error) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var v JsonCxClient
	err = json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}
	if v.Result == 1 && v.Data != nil {
		return v.Data, nil
	}
	msg := "parseCxClientJson\n" + strconv.Itoa(v.Result) + ": " + v.Msg
	if v.ErrorMsg != nil {
		msg += "\n" + *v.ErrorMsg
	}
	return nil, errors.New(msg)
}

var HttpTransport = (func() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// 抓包调试
	//proxy, err := url.Parse("http://127.0.0.1:8888")
	//if !errs.Print(err) {
	//	transport.Proxy = http.ProxyURL(proxy)
	//}
	return transport
})()

var HttpClient = &http.Client{
	Transport: HttpTransport,
}

const MimeJson = "application/json"
