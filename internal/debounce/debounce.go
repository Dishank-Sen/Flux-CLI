package debounce

import (
	"encoding/json"
	"exp1/internal/types"
	"os"
	"path"
	"sync"
	"time"
)

type Debouncer struct {
	mu     sync.Mutex
	timers map[string]*time.Timer
}

func NewDebouncer() *Debouncer {
	return &Debouncer{
		timers: make(map[string]*time.Timer),
	}
}

func (d *Debouncer) Debounce(key string, delay time.Duration, fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if timer, ok := d.timers[key]; ok {
		timer.Stop()
	}

	d.timers[key] = time.AfterFunc(delay, func() {
		d.mu.Lock()
		delete(d.timers, key)
		d.mu.Unlock()

		fn()
	})
}

//returns debounce time in seconds
func GetDebounceTime() (int64, error){
	// reads the config file for debounce time
	path := path.Join(".rec", "config.json")
	byteContent, err := os.ReadFile(path)
	if err != nil{
		return 0,err
	}
	var jsonData types.Config
	err = json.Unmarshal(byteContent, &jsonData)
	if err != nil{
		return 0,err
	}
	debounceTime := jsonData.Recorder.DebounceTime
	return debounceTime, nil
}

func SetDebounceTime(time int64) error{
	path := path.Join(".rec", "config.json")
	byteContent, err := os.ReadFile(path)
	if err != nil{
		return err
	}
	var jsonData types.Config
	err = json.Unmarshal(byteContent, &jsonData)
	if err != nil{
		return err
	}
	jsonData.Recorder.DebounceTime = time
	data, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil{
		return err
	}
	
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil{
		return err
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil{
		return err
	}

	return nil
}