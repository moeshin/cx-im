package core

import (
	"encoding/json"
	"github.com/moeshin/go-errs"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

var CacheImage = &struct {
	Mutex *sync.Mutex
	Map   map[string]string
	Save  bool
}{
	Mutex: &sync.Mutex{},
	Map:   map[string]string{},
	Save:  false,
}

const CacheImagePath = "./cache-image.json"

func loadCacheImage() {
	f, err := os.Open(CacheImagePath)
	if err != nil {
		return
	}
	defer errs.Close(f)
	data, err := ioutil.ReadAll(f)
	errs.Panic(err)
	err = json.Unmarshal(data, &CacheImage.Map)
	errs.Panic(err)
}

func saveCacheImage() {
	CacheImage.Mutex.Lock()
	defer CacheImage.Mutex.Unlock()
	if !CacheImage.Save {
		return
	}
	data, err := json.MarshalIndent(CacheImage.Map, "", "  ")
	if errs.Print(err) {
		return
	}
	err = os.WriteFile(CacheImagePath, data, 0666)
	if errs.Print(err) {
		return
	}
	CacheImage.Save = false
}

func init() {
	loadCacheImage()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			saveCacheImage()
		}
	}()
}
