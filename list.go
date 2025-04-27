package rfctree

import (
	"fmt"
)

func List(rfcsDir, outputFilename string, keywords []string) error {
	rfcIndex, err := NewRfcIndex(rfcsDir)
	if err != nil {
		return err
	}

	// for _, docId := range rfcIndex.Keys() {
	// 	rfc := rfcIndex.Find(docId)
	// 	for _, v := range rfc.Doc.Format {
	// 		if strings.HasPrefix(v, " ") {
	// 			fmt.Printf("%s %s\n", rfc.DocId, v)
	// 		}
	// 		if strings.HasSuffix(v, " ") {
	// 			fmt.Printf("%s %s\n", rfc.DocId, v)
	// 		}
	// 	}
	// }

	for _, rfc := range rfcIndex.FindAllByKeywords(keywords) {
		fmt.Printf("%s %s\n", rfc.DocId, rfc.Doc.Title)
	}

	return nil
}
