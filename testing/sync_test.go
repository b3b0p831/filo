package testing

import (
	"testing"
	"time"

	"bebop831.com/filo/config"
	"bebop831.com/filo/util"
)

func TestSync(t *testing.T) {

	cfg, err := config.Load()
	if err != nil {
		t.Error(err)
	}

	eventChan := make(chan struct{})
	exitChan := make(chan struct{})
	syncChan := make(chan struct{})
	ti, err := util.GetTimeInterval(cfg.SyncDelay)

	go util.Sync(eventChan, exitChan, syncChan, cfg)

	for range 1000 {
		eventChan <- struct{}{}
	}

	if err != nil {
		t.Error(err)
	}

	for range syncChan {
		t.Log("Performing Sync!")
		time.Sleep(ti + 10*time.Second)
		break
	}

	//How do I ensure that util sync has the proper behavior?
}
