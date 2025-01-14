package EventBus

import (
	"reflect"
	"sort"
)

/**
 *
 * bus: 总线
 * hdl: 订阅的内容
 */
func SubscribeEvent(bus Bus, hdl EventHandler) (func(), error) {
	kind, topic, handler := hdl.Subscribe()
	switch kind {
	case BusAsync:
		if err := bus.SubscribeAsync(topic, handler, false); err != nil {
			return nil, err
		}
	case BusOnceSync:
		if err := bus.SubscribeOnce(topic, handler); err != nil {
			return nil, err
		}
	case BusOnceAsync:
		if err := bus.SubscribeOnceAsync(topic, handler); err != nil {
			return nil, err
		}
	default:
		if err := bus.Subscribe(topic, handler); err != nil {
			return nil, err
		}
	}
	return func() { bus.Unsubscribe(topic, handler) }, nil
}

/**
 *
 * bus: 总线
 * data: 订阅集合
 * verify: true: 发现错误，立即返回
 * excludes: 集合中排除的属性
 */
func SubscribeBatch(bus Bus, data interface{}, verify bool, excludes ...string) (func(), error) {
	sort.Strings(excludes)
	clss := []func(){}
	errs := []error{}
	typ := reflect.TypeOf(data).Elem() // 指针
	val := reflect.ValueOf(data).Elem()
	for k := 0; k < typ.NumField(); k++ {
		if idx := SearchStrings(&excludes, typ.Field(k).Name); idx >= 0 {
			continue
		}
		var cls func()
		var err error
		switch value := val.Field(k).Interface().(type) {
		case EventHandler:
			cls, err = SubscribeEvent(bus, value)
		case EventSubscriber:
			cls, err = value.Subscribe(bus)
		}
		if err != nil && verify {
			return nil, err
		} else if err != nil {
			errs = append(errs, err)
		} else if cls != nil {
			clss = append(clss, cls)
		}
	}
	clear := func() {
		for _, opt := range clss {
			opt()
		}
	}
	if len(errs) > 0 {
		return clear, NewMultiError(&errs)
	}
	return clear, nil
}
