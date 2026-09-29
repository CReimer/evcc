package util

import (
	"testing"

	"github.com/evcc-io/evcc/util/encode"
	"github.com/stretchr/testify/assert"
)

func TestParam(t *testing.T) {
	lp := 2
	p := Param{
		Key: "power",
		Val: 4711,
	}
	assert.Equal(t, "power", p.UniqueID())

	p.Loadpoint = &lp
	assert.Equal(t, "2.power", p.UniqueID())

	p.Site = "office"
	assert.Equal(t, "sites.office.loadpoints.2.power", p.UniqueID())
}

func TestParamCache(t *testing.T) {
	NewParamCache().Add("foo", Param{})
}

func TestParamCacheSites(t *testing.T) {
	cache := NewParamCache()
	lp := 0
	cache.Add("sites.office.gridPower", Param{Site: "office", Key: "gridPower", Val: 1200})
	cache.Add("sites.office.loadpoints.0.chargePower", Param{Site: "office", Loadpoint: &lp, Key: "chargePower", Val: 900})

	state := cache.State(encode.NewEncoder())
	sites := state["sites"].(map[string]map[string]any)
	office := sites["office"]
	assert.Equal(t, 1200, office["gridPower"])
	assert.Equal(t, 900, office["loadpoints"].([]map[string]any)[0]["chargePower"])
}

func TestParamCacheSnapshot(t *testing.T) {
	in := make(chan Param)
	go NewParamCache().Run(in)

	in <- Param{Key: "before", Val: 1}

	res := make(chan []Param, 1)
	in <- Param{Val: Snapshot(func(state []Param) { res <- state })}

	// published after the snapshot request, must not be included
	in <- Param{Key: "after", Val: 2}

	state := <-res
	assert.Len(t, state, 1)
	assert.Equal(t, "before", state[0].Key)
}
