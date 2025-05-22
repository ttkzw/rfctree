package rfctree

import (
	"fmt"
)

func List(rfcsDir string, targets []*Target, keywords []string, follow bool, excludes []string, outputFilename string) error {
	rfcIndex, err := NewRfcIndex(rfcsDir, targets, keywords, follow, excludes)
	if err != nil {
		return err
	}

	// for debug
	// for _, docIdA := range slices.Sorted(maps.Keys(rfcIndex.relationScoreMap)) {
	// 	fmt.Printf("%s\n", docIdA)
	// 	for _, docIdB := range slices.Sorted(maps.Keys(rfcIndex.relationScoreMap[docIdA])) {
	// 		fmt.Printf("    %s: %d\n", docIdB, rfcIndex.relationScoreMap[docIdA][docIdB])
	// 	}
	// }

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

	for _, rfc := range rfcIndex.GetTargets() {
		fmt.Printf("%s %s\n", rfc.DocId, rfc.Doc.Title)
	}

	return nil
}
