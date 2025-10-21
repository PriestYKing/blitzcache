package blitzcache

import (
	"sync"
	"time"
)

type TimingWheel struct {
	tickDuration time.Duration
	wheelSize    int
	slots        [][]string
	currentSlot  int
	mu           sync.Mutex
	ticker       *time.Ticker
	stopChan     chan struct{}
	onExpire     func(string)
}

func NewTimingWheel(tickDuration, wheelDuration time.Duration, onExpire func(string)) *TimingWheel {
	wheelSize := int(wheelDuration / tickDuration)

	tw := &TimingWheel{
		tickDuration: tickDuration,
		wheelSize:    wheelSize,
		slots:        make([][]string, wheelSize),
		currentSlot:  0,
		stopChan:     make(chan struct{}),
		onExpire:     onExpire,
	}

	for i := 0; i < wheelSize; i++ {
		tw.slots[i] = make([]string, 0, 16)
	}

	return tw
}

func (tw *TimingWheel) Start() {
	tw.ticker = time.NewTicker(tw.tickDuration)

	go func() {
		for {
			select {
			case <-tw.ticker.C:
				tw.tick()
			case <-tw.stopChan:
				return
			}
		}
	}()
}

func (tw *TimingWheel) Stop() {
	tw.ticker.Stop()
	close(tw.stopChan)
}

func (tw *TimingWheel) Add(key string, ttl time.Duration) {
	slots := int(ttl / tw.tickDuration)
	if slots >= tw.wheelSize {
		slots = tw.wheelSize - 1
	}

	tw.mu.Lock()
	targetSlot := (tw.currentSlot + slots) % tw.wheelSize
	tw.slots[targetSlot] = append(tw.slots[targetSlot], key)
	tw.mu.Unlock()
}

func (tw *TimingWheel) tick() {
	tw.mu.Lock()
	tw.currentSlot = (tw.currentSlot + 1) % tw.wheelSize
	expired := tw.slots[tw.currentSlot]
	tw.slots[tw.currentSlot] = make([]string, 0, 16)
	tw.mu.Unlock()

	for _, key := range expired {
		if tw.onExpire != nil {
			tw.onExpire(key)
		}
	}
}
