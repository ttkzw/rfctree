package rfctree_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ttkzw/rfctree"
	"github.com/ttkzw/rfctree/internal/cliutil"
)

func rfcLabelsToDocIds(rfcs []*rfctree.RfcLabel) []string {
	docIds := make([]string, 0, len(rfcs))
	for _, rfc := range rfcs {
		docIds = append(docIds, rfc.DocId)
	}
	return docIds
}

func TestRfcIndexSetTargets(t *testing.T) {
	assert := assert.New(t)

	var rfcs []*rfctree.RfcLabel
	rfcIndex, err := rfctree.NewRfcIndex("testdata/rfcdata", []*rfctree.Target{}, []string{}, false, []string{})
	assert.Nil(err)
	excludes, _ := cliutil.GetExcludes("testdata/exclude.txt", []string{})

	// target, follow=false
	rfcIndex.SetTargets([]*rfctree.Target{{DocId: "RFC0821"}}, []string{}, false, excludes)
	rfcs = rfcIndex.GetTargets()
	assert.Equal([]string{"RFC0821"}, rfcLabelsToDocIds(rfcs))
	rfcIndex.Clear()

	// keyword, follow=false
	rfcIndex.SetTargets([]*rfctree.Target{}, []string{"SMTP"}, false, excludes)
	rfcs = rfcIndex.GetTargets()
	assert.Equal([]string{"RFC0821", "RFC1893", "RFC2852", "RFC2821", "RFC3463", "RFC3885", "RFC4952", "RFC5337", "RFC6530", "RFC6531"}, rfcLabelsToDocIds(rfcs))
	rfcIndex.Clear()

	// target, follow=true
	rfcIndex.SetTargets([]*rfctree.Target{{DocId: "RFC0821"}}, []string{}, true, excludes)
	rfcs = rfcIndex.GetTargets()
	assert.Equal([]string{"RFC0821", "RFC0822", "RFC0974", "RFC1123", "RFC1341", "RFC1342", "RFC1425", "RFC1521", "RFC1522", "RFC1590", "RFC1651", "RFC1806", "RFC1846", "RFC1869", "RFC1891", "RFC1892", "RFC1893", "RFC1894", "RFC2045", "RFC2046", "RFC2047", "RFC2048", "RFC2049", "RFC2183", "RFC2184", "RFC2231", "RFC2298", "RFC2554", "RFC2646", "RFC2852", "RFC2821", "RFC2822", "RFC3461", "RFC3462", "RFC3463", "RFC3464", "RFC3676", "RFC3798", "RFC3885", "RFC3886", "RFC4021", "RFC4288", "RFC4289", "RFC4468", "RFC4865", "RFC4952", "RFC4954", "RFC5248", "RFC5335", "RFC5336", "RFC5337", "RFC5321", "RFC5322", "RFC5504", "RFC5825", "RFC6530", "RFC6531", "RFC6532", "RFC6533", "RFC6657", "RFC6838", "RFC6854", "RFC7504", "RFC8098", "RFC9694"}, rfcLabelsToDocIds(rfcs))
	rfcIndex.Clear()
}

func TestRfcIndexExist(t *testing.T) {
	assert := assert.New(t)

	excludes, _ := cliutil.GetExcludes("testdata/exclude.txt", []string{})
	rfcIndex, err := rfctree.NewRfcIndex("testdata/rfcdata", []*rfctree.Target{}, []string{}, false, excludes)
	assert.Nil(err)

	assert.True(rfcIndex.Exist("RFC0821"))
	assert.False(rfcIndex.Exist("RFC99999"))
	assert.False(rfcIndex.Exist(""))
}

func TestRfcIndexKeys(t *testing.T) {
	assert := assert.New(t)

	excludes, _ := cliutil.GetExcludes("testdata/exclude.txt", []string{})
	rfcIndex, err := rfctree.NewRfcIndex("testdata/rfcdata", []*rfctree.Target{}, []string{}, false, excludes)
	assert.Nil(err)

	docIds := rfcIndex.Keys()
	assert.Equal([]string{"RFC0821", "RFC0822", "RFC0974", "RFC1123", "RFC1341", "RFC1342", "RFC1425", "RFC1521", "RFC1522", "RFC1590", "RFC1651", "RFC1806", "RFC1846", "RFC1869", "RFC1891", "RFC1892", "RFC1893", "RFC1894", "RFC2045", "RFC2046", "RFC2047", "RFC2048", "RFC2049", "RFC2183", "RFC2184", "RFC2231", "RFC2298", "RFC2554", "RFC2646", "RFC2821", "RFC2822", "RFC2852", "RFC3461", "RFC3462", "RFC3463", "RFC3464", "RFC3676", "RFC3798", "RFC3885", "RFC3886", "RFC4021", "RFC4288", "RFC4289", "RFC4468", "RFC4865", "RFC4952", "RFC4954", "RFC5248", "RFC5321", "RFC5322", "RFC5335", "RFC5336", "RFC5337", "RFC5504", "RFC5825", "RFC6530", "RFC6531", "RFC6532", "RFC6533", "RFC6657", "RFC6838", "RFC6854", "RFC7504", "RFC8098", "RFC9694"}, docIds)
}

func TestRfcIndexValues(t *testing.T) {
	assert := assert.New(t)

	excludes, _ := cliutil.GetExcludes("testdata/exclude.txt", []string{})
	rfcIndex, err := rfctree.NewRfcIndex("testdata/rfcdata", []*rfctree.Target{}, []string{}, false, excludes)
	assert.Nil(err)

	rfcs := rfcIndex.Values()
	assert.Equal([]string{"RFC0821", "RFC0822", "RFC0974", "RFC1123", "RFC1341", "RFC1342", "RFC1425", "RFC1521", "RFC1522", "RFC1590", "RFC1651", "RFC1806", "RFC1846", "RFC1869", "RFC1891", "RFC1892", "RFC1893", "RFC1894", "RFC2045", "RFC2046", "RFC2047", "RFC2048", "RFC2049", "RFC2183", "RFC2184", "RFC2231", "RFC2298", "RFC2554", "RFC2646", "RFC2821", "RFC2822", "RFC2852", "RFC3461", "RFC3462", "RFC3463", "RFC3464", "RFC3676", "RFC3798", "RFC3885", "RFC3886", "RFC4021", "RFC4288", "RFC4289", "RFC4468", "RFC4865", "RFC4952", "RFC4954", "RFC5248", "RFC5321", "RFC5322", "RFC5335", "RFC5336", "RFC5337", "RFC5504", "RFC5825", "RFC6530", "RFC6531", "RFC6532", "RFC6533", "RFC6657", "RFC6838", "RFC6854", "RFC7504", "RFC8098", "RFC9694"}, rfcLabelsToDocIds(rfcs))
}

func TestRfcIndexValuesSortedByPubDate(t *testing.T) {
	assert := assert.New(t)

	excludes, _ := cliutil.GetExcludes("testdata/exclude.txt", []string{})
	rfcIndex, err := rfctree.NewRfcIndex("testdata/rfcdata", []*rfctree.Target{}, []string{}, false, excludes)
	assert.Nil(err)

	rfcs := rfcIndex.ValuesSortedByPubDate()
	assert.Equal([]string{"RFC0821", "RFC0822", "RFC0974", "RFC1123", "RFC1341", "RFC1342", "RFC1425", "RFC1521", "RFC1522", "RFC1590", "RFC1651", "RFC1806", "RFC1846", "RFC1869", "RFC1891", "RFC1892", "RFC1893", "RFC1894", "RFC2045", "RFC2046", "RFC2047", "RFC2048", "RFC2049", "RFC2183", "RFC2184", "RFC2231", "RFC2298", "RFC2554", "RFC2646", "RFC2852", "RFC2821", "RFC2822", "RFC3461", "RFC3462", "RFC3463", "RFC3464", "RFC3676", "RFC3798", "RFC3885", "RFC3886", "RFC4021", "RFC4288", "RFC4289", "RFC4468", "RFC4865", "RFC4952", "RFC4954", "RFC5248", "RFC5335", "RFC5336", "RFC5337", "RFC5321", "RFC5322", "RFC5504", "RFC5825", "RFC6530", "RFC6531", "RFC6532", "RFC6533", "RFC6657", "RFC6838", "RFC6854", "RFC7504", "RFC8098", "RFC9694"}, rfcLabelsToDocIds(rfcs))
}

func TestRfcIndexFind(t *testing.T) {
	assert := assert.New(t)

	var rfc *rfctree.RfcLabel
	excludes, _ := cliutil.GetExcludes("testdata/exclude.txt", []string{})
	rfcIndex, err := rfctree.NewRfcIndex("testdata/rfcdata", []*rfctree.Target{}, []string{}, false, excludes)
	assert.Nil(err)

	rfc = rfcIndex.Find("RFC0821")
	assert.Equal("RFC0821", rfc.DocId)

	// not exist
	rfc = rfcIndex.Find("RFC99999")
	assert.Nil(rfc)
}

func TestRfcIndexIsExcluded(t *testing.T) {
	assert := assert.New(t)

	excludes, _ := cliutil.GetExcludes("testdata/exclude.txt", []string{})
	rfcIndex, err := rfctree.NewRfcIndex("testdata/rfcdata", []*rfctree.Target{}, []string{}, false, excludes)
	assert.Nil(err)

	assert.True(rfcIndex.IsExcluded("RFC0952"))
	assert.False(rfcIndex.IsExcluded("RFC0821"))
}

func TestRfcIndexGetRelationScore(t *testing.T) {
	assert := assert.New(t)

	excludes, _ := cliutil.GetExcludes("testdata/exclude.txt", []string{})
	rfcIndex, err := rfctree.NewRfcIndex("testdata/rfcdata", []*rfctree.Target{}, []string{}, false, excludes)
	assert.Nil(err)

	assert.Equal(100, rfcIndex.GetRelationScore("RFC0821", "RFC2821").Value())
}
