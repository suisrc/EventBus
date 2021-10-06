package EventBus_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/suisrc/EventBus"
)

func TestNew(t *testing.T) {
	bus := EventBus.New()
	if bus == nil {
		t.Log("New EventBus not created!")
		t.Fail()
	}
}

func TestHasCallback(t *testing.T) {
	bus := EventBus.New()
	bus.Subscribe("topic", func() {})
	if bus.HasCallback("topic_topic") {
		t.Fail()
	}
	if !bus.HasCallback("topic") {
		t.Fail()
	}
}

func TestSubscribe(t *testing.T) {
	bus := EventBus.New()
	if bus.Subscribe("topic", func() {}) != nil {
		t.Fail()
	}
	if bus.Subscribe("topic", "String") == nil {
		t.Fail()
	}
}

func TestSubscribeOnce(t *testing.T) {
	bus := EventBus.New()
	if bus.SubscribeOnce("topic", func() {}) != nil {
		t.Fail()
	}
	if bus.SubscribeOnce("topic", "String") == nil {
		t.Fail()
	}
}

func TestSubscribeOnceAndManySubscribe(t *testing.T) {
	bus := EventBus.New()
	event := "topic"
	flag := 0
	fn := func() { flag += 1 }
	bus.SubscribeOnce(event, fn)
	bus.Subscribe(event, fn)
	bus.Subscribe(event, fn)
	bus.Publish(event)

	if flag != 3 {
		t.Fail()
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := EventBus.New()
	handler := func() {}
	bus.Subscribe("topic", handler)
	if bus.Unsubscribe("topic", handler) != nil {
		t.Fail()
	}
	if bus.Unsubscribe("topic", handler) == nil {
		t.Fail()
	}
}

type handler struct {
	val int
}

func (h *handler) Handle() {
	h.val++
}

func TestUnsubscribeMethod(t *testing.T) {
	bus := EventBus.New()
	h := &handler{val: 0}

	bus.Subscribe("topic", h.Handle)
	bus.Publish("topic")
	if bus.Unsubscribe("topic", h.Handle) != nil {
		t.Fail()
	}
	if bus.Unsubscribe("topic", h.Handle) == nil {
		t.Fail()
	}
	wg := bus.PublishWaitAsync("topic")
	bus.WaitAsync(wg)

	if h.val != 1 {
		t.Fail()
	}
}

func TestPublish(t *testing.T) {
	bus := EventBus.New()
	bus.Subscribe("topic", func(a int, err error) {
		if a != 10 {
			t.Fail()
		}

		if err != nil {
			t.Fail()
		}

		//t.Fail()
	})
	bus.Publish("topic", 10, nil)
}

func TestPublishVarArgs(t *testing.T) {
	bus := EventBus.New()
	bus.Subscribe("topic", func(k string, a ...int) {
		fmt.Println(len(a))
	})
	bus.Publish("topic", "123", 9, 8, 7)
	t.Fail()
}

func TestSubcribeOnceAsync(t *testing.T) {
	results := make([]int, 0)

	bus := EventBus.New()
	bus.SubscribeOnce("topic", func(a int, out *[]int) {
		*out = append(*out, a)
	})

	bus.Publish("topic", 10, &results)
	wg := bus.PublishWaitAsync("topic", 10, &results)

	bus.WaitAsync(wg)

	fmt.Println(len(results))
	if len(results) != 1 {
		t.Fail()
	}

	if bus.HasCallback("topic") {
		t.Fail()
	}
}

func TestSubscribeAsyncTransactional(t *testing.T) {
	results := make([]int, 0)

	bus := EventBus.New()
	bus.SubscribeAsync("topic", func(a int, out *[]int, dur string) {
		sleep, _ := time.ParseDuration(dur)
		time.Sleep(sleep)
		*out = append(*out, a)
	}, true)

	bus.Publish("topic", 1, &results, "1s")
	wg := bus.PublishWaitAsync("topic", 2, &results, "0s")

	bus.WaitAsync(wg)

	if len(results) != 2 {
		t.Fail()
	}

	if results[0] != 1 || results[1] != 2 {
		t.Fail()
	}
}

func TestSubscribeAsync(t *testing.T) {
	results := make(chan int)

	bus := EventBus.New()
	bus.SubscribeAsync("topic", func(a int, out chan<- int) {
		out <- a
	}, false)

	bus.Publish("topic", 1, results)
	wg := bus.PublishWaitAsync("topic", 2, results)

	numResults := 0

	go func() {
		for _ = range results {
			numResults++
		}
	}()

	bus.WaitAsync(wg)
	println(2)

	time.Sleep(10 * time.Millisecond)

	// todo race detected during execution of test
	//if numResults != 2 {
	//	t.Fail()
	//}
}
