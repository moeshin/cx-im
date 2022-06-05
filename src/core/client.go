package core

import (
	"bytes"
	"crypto/sha256"
	"cx-im/src/config"
	"cx-im/src/model"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/orirawlings/persistent-cookiejar"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	regexpCourses = regexp.MustCompile(`<a href="https://mooc1\.chaoxing\.com/visit/stucoursemiddle\?courseid=(\d+?)&clazzid=(\d+)&cpi=\d+["&]`)
	regexpImToken = regexp.MustCompile(`loginByToken\('(\d+?)', '([^']+?)'\);`)
)

var ClientNoCache bool

type CxClient struct {
	User   *User
	Fid    string
	Uid    string
	Logged bool
	Jar    *cookiejar.Jar
	Client *http.Client
}

func NewClient(user *User) (*CxClient, error) {
	filename := user.Dir.GetCookiesPath()
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename:  filename,
		NoPersist: ClientNoCache,
	})
	if err != nil {
		return nil, err
	}
	client := &CxClient{
		User: user,
		Jar:  jar,
		Client: &http.Client{
			Transport: HttpTransport,
			Jar:       jar,
		},
	}
	if !ClientNoCache && CanFileStat(filename) {
		user.Log.Println("加载 Cookie 缓存：" + filename)
		client.Logged = true
	} else {
		err = client.Login()
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

func (c *CxClient) GetConfigString(key string) string {
	return config.GodCI(c.User.Config, key, "")
}

func (c *CxClient) GetUsername() string {
	return c.GetConfigString(config.Username)
}

func (c *CxClient) GetPassword() string {
	return c.GetConfigString(config.Password)
}

func (c *CxClient) GetFid() string {
	return c.GetConfigString(config.Fid)
}

func (c *CxClient) Login() error {
	username := c.GetUsername()
	if username == "" {
		return errors.New("用户名为空")
	}
	password := c.GetPassword()
	if password == "" {
		return errors.New("密码为空")
	}
	fid := c.GetFid()
	c.User.Log.Printf("正在登录账号：%s", username)
	var req *http.Request
	var err error
	if fid == "" {
		query := url.Values{
			"uname": {username},
			"code":  {password},
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
	c.Jar.RemoveAll()
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer c.User.Log.CloseResponse(resp)
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
	c.User.Log.Println("成功登录账号")
	c.User.Log.Println("保存 Cookie 缓存")
	c.User.Log.ErrPrint(c.Jar.Save())
	return c.User.SaveImageToken(c)
}

func (c *CxClient) Auth() error {
	if c.Logged {
		return nil
	}
	return c.Login()
}

func (c *CxClient) GetCourses(courses *config.Object) error {
	c.User.Log.Println("获取课程数据汇总……")
	resp, err := c.Client.Get("https://mooc2-ans.chaoxing.com/visit/courses/list?rss=1&catalogId=0&searchname=")
	if err != nil {
		return err
	}
	defer c.User.Log.CloseResponse(resp)
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
		c.User.Log.ErrPrint(c.GetCourseDetail(courses, courseId, classId))
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
	defer c.User.Log.CloseResponse(resp)
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
	c.User.Log.Printf("发现课程：《%s》『%s』(%s, %s) %s", courseName, className, courseId, classId, chatId)

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
	defer c.User.Log.CloseResponse(resp)
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
	defer c.User.Log.CloseResponse(resp)
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
	defer c.User.Log.CloseResponse(resp)
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
	defer c.User.Log.CloseResponse(resp)
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

func (c *CxClient) GetImageToken() (string, error) {
	resp, err := c.Client.Get("https://pan-yz.chaoxing.com/api/token/uservalid")
	defer c.User.Log.CloseResponse(resp)
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

func (c *CxClient) buildUploadImageBody(filename string, file io.ReadCloser) (string, io.Reader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer c.User.Log.ErrClose(writer)
	err := writer.WriteField("puid", c.Uid)
	if err != nil {
		return "", nil, err
	}
	err = writer.WriteField("_token", c.User.ImageToken)
	if err != nil {
		return "", nil, err
	}
	fw, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", nil, err
	}
	if file == nil {
		file, err = os.Open(filename)
		if err != nil {
			return "", nil, err
		}
		defer c.User.Log.ErrClose(file)
	}
	_, err = io.Copy(fw, file)
	if err != nil {
		return "", nil, err
	}
	return writer.FormDataContentType(), body, nil
}

func (c *CxClient) UploadImage(filename string, file io.ReadCloser) (string, error) {
	ext := filepath.Ext(filename)
	if ext == "" || !strings.HasPrefix(ext, ".") {
		return "", errors.New("图片扩展名错误：" + filename)
	}
	ext = ext[1:]
	_, ok := config.ImageExt[ext[1:]]
	if !ok {
		return "", errors.New("不支持该图片扩展名：" + ext)
	}
	contentType, body, err := c.buildUploadImageBody(filename, file)
	if err != nil {
		return "", err
	}
	resp, err := c.Client.Post("https://pan-yz.chaoxing.com/upload", contentType, body)
	if err != nil {
		return "", err
	}
	defer c.User.Log.CloseResponse(resp)
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

func (c *CxClient) GetImageId(filename string, file io.ReadSeekCloser, size int64) (string, error) {
	CacheImage.Mutex.Lock()
	defer CacheImage.Mutex.Unlock()
	hasFile := file != nil
	if !hasFile {
		i, err := os.Stat(filename)
		if err != nil {
			return "", err
		}
		file, err = os.Open(filename)
		if err != nil {
			return "", err
		}
		defer c.User.Log.ErrClose(file)
		size = i.Size()
	}
	h := sha256.New()
	_, err := io.Copy(h, file)
	if err != nil {
		return "", err
	}
	key := hex.EncodeToString(h.Sum(nil)) + strconv.FormatInt(size, 10)
	v, ok := CacheImage.Map[key]
	if ok {
		return v, nil
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}
	v, err = c.UploadImage(filename, file)
	if err != nil {
		return "", err
	}
	CacheImage.Map[key] = v
	CacheImage.Save = true
	if hasFile {
		saveCacheImage()
	}
	return v, nil
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

var HttpTransport = func() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// 抓包调试
	//proxy, err := url.Parse("http://127.0.0.1:8888")
	//if !errs.Print(err) {
	//	transport.Proxy = http.ProxyURL(proxy)
	//}
	return transport
}()

var HttpClient = &http.Client{
	Transport: HttpTransport,
}

const MimeJson = "application/json"
