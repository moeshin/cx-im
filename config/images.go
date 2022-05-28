package config

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const ImageDir = "Images"
const TimeLayout = "2006-01-02 15:04:05"

var ImageExt = (func() map[string]bool {
	set := map[string]bool{}
	for _, v := range []string{"png", "jpg", "jpeg", "bmp", "gif", "webp"} {
		set[v] = true
	}
	return set
})()

type Images map[string]bool

func (i Images) addFile(path string) {
	ext := filepath.Ext(path)
	if strings.HasPrefix(ext, ".") {
		_, ok := ImageExt[ext[1:]]
		if ok {
			i[path] = true
		}
	}
}

func (i Images) addDir(path string) {
	dirs, err := os.ReadDir(path)
	if err != nil {
		return
	}
	for _, dir := range dirs {
		path := filepath.Join(path, dir.Name())
		if dir.IsDir() {
			i.addDir(path)
		} else {
			i.addFile(path)
		}
	}
}

func (i Images) addPath(path string) {
	if path == "" {
		i[""] = true
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.IsDir() {
		i.addDir(path)
	} else {
		i.addFile(path)
	}
}

func (i Images) addAny(v any) {
	path, ok := v.(string)
	if !ok {
		return
	}
	i.addPath(getImageFullPath(path))
}

func (c *Config) GetImages(time time.Time) []string {
	var images Images
	v := c.GetR(SignPhoto)
	add := func(v any) {
		switch arr := v.(type) {
		case []map[string]any:
			for _, v := range arr {
				images.addAny(v)
			}
		default:
			images.addAny(v)
		}
	}
	switch obj := v.(type) {
	case map[string]any:
		for k, v := range obj {
			if MatchSignPhotoKey(time, k) {
				images.addAny(v)
			}
		}
	default:
		add(v)
	}
	size := len(images)
	if size == 0 {
		return nil
	}
	arr := make([]string, size)
	i := 0
	for p := range images {
		arr[i] = p
		i++
	}
	return arr
}

func getImageFullPath(path string) string {
	if path != "" && !filepath.IsAbs(path) {
		path = filepath.Join(ImageDir, path)
		abs, err := filepath.Abs(path)
		if err == nil {
			path = abs
		}
	}
	return path
}

const (
	TimeHour = 60 * 60
	TimeNoon = 12 * TimeHour
	TimeDay  = 2 * TimeNoon
)

var regexpTime = regexp.MustCompile(`^(\d+):(\d+)$`)

func matchSignPhotoErr(rule string) {
	log.Println("拍照签到参数错误：", rule)
}

func parseSignPhotoTime(tm string) int64 {
	match := regexpTime.FindStringSubmatch(tm)
	if match == nil {
		return -1
	}
	hour, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return -1
	}
	minute, err := strconv.ParseInt(match[2], 10, 64)
	if err != nil {
		return -1
	}
	return hour*TimeHour + minute*60
}

func matchSignPhotoDay(time time.Time, rule string) bool {
	if rule == "" {
		return true
	}
	day := int(time.Weekday())
	if day == 0 {
		day = 7
	}
	sDay := strconv.Itoa(day)
	for _, s := range strings.Split(rule, ",") {
		arr := strings.Split(s, "-")
		l := len(arr)
		if l == 1 && sDay == arr[0] {
			return true
		}
		if l < 2 {
			continue
		}
		start, err1 := strconv.Atoi(arr[0])
		end, err2 := strconv.Atoi(arr[1])
		if err1 != nil || err2 != nil {
			log.Println(s, err1, err2)
			matchSignPhotoErr(rule)
			continue
		}
		for i := start; i <= end; i++ {
			if i == day {
				return true
			}
		}
	}
	return false
}

func matchSignPhotoTime(tm time.Time, rule string) bool {
	if rule == "" {
		return true
	}
	t := tm.Unix() % TimeDay
	for _, s := range strings.Split(rule, ",") {
		switch s {
		case "am":
			if t < TimeNoon {
				return true
			}
		case "pm":
			if t >= TimeNoon {
				return true
			}
		default:
			arr := strings.Split(s, "-")
			l := len(arr)
			if l < 2 {
				log.Println(s)
				matchSignPhotoErr(rule)
				continue
			}
			start := parseSignPhotoTime(arr[0])
			end := parseSignPhotoTime(arr[1])
			if start < 0 || end < 0 {
				log.Println(s, start, end)
				matchSignPhotoErr(rule)
				continue
			}
			return start <= t && t <= end
		}
	}
	return false
}

func MatchSignPhotoKey(time time.Time, key string) bool {
	if key == "" {
		return true
	}
	rules := strings.Split(key, "|")
	if len(rules) == 0 {
		matchSignPhotoErr(key)
		return false
	}
	return matchSignPhotoDay(time, rules[0]) && matchSignPhotoTime(time, rules[1])
}

func GetSignPhotoImageUrl(imageId string, src bool) string {
	var u string
	if src {
		u = "https://p.ananas.chaoxing.com/star3/origin/"
	} else {
		u = "https://p.ananas.chaoxing.com/star3/170_220c/"
	}
	return u + imageId
}
