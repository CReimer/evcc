package util

import (
	"maps"
	"slices"
	"strconv"
	"sync"

	"github.com/evcc-io/evcc/util/encode"
)

// Param is the broadcast channel data type
type Param struct {
	Site      string
	Loadpoint *int
	Key       string
	Val       any
}

// UniqueID returns unique identifier for parameter Loadpoint/Key combination
func (p Param) UniqueID() string {
	prefix := ""
	if p.Site != "" {
		prefix = "sites." + p.Site + "."
	}
	if p.Loadpoint != nil {
		if p.Site == "" {
			return strconv.Itoa(*p.Loadpoint) + "." + p.Key
		}
		return prefix + "loadpoints." + strconv.Itoa(*p.Loadpoint) + "." + p.Key
	}

	return prefix + p.Key
}

// ParamCache is a data store
type ParamCache struct {
	mu  sync.RWMutex
	val map[string]Param
}

// Snapshot requests a copy of the cache state at the parameter's position in
// the stream. It runs on the cache's goroutine and must not block for long.
type Snapshot func([]Param)

// NewCache creates cache
func NewParamCache() *ParamCache {
	return &ParamCache{
		val: make(map[string]Param),
	}
}

// Run adds input channel's values to cache
func (c *ParamCache) Run(in <-chan Param) {
	for p := range in {
		if snapshot, ok := p.Val.(Snapshot); ok {
			snapshot(c.All())
			continue
		}

		c.Add(p.UniqueID(), p)
	}
}

// State provides a structured copy of the cached values.
// Loadpoints are aggregated as loadpoints array.
// Result values are formatted using encoder.
func (c *ParamCache) State(enc encode.Encoder) map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	res := make(map[string]any)
	lps := make(map[int]map[string]any)
	sites := make(map[string]map[string]any)
	siteLoadpoints := make(map[string]map[int]map[string]any)

	for _, param := range c.val {
		if param.Site != "" {
			site, ok := sites[param.Site]
			if !ok {
				site = make(map[string]any)
				sites[param.Site] = site
			}
			if param.Loadpoint == nil {
				site[param.Key] = enc.Encode(param.Val)
				continue
			}
			byID, ok := siteLoadpoints[param.Site]
			if !ok {
				byID = make(map[int]map[string]any)
				siteLoadpoints[param.Site] = byID
			}
			lp, ok := byID[*param.Loadpoint]
			if !ok {
				lp = make(map[string]any)
				byID[*param.Loadpoint] = lp
			}
			lp[param.Key] = enc.Encode(param.Val)
			continue
		}
		if param.Loadpoint == nil {
			res[param.Key] = enc.Encode(param.Val)
		} else {
			lp, ok := lps[*param.Loadpoint]
			if !ok {
				lp = make(map[string]any)
				lps[*param.Loadpoint] = lp
			}
			lp[param.Key] = enc.Encode(param.Val)
		}
	}

	// convert map to array
	loadpoints := make([]map[string]any, len(lps))
	for id, lp := range lps {
		loadpoints[id] = lp
	}
	res["loadpoints"] = loadpoints
	for name, byID := range siteLoadpoints {
		loadpoints := make([]map[string]any, len(byID))
		for id, lp := range byID {
			loadpoints[id] = lp
		}
		sites[name]["loadpoints"] = loadpoints
	}
	if len(sites) > 0 {
		res["sites"] = sites
	}

	return res
}

// All provides a copy of the cached values
func (c *ParamCache) All() []Param {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return slices.Collect(maps.Values(c.val))
}

// Add entry to cache
func (c *ParamCache) Add(key string, param Param) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.val[key] = param
}

// Get entry from cache
func (c *ParamCache) Get(key string) Param {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if val, ok := c.val[key]; ok {
		return val
	}

	return Param{}
}
