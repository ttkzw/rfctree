package rfctree

import (
	"errors"
	"image/color"
	"regexp"
	"strings"
	"time"
)

type Rfc struct {
	Doc *RfcDoc

	// DocIs is a document ID for display
	DocId string

	// Status is a status of RFC.
	Status Status

	// SubSeries is a sub-series of RFC.
	SubSeries string

	// PubDateInMonth is the number of months since January of the earliest publication year.
	PubDateInMonth int

	// PubYear is a publication year.
	PubYear int

	// Keywords are keywords that are normalized to uppercase.
	Keywords []string

	// IsTarget
	IsTarget bool

	// Position
	Position Position
}

func NewRfc(data []byte) (*Rfc, error) {
	rfcDoc, err := NewRfcDoc(data)
	if err != nil {
		return nil, err
	}

	t := toTime(rfcDoc.PubDate)

	rfc := Rfc{
		Doc:            rfcDoc,
		DocId:          toDisplayDocId(rfcDoc.DocId),
		Status:         NewStatus(rfcDoc.Status),
		SubSeries:      toSubSeries(rfcDoc.SeeAlso),
		PubDateInMonth: (t.Year()-yearOfOrigin)*12 + int(t.Month()) - 1,
		PubYear:        t.Year(),
		Keywords:       normalizeKeywords(rfcDoc.Keywords),
	}
	return &rfc, nil
}

type Position struct {
	X int
	Y int
}

// Status is a status of RFC.
type Status int

const (
	NOT_ISSUED = Status(iota)
	PROPOSED_STANDARD
	DRAFT_STANDARD
	INTERNET_STANDARD
	INFORMATIONAL
	EXPERIMENTAL
	HISTORIC
	BEST_CURRENT_PRACTICE
	UNKNOWN
)

var ErrRfcNotIssued = errors.New("rfc is not issued")

var statusString = [...]string{
	NOT_ISSUED:            "NOT ISSUED",
	PROPOSED_STANDARD:     "PROPOSED STANDARD",
	DRAFT_STANDARD:        "DRAFT STANDARD",
	INTERNET_STANDARD:     "INTERNET STANDARD",
	INFORMATIONAL:         "INFORMATIONAL",
	EXPERIMENTAL:          "EXPERIMENTAL",
	HISTORIC:              "HISTORIC",
	BEST_CURRENT_PRACTICE: "BEST CURRENT PRACTICE",
	UNKNOWN:               "UNKNOWN",
}

var statusDisplay = [...]string{
	NOT_ISSUED:            "NI",
	PROPOSED_STANDARD:     "PS",
	DRAFT_STANDARD:        "DS",
	INTERNET_STANDARD:     "STD",
	INFORMATIONAL:         "Informational",
	EXPERIMENTAL:          "Experimental",
	HISTORIC:              "Historic",
	BEST_CURRENT_PRACTICE: "BCP",
	UNKNOWN:               "Unknown",
}

var statusColor = [...]color.Color{
	NOT_ISSUED:            color.Black,
	PROPOSED_STANDARD:     color.RGBA{0, 0, 191, 223},
	DRAFT_STANDARD:        color.RGBA{63, 191, 191, 223},
	INTERNET_STANDARD:     color.RGBA{63, 127, 63, 223},
	INFORMATIONAL:         color.RGBA{191, 127, 63, 223},
	EXPERIMENTAL:          color.RGBA{191, 191, 63, 223},
	HISTORIC:              color.RGBA{127, 127, 127, 223},
	BEST_CURRENT_PRACTICE: color.RGBA{127, 63, 127, 223},
	UNKNOWN:               color.RGBA{255, 255, 255, 223},
}

func NewStatus(status string) Status {
	for i, s := range statusString {
		if status == s {
			return Status(i)
		}
	}
	return NOT_ISSUED
}

func (s Status) String() string {
	i := int(s)
	if 0 <= i && i < len(statusString) {
		return statusString[i]
	}
	return statusString[0]
}

func (s Status) Display() string {
	i := int(s)
	if 0 <= i && i < len(statusDisplay) {
		return statusDisplay[i]
	}
	return statusDisplay[0]
}

func (s Status) Color() color.Color {
	i := int(s)
	if 0 <= i && i < len(statusColor) {
		return statusColor[i]
	}
	return statusColor[0]
}

var docIdRe = regexp.MustCompile(`^([A-Z]+)0+([0-9]+?)$`)

// toDisplayDocId converts a string like "RFC0123" to a string like "RFC 123" for display purposes.
func toDisplayDocId(docId string) string {
	return docIdRe.ReplaceAllString(docId, "$1 $2")
}

func toSubSeries(seeAlso []string) string {
	for _, ss := range []string{"STD", "BCP", "FYI"} {
		for _, sa := range seeAlso {
			if strings.HasPrefix(sa, ss) {
				return toDisplayDocId(sa)
			}
		}
	}
	return ""
}

// the year of the oldest publish date, RFC 31.
const yearOfOrigin int = 1968

func toTime(pubDate string) time.Time {
	var t time.Time
	var err error
	layout := "2 January 2006"
	if len(strings.Split(pubDate, " ")) == 2 {
		pubDate = "1 " + pubDate
	}

	t, err = time.Parse(layout, pubDate)
	if err != nil {
		panic(err)
	}
	return t
}

func normalizeKeywords(keywords []string) []string {
	normalized := make([]string, len(keywords))
	for _, keyword := range keywords {
		normalized = append(normalized, strings.ToUpper(keyword))
	}
	return normalized
}
