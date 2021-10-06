package EventBus

import (
	"fmt"
	"reflect"
	"sync"
)

//BusSubscriber defines subscription-related bus behavior
type BusSubscriber interface {
	Subscribe(topic string, fn interface{}) error
	SubscribeAsync(topic string, fn interface{}, transactional bool) error
	SubscribeOnce(topic string, fn interface{}) error
	SubscribeOnceAsync(topic string, fn interface{}) error
	Unsubscribe(topic string, handler interface{}) error
}

//BusPublisher defines publishing-related bus behavior
type BusPublisher interface {
	Publish(topic string, args ...interface{})
	PublishWaitAsync(topic string, args ...interface{}) *sync.WaitGroup
}

//BusController defines bus control behavior (checking handler's presence, synchronization)
type BusController interface {
	HasCallback(topic string) bool
	WaitAsync(wg *sync.WaitGroup)
}

//Bus englobes global (subscribe, publish, control) bus behavior
type Bus interface {
	BusController
	BusSubscriber
	BusPublisher
}

// EventBus - box for handlers and callbacks.
type EventBus struct {
	handlers map[string][]*eventHandler
	lock     sync.RWMutex // a lock for the map
}

type eventHandler struct {
	callBack      reflect.Value
	flagOnce      bool
	async         bool
	transactional bool
	sync.Mutex    // lock for an event handler - useful for running async callbacks serially
}

// New returns new EventBus with empty handlers.
func New() Bus {
	b := &EventBus{
		make(map[string][]*eventHandler),
		sync.RWMutex{},
	}
	return Bus(b)
}

// doSubscribe handles the subscription logic and is utilized by the public Subscribe functions
func (bus *EventBus) doSubscribe(topic string, fn interface{}, handler *eventHandler) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	if !(reflect.TypeOf(fn).Kind() == reflect.Func) {
		return fmt.Errorf("%s is not of type reflect.Func", reflect.TypeOf(fn).Kind())
	}
	bus.handlers[topic] = append(bus.handlers[topic], handler)
	return nil
}

// Subscribe subscribes to a topic.
// Returns error if `fn` is not a function.
func (bus *EventBus) Subscribe(topic string, fn interface{}) error {
	return bus.doSubscribe(topic, fn, &eventHandler{
		reflect.ValueOf(fn), false, false, false, sync.Mutex{},
	})
}

// SubscribeAsync subscribes to a topic with an asynchronous callback
// Transactional determines whether subsequent callbacks for a topic are
// run serially (true) or concurrently (false)
// Returns error if `fn` is not a function.
func (bus *EventBus) SubscribeAsync(topic string, fn interface{}, transactional bool) error {
	return bus.doSubscribe(topic, fn, &eventHandler{
		reflect.ValueOf(fn), false, true, transactional, sync.Mutex{},
	})
}

// SubscribeOnce subscribes to a topic once. Handler will be removed after executing.
// Returns error if `fn` is not a function.
func (bus *EventBus) SubscribeOnce(topic string, fn interface{}) error {
	return bus.doSubscribe(topic, fn, &eventHandler{
		reflect.ValueOf(fn), true, false, false, sync.Mutex{},
	})
}

// SubscribeOnceAsync subscribes to a topic once with an asynchronous callback
// Handler will be removed after executing.
// Returns error if `fn` is not a function.
func (bus *EventBus) SubscribeOnceAsync(topic string, fn interface{}) error {
	return bus.doSubscribe(topic, fn, &eventHandler{
		reflect.ValueOf(fn), true, true, false, sync.Mutex{},
	})
}

// HasCallback returns true if exists any callback subscribed to the topic.
func (bus *EventBus) HasCallback(topic string) bool {
	bus.lock.RLock()
	defer bus.lock.RUnlock()
	_, ok := bus.handlers[topic]
	if ok {
		return len(bus.handlers[topic]) > 0
	}
	return false
}

// Unsubscribe removes callback defined for a topic.
// Returns error if there are no callbacks subscribed to the topic.
func (bus *EventBus) Unsubscribe(topic string, handler interface{}) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	if _, ok := bus.handlers[topic]; ok && len(bus.handlers[topic]) > 0 {
		bus.removeHandler(topic, bus.findHandlerIdx(topic, reflect.ValueOf(handler)))
		return nil
	}
	return fmt.Errorf("topic %s doesn't exist", topic)
}

// Publish executes callback defined for a topic. Any additional argument will be transferred to the callback.
func (bus *EventBus) Publish(topic string, args ...interface{}) {
	// bus.lock.Lock() // will unlock if handler is not found or always after setUpPublish
	// defer bus.lock.Unlock()
	bus.PublishWaitAsync(topic, args...)
}

// Publish executes callback defined for a topic. Any additional argument will be transferred to the callback.
func (bus *EventBus) PublishWaitAsync(topic string, args ...interface{}) *sync.WaitGroup {
	// bus.lock.RLock() // will unlock if handler is not found or always after setUpPublish
	// defer bus.lock.RUnlock() // 执行once handler， 无法确定有多少个读锁，所以通过copy方式解决多线程处理问题
	wg := &sync.WaitGroup{} // 同步锁
	if handlers, ok := bus.handlers[topic]; ok && 0 < len(handlers) {
		// Handlers slice may be changed by removeHandler and Unsubscribe during iteration,
		// so make a copy and iterate the copied slice.
		copyHandlers := make([]*eventHandler, len(handlers))
		copy(copyHandlers, handlers)
		for i, handler := range copyHandlers {
			arguments, ok := bus.PassedArguments(handler.callBack.Type(), args...)
			if !ok {
				continue // 参数类型不匹配
			}
			if handler.flagOnce {
				bus.lock.Lock()             // 加锁map
				bus.removeHandler(topic, i) // 修改map， Unsubscribe(topic, handler)
				bus.lock.Unlock()           // 解锁map
			}
			if !handler.async {
				handler.callBack.Call(arguments)
			} else {
				wg.Add(1)
				if handler.transactional {
					handler.Lock()
				}
				go bus.doPublishAsync(wg, handler, arguments)
			}
		}
	}
	return wg
}

func (bus *EventBus) doPublishAsync(wg *sync.WaitGroup, handler *eventHandler, arguments []reflect.Value) {
	defer wg.Done()
	if handler.transactional {
		defer handler.Unlock()
	}
	handler.callBack.Call(arguments)
}

func (bus *EventBus) removeHandler(topic string, idx int) {
	if _, ok := bus.handlers[topic]; !ok {
		return
	}
	l := len(bus.handlers[topic])

	if !(0 <= idx && idx < l) {
		return
	}

	copy(bus.handlers[topic][idx:], bus.handlers[topic][idx+1:])
	bus.handlers[topic][l-1] = nil // or the zero value of T
	bus.handlers[topic] = bus.handlers[topic][:l-1]
}

func (bus *EventBus) findHandlerIdx(topic string, callback reflect.Value) int {
	if _, ok := bus.handlers[topic]; ok {
		for idx, handler := range bus.handlers[topic] {
			if handler.callBack.Type() == callback.Type() &&
				handler.callBack.Pointer() == callback.Pointer() {
				return idx
			}
		}
	}
	return -1
}

// 处理调用的参数
func (bus *EventBus) PassedArguments(funcType reflect.Type, args ...interface{}) ([]reflect.Value, bool) {
	if funcType.NumIn() == 0 {
		return make([]reflect.Value, 0), true // 无参数，直接调用
	}
	var variadicType reflect.Type
	variadicIdx := funcType.NumIn() // 可变参数位置，不存在可变参数，idx为参数数量
	if funcType.IsVariadic() {      // 具有可变参数，纠正可变参数位置
		if funcType.NumIn()-1 > len(args) {
			return nil, false // 缺少参数，禁止调用
		}
		variadicIdx -= 1
		variadicType = funcType.In(variadicIdx).Elem()
	} else if funcType.NumIn() != len(args) {
		return nil, false // 不存在可变参数，参数数量不相等
	}
	// 处理参数
	arguments := make([]reflect.Value, len(args))
	for i, v := range args {
		if v == nil {
			//arguments[i] = reflect.New(funcType.In(i)).Elem()
			arguments[i] = reflect.ValueOf(nil)
		} else if i >= variadicIdx && !reflect.TypeOf(v).AssignableTo(variadicType) {
			// variadic index
			return nil, false // 可变参数不匹配
		} else if i < variadicIdx && !reflect.TypeOf(v).AssignableTo(funcType.In(i)) {
			// ConvertibleTo or AssignableTo 不知道其区别， 懂的可以解释一下
			return nil, false // 参数类型无法匹配
		} else {
			arguments[i] = reflect.ValueOf(v)
		}
	}

	return arguments, true
}

// WaitAsync waits for all async callbacks to complete
func (bus *EventBus) WaitAsync(wg *sync.WaitGroup) {
	wg.Wait()
}
