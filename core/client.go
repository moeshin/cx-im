package core

import (
	"cx-im/config"
	"encoding/json"
	"errors"
	"github.com/moeshin/go-errs"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
)

type CxClient struct {
	Username string
	Password string
	Fid      string
	Uid      string
	Logged   bool
	Client   *http.Client
}

var (
	regexpCourses,
	regexpImToken,
	_ *regexp.Regexp
)

func init() {
	var err error
	regexpCourses, err = regexp.Compile(`<a href="https://mooc1\.chaoxing\.com/visit/stucoursemiddle\?courseid=(\d+?)&clazzid=(\d+)&cpi=\d+["&]`)
	regexpImToken, err = regexp.Compile(`loginByToken\('(\d+?)', '([^']+?)'\);`)
	errs.Panic(err)
}

func NewClient(username, password, fid string) (*CxClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &CxClient{
		username,
		password,
		fid,
		"",
		false,
		&http.Client{
			Jar: jar,
		},
	}, nil
}

func NewClientFromConfig(cfg *config.Config) (*CxClient, error) {
	var username, password, fid string
	v, ok := cfg.Data.Get(config.Username)
	if ok {
		username, ok = v.(string)
	}
	if username == "" {
		return nil, errors.New("账号不存在")
	}
	v, ok = cfg.Data.Get(config.Password)
	if ok {
		password, ok = v.(string)
	}
	if password == "" {
		return nil, errors.New("密码不存在")
	}
	v, _ = cfg.Data.Get(config.Fid)
	if ok {
		fid, ok = v.(string)
	}
	return NewClient(username, password, fid)
}

func (c *CxClient) Login() error {
	log.Printf("正在登录账号：%s\n", c.Username)
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
	defer errs.CloseResponse(resp)
	err = testCxClientStatus(resp)
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
	defer errs.CloseResponse(resp)
	err = testCxClientStatus(resp)
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
		errs.Print(c.GetCourseDetail(courses, courseId, classId))
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
	defer errs.CloseResponse(resp)
	err = testCxClientStatus(resp)
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
	log.Printf("发现课程：《%s》『%s』(%s, %s) %s\n", courseName, className, courseId, classId, chatId)

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
	defer errs.CloseResponse(resp)
	err = testCxClientStatus(resp)
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
	defer errs.CloseResponse(resp)
	err = testCxClientStatus(resp)
	if err != nil {
		return nil, err
	}
	return parseCxClientJson(resp)
}

func testCxClientStatus(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return errors.New("响应状态非 200 OK")
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
