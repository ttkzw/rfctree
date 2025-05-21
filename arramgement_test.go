package rfctree_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ttkzw/rfctree"
	"github.com/ttkzw/rfctree/internal/cliutil"
)

func TestArrangementArrange(t *testing.T) {

	assert := assert.New(t)

	excludes, _ := cliutil.GetExcludes("testdata/exclude.txt", []string{})
	rfcIndex, err := rfctree.NewRfcIndex("testdata/rfcdata", []*rfctree.Target{}, []string{}, false, excludes)
	assert.Nil(err)

	rfcIndex.SetTargets([]*rfctree.Target{{DocId: "RFC0821"}}, []string{}, true, excludes)
	rfcs := rfcIndex.GetTargets()

	arrangement := rfctree.NewArrangement()
	arrangement.Arrange(rfcIndex, rfcs)

}
