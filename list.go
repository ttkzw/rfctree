package rfctree

import "fmt"

func List(rfcsDir, outputFilename string, keywords []string) error {
	rfcMap, err := ReadRfcJsonFiles(rfcsDir)
	if err != nil {
		return err
	}

	// for _, docId := range slices.Sorted(maps.Keys(rfcMap)) {
	// 	rfc := rfcMap[docId]
	// 	for _, v := range rfc.Format {
	// 		if strings.HasPrefix(v, " ") {
	// 			fmt.Printf("%s %s\n", rfc.DocId, v)
	// 		}
	// 		if strings.HasSuffix(v, " ") {
	// 			fmt.Printf("%s %s\n", rfc.DocId, v)
	// 		}
	// 	}
	// }

	targetDocIds := getTargetDocIds(rfcMap, keywords)

	for _, docId := range targetDocIds {
		rfc := rfcMap[docId]
		fmt.Printf("%s %s\n", docId, rfc.Title)
	}

	return nil
}
