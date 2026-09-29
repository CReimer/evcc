package pipe

import (
	"testing"

	"github.com/evcc-io/evcc/util"
	"github.com/stretchr/testify/assert"
)

func TestSiteFilter(t *testing.T) {
	in := make(chan util.Param, 2)
	in <- util.Param{Site: "home", Key: "gridPower"}
	in <- util.Param{Site: "office", Key: "gridPower"}
	close(in)

	out := NewSiteFilter("office").Pipe(in)
	assert.Equal(t, "office", (<-out).Site)
	_, open := <-out
	assert.False(t, open)
}
