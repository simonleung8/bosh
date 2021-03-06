package agent

import (
	boshmbus "bosh/mbus"
	testmbus "bosh/mbus/testhelpers"
	boshplatform "bosh/platform"
	testplatform "bosh/platform/testhelpers"
	boshsettings "bosh/settings"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRunHandlesAMessage(t *testing.T) {
	req := boshmbus.Request{Method: "ping"}
	expectedResp := boshmbus.Response{Value: "pong"}

	assertResponseForRequest(t, req, expectedResp)
}

func assertResponseForRequest(t *testing.T, req boshmbus.Request, expectedResp boshmbus.Response) {
	settings, handler, platform := getAgentDependencies()
	agent := New(settings, handler, platform)

	err := agent.Run()
	assert.NoError(t, err)
	assert.True(t, handler.ReceivedRun)

	resp := handler.Func(req)

	assert.Equal(t, resp, expectedResp)
}

func TestRunSetsUpHeartbeats(t *testing.T) {
	settings, handler, platform := getAgentDependencies()
	settings.Disks = boshsettings.Disks{
		System:     "/dev/sda1",
		Ephemeral:  "/dev/sdb",
		Persistent: map[string]string{"vol-xxxx": "/dev/sdf"},
	}

	platform = &testplatform.FakePlatform{
		CpuLoad:   boshplatform.CpuLoad{One: 1.0, Five: 5.0, Fifteen: 15.0},
		CpuStats:  boshplatform.CpuStats{User: 55, Sys: 44, Wait: 11, Total: 1000},
		MemStats:  boshplatform.MemStats{Used: 40 * 1024 * 1024, Total: 100 * 1024 * 1024},
		SwapStats: boshplatform.MemStats{Used: 10 * 1024 * 1024, Total: 100 * 1024 * 1024},
		DiskStats: map[string]boshplatform.DiskStats{
			"/":               boshplatform.DiskStats{Used: 25, Total: 100, InodeUsed: 300, InodeTotal: 1000},
			"/var/vcap/data":  boshplatform.DiskStats{Used: 5, Total: 100, InodeUsed: 150, InodeTotal: 1000},
			"/var/vcap/store": boshplatform.DiskStats{Used: 0, Total: 100, InodeUsed: 0, InodeTotal: 1000},
		},
	}

	agent := New(settings, handler, platform)
	agent.heartbeatInterval = time.Millisecond
	err := agent.Run()
	assert.NoError(t, err)

	hb := <-handler.HeartbeatChan

	assert.Equal(t, []string{"1.00", "5.00", "15.00"}, hb.Vitals.CpuLoad)

	assert.Equal(t, boshmbus.CpuStats{
		User: "5.5",
		Sys:  "4.4",
		Wait: "1.1",
	}, hb.Vitals.Cpu)

	assert.Equal(t, boshmbus.MemStats{
		Percent: "40",
		Kb:      "40960",
	}, hb.Vitals.UsedMem)

	assert.Equal(t, boshmbus.MemStats{
		Percent: "10",
		Kb:      "10240",
	}, hb.Vitals.UsedSwap)

	assert.Equal(t, boshmbus.Disks{
		System:     boshmbus.DiskStats{Percent: "25", InodePercent: "30"},
		Ephemeral:  boshmbus.DiskStats{Percent: "5", InodePercent: "15"},
		Persistent: boshmbus.DiskStats{Percent: "0", InodePercent: "0"},
	}, hb.Vitals.Disks)
}

func TestRunSetsUpHeartbeatsWithoutEphemeralOrPersistentDisk(t *testing.T) {
	settings, handler, platform := getAgentDependencies()
	settings.Disks = boshsettings.Disks{
		System: "/dev/sda1",
	}

	platform = &testplatform.FakePlatform{
		DiskStats: map[string]boshplatform.DiskStats{
			"/":               boshplatform.DiskStats{Used: 25, Total: 100, InodeUsed: 300, InodeTotal: 1000},
			"/var/vcap/data":  boshplatform.DiskStats{Used: 5, Total: 100, InodeUsed: 150, InodeTotal: 1000},
			"/var/vcap/store": boshplatform.DiskStats{Used: 0, Total: 100, InodeUsed: 0, InodeTotal: 1000},
		},
	}

	agent := New(settings, handler, platform)
	agent.heartbeatInterval = time.Millisecond
	err := agent.Run()
	assert.NoError(t, err)

	hb := <-handler.HeartbeatChan

	assert.Equal(t, boshmbus.Disks{
		System:     boshmbus.DiskStats{Percent: "25", InodePercent: "30"},
		Ephemeral:  boshmbus.DiskStats{},
		Persistent: boshmbus.DiskStats{},
	}, hb.Vitals.Disks)
}

func getAgentDependencies() (settings boshsettings.Settings, handler *testmbus.FakeHandler, platform *testplatform.FakePlatform) {
	settings = boshsettings.Settings{}
	handler = &testmbus.FakeHandler{}
	platform = &testplatform.FakePlatform{}
	return
}
