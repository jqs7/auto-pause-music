package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/micmonay/keybd_event"
	"github.com/moutend/go-hook/keyboard"
	"github.com/moutend/go-hook/mouse"
)

func main() {
	var isInterrupted bool
	var wg sync.WaitGroup

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	mouseChan := make(chan mouse.MSLLHOOKSTRUCT)
	keyboardChan := make(chan keyboard.KBDLLHOOKSTRUCT, 1024*1024)

	go func() {
		wg.Add(1)
		mouse.Notify(ctx, mouseChan)
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		keyboard.Notify(ctx, keyboardChan)
		wg.Done()
	}()
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		panic(err)
	}
	active := true
	lastActive := time.Now()
	for {
		if isInterrupted {
			cancel()
			break
		}
		select {
		case <-signalChan:
			isInterrupted = true
		case k := <-keyboardChan:
			if !active && k.VKCode != keybd_event.VK_MEDIA_PLAY_PAUSE {
				active = true
				kb.SetKeys(keybd_event.VK_MEDIA_PLAY_PAUSE)
				kb.Launching()
			}
			lastActive = time.Now()
		case <-mouseChan:
			if !active {
				active = true
				kb.SetKeys(keybd_event.VK_MEDIA_PLAY_PAUSE)
				kb.Launching()
			}
			lastActive = time.Now()
		case <-time.Tick(time.Second):
			if active && time.Now().Sub(lastActive) > time.Second*3 {
				active = false
				kb.SetKeys(keybd_event.VK_MEDIA_PLAY_PAUSE)
				kb.Launching()
			}
		}
	}
	wg.Wait()
}
