package core

import (
	"encoding/json"
	"github.com/moeshin/go-errs"
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

func SaveCacheImage() {
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
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			SaveCacheImage()
		}
	}()
}
