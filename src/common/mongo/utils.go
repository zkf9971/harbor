package mongo

import (
	"strings"
)

func getDBAndCollectionName(collectionPrefixedWithDB string) (dBName, collectionName string) {

	names := strings.Split(collectionPrefixedWithDB, ":")
	dBName = names[0]
	collectionName = names[1]

	return
}
