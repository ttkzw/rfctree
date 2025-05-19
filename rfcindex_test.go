package rfctree_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ttkzw/rfctree"
)

func rfcLabelsToDocIds(rfcs []*rfctree.RfcLabel) []string {
	docIds := make([]string, 0, len(rfcs))
	for _, rfc := range rfcs {
		docIds = append(docIds, rfc.DocId)
	}
	return docIds
}

func TestNewRfcIndex(t *testing.T) {
	assert := assert.New(t)

	var docIds []string
	rfcIndex, err := rfctree.NewRfcIndex("rfcs", []*rfctree.Target{}, []string{}, false, []string{})
	assert.Nil(err)

	// follow=false
	rfcIndex.SetTargets([]*rfctree.Target{{DocId: "RFC0821"}}, []string{}, false, []string{})
	docIds = rfcLabelsToDocIds(rfcIndex.GetTargets())
	assert.Equal([]string{"RFC0821"}, docIds)
	rfcIndex.Clear()

	// follow=true
	rfcIndex.SetTargets([]*rfctree.Target{{DocId: "RFC0821"}}, []string{}, true, []string{})
	docIds = rfcLabelsToDocIds(rfcIndex.GetTargets())
	assert.Equal([]string{"RFC0772", "RFC0780", "RFC0788", "RFC0821", "RFC0974", "RFC1425", "RFC1651", "RFC1869", "RFC2821", "RFC5321"}, docIds)
	rfcIndex.Clear()
}
