package EventBus

import (
	"reflect"
	"strings"
)

/**
 *
 * bus: 总线
 * data: 订阅集合
 * verify: true: 发现错误，立即返回
 * tag: 订阅标签, Method=~topic,Method=topic, ~：表示同步， 否则为异步
 */
func SubscribeTag(bus Bus, data interface{}, verify bool, tag string) (func(), error) {
	if tag == "" {
		tag = "gbus"
	}

	clss := []func(){}
	errs := []error{}
	typ := reflect.TypeOf(data)
	val := reflect.ValueOf(data)
	if typ.Kind() == reflect.Ptr { // 指针
		typ = typ.Elem()
		val = val.Elem()
	}
	for k := 0; k < typ.NumField(); k++ {
		tag := typ.Field(k).Tag.Get(tag)
		if tag == "" {
			continue
		}
		// method=~topic
		cfs := make(map[string]string)
		for _, v := range strings.Split(tag, ",") {
			if v == "" {
				continue
			}
			v2 := strings.SplitN(v, "=", 2)
			if len(v2) == 1 || v2[1] == "" {
				cfs[v2[0]] = ""
			} else {
				cfs[v2[0]] = v2[1]
			}
		}
		tfv := val.Field(k)
		tft := tfv.Type()
		for name, conf := range cfs {
			_, exsit := tft.MethodByName(name)
			if !exsit {
				continue // 方法不存在跳过
			}
			kind := BusAsync
			topic := conf
			if conf != "" && conf[0] != '~' {
				kind = BusSync
				topic = conf[1:]
			}
			fn := tfv.MethodByName(name).Interface()

			var err error
			// 注意， 内部总线是支持空主题的，空主题="default"主题
			switch kind {
			case BusAsync:
				err = bus.SubscribeAsync(topic, fn, false)
			default:
				err = bus.Subscribe(topic, fn)
			}
			if err != nil && verify {
				return nil, err
			} else if err != nil {
				errs = append(errs, err)
			} else {
				clss = append(clss, func() { bus.Unsubscribe(topic, fn) })
			}
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
