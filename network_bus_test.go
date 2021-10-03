package EventBus_test

import (
	"testing"

	"github.com/suisrc/EventBus"
)

func TestNewServer(t *testing.T) {
	serverBus := EventBus.NewServer(":2010", "/_server_bus_", EventBus.New())
	serverBus.Start()
	if serverBus == nil || !serverBus.Started() {
		t.Log("New server EventBus not created!")
		t.Fail()
	}
	serverBus.Stop()
}

func TestNewClient(t *testing.T) {
	clientBus := EventBus.NewClient(":2015", "/_client_bus_", EventBus.New())
	clientBus.Start()
	if clientBus == nil || !clientBus.Started() {
		t.Log("New client EventBus not created!")
		t.Fail()
	}
	clientBus.Stop()
}

func TestRegister(t *testing.T) {
	serverPath := "/_server_bus_"
	serverBus := EventBus.NewServer(":2010", serverPath, EventBus.New())

	args := &EventBus.SubscribeArg{":2010", serverPath, EventBus.PublishService, EventBus.Subscribe, "topic"}
	reply := new(bool)

	serverBus.Service().Register(args, reply)

	if serverBus.EventBus().HasCallback("topic_topic") {
		t.Fail()
	}
	if !serverBus.EventBus().HasCallback("topic") {
		t.Fail()
	}
}

func TestPushEvent(t *testing.T) {
	clientBus := EventBus.NewClient("localhost:2015", "/_client_bus_", EventBus.New())

	eventArgs := make([]interface{}, 1)
	eventArgs[0] = 10

	clientArg := &EventBus.ClientArg{eventArgs, "topic"}
	reply := new(bool)

	fn := func(a int) {
		if a != 10 {
			t.Fail()
		}
	}

	clientBus.EventBus().Subscribe("topic", fn)
	clientBus.Service().PushEvent(clientArg, reply)
	if !(*reply) {
		t.Fail()
	}
}

func TestServerPublish(t *testing.T) {
	serverBus := EventBus.NewServer(":2020", "/_server_bus_b", EventBus.New())
	serverBus.Start()

	fn := func(a int) {
		if a != 10 {
			t.Fail()
		}
	}

	clientBus := EventBus.NewClient(":2025", "/_client_bus_b", EventBus.New())
	clientBus.Start()

	clientBus.Subscribe("topic", fn, ":2010", "/_server_bus_b")

	serverBus.EventBus().Publish("topic", 10)

	clientBus.Stop()
	serverBus.Stop()
}

func TestNetworkBus(t *testing.T) {
	networkBusA := EventBus.NewNetworkBus(":2035", "/_net_bus_A")
	networkBusA.Start()

	networkBusB := EventBus.NewNetworkBus(":2030", "/_net_bus_B")
	networkBusB.Start()

	fnA := func(a int) {
		if a != 10 {
			t.Fail()
		}
	}
	networkBusA.Subscribe("topic-A", fnA, ":2030", "/_net_bus_B")
	networkBusB.EventBus().Publish("topic-A", 10)

	fnB := func(a int) {
		if a != 20 {
			t.Fail()
		}
	}
	networkBusB.Subscribe("topic-B", fnB, ":2035", "/_net_bus_A")
	networkBusA.EventBus().Publish("topic-B", 20)

	networkBusA.Stop()
	networkBusB.Stop()
}
