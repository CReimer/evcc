package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/evcc-io/evcc/api/globalconfig"
	"github.com/evcc-io/evcc/core/keys"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/config"
)

var siteNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type siteConfig struct {
	name       string
	config     map[string]any
	loadpoints []string
	circuit    string
	tariffs    globalconfig.TariffRefs
	hems       globalconfig.Hems
}

func scopedSiteValues(out chan<- util.Param, name string, mirrorRoot bool) chan util.Param {
	in := make(chan util.Param)
	go func() {
		for param := range in {
			if param.Loadpoint == nil && param.Key == keys.Vehicles {
				if mirrorRoot {
					param.Site = ""
					out <- param
				}
				continue
			}
			param.Site = name
			out <- param
			if mirrorRoot {
				param.Site = ""
				out <- param
			}
		}
	}()
	return in
}

func normalizedSites(conf globalconfig.All, additionalLoadpoints ...config.Named) ([]siteConfig, error) {
	if len(conf.Sites) == 0 {
		return []siteConfig{{name: "default", config: conf.Site}}, nil
	}
	if len(conf.Site) != 0 {
		return nil, errors.New("site and sites cannot be used together")
	}

	res := make([]siteConfig, 0, len(conf.Sites))
	names := make(map[string]struct{}, len(conf.Sites))
	assigned := make(map[string]string, len(conf.Loadpoints))
	allLoadpoints := append(slices.Clone(conf.Loadpoints), additionalLoadpoints...)
	known := make(map[string]struct{}, len(allLoadpoints))
	for id, lp := range allLoadpoints {
		known[loadpointName(id, lp)] = struct{}{}
	}

	for _, site := range conf.Sites {
		name := strings.TrimSpace(site.Name)
		if !siteNamePattern.MatchString(name) {
			return nil, fmt.Errorf("invalid site name %q", site.Name)
		}
		if _, exists := site.Other["vehicles"]; exists {
			return nil, fmt.Errorf("site %q cannot assign global vehicles", name)
		}
		if _, exists := names[name]; exists {
			return nil, fmt.Errorf("duplicate site name %q", name)
		}
		names[name] = struct{}{}
		for _, lp := range site.Loadpoints {
			if _, exists := known[lp]; !exists {
				return nil, fmt.Errorf("site %q references unknown loadpoint %q", name, lp)
			}
			if owner, exists := assigned[lp]; exists {
				return nil, fmt.Errorf("loadpoint %q is assigned to sites %q and %q", lp, owner, name)
			}
			assigned[lp] = name
		}

		res = append(res, siteConfig{
			name:       name,
			config:     site.Other,
			loadpoints: slices.Clone(site.Loadpoints),
			circuit:    site.Circuit,
			tariffs:    site.Tariffs,
			hems:       site.HEMS,
		})
	}

	for id, lp := range allLoadpoints {
		name := loadpointName(id, lp)
		requested, _ := lp.Property("site").(string)
		if requested == "" {
			if _, exists := assigned[name]; !exists {
				return nil, fmt.Errorf("loadpoint %q is not assigned to a site", name)
			}
			continue
		}
		if _, exists := names[requested]; !exists {
			return nil, fmt.Errorf("loadpoint %q references unknown site %q", name, requested)
		}
		if owner, exists := assigned[name]; exists && owner != requested {
			return nil, fmt.Errorf("loadpoint %q is assigned to sites %q and %q", name, owner, requested)
		}
		assigned[name] = requested
		for i := range res {
			if res[i].name == requested && !slices.Contains(res[i].loadpoints, name) {
				res[i].loadpoints = append(res[i].loadpoints, name)
			}
		}
	}

	return res, nil
}

func loadpointName(id int, conf config.Named) string {
	if conf.Name != "" {
		return conf.Name
	}
	return fmt.Sprintf("lp-%d", id+1)
}
