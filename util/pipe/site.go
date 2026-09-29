package pipe

import "github.com/evcc-io/evcc/util"

type siteFilter struct {
	site string
}

// NewSiteFilter forwards only parameters belonging to the requested site.
func NewSiteFilter(site string) Piper {
	return &siteFilter{site: site}
}

func (f *siteFilter) Pipe(in <-chan util.Param) <-chan util.Param {
	out := make(chan util.Param)
	go func() {
		defer close(out)
		for param := range in {
			if param.Site == f.site {
				out <- param
			}
		}
	}()
	return out
}
