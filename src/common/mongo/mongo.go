package mongo

import (
	"strings"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//http://goinbigdata.com/how-to-build-microservice-with-mongodb-in-golang/
//https://godoc.org/gopkg.in/mgo.v2/bson#M
//http://denis.papathanasiou.org/posts/2012.10.14.post.html
//http://crypticjags.com/golang/golang-mgo-find-all-documents-in-mongodb.html
//http://stackoverflow.com/questions/20215510/cannot-retrieve-id-value-using-mgo-with-golang
func FindOneDocument(collectionPrefixedWithDB string, query bson.M) (result bson.M, err error) {

	dBName, collectionName := getDBAndCollectionName(collectionPrefixedWithDB)

	session := GetSession()
	defer session.Close()

	result = bson.M{}

	c := session.DB(dBName).C(collectionName)

	//https://godoc.org/gopkg.in/mgo.v2#Collection.Find
	err = c.Find(query).One(&result)

	if err != nil {
		log.Errorf("query err: %v", err)
	}

	return
}

// GetQuery constructs mgo.Query based on query parameters
func GetQuery(collectionPrefixedWithDB string, query bson.M, sort []string, pagingOps map[string]int) (q *mgo.Query, session *mgo.Session) {

	dBName, collectionName := getDBAndCollectionName(collectionPrefixedWithDB)

	session = GetSession()

	c := session.DB(dBName).C(collectionName)
	q = c.Find(query)

	if sort != nil && len(sort) > 0 {
		q = q.Sort(strings.Join(sort, ","))
	}

	if pagingOps != nil && pagingOps["offset"] > 0 {
		q = q.Skip(pagingOps["offset"])
	}
	if pagingOps != nil && pagingOps["limit"] > 0 {
		q = q.Limit(pagingOps["limit"])
	}

	return
}

// FindAccessLogs writes result to []models.AccessLog
func FindAccessLogs(collectionPrefixedWithDB string, query bson.M, sort []string, pagingOps map[string]int, result *[]models.AccessLog) (err error) {

	log.Debugf("Repo Query: %v", query)
	q, session := GetQuery(collectionPrefixedWithDB, query, sort, pagingOps)
	defer session.Close()

	//https://godoc.org/gopkg.in/mgo.v2#Collection.Find
	err = q.All(result)

	if err != nil {
		log.Errorf("query err: %v", err)
	}
	return
}

// FindRepos writes result to []models.RepoRecord
func FindRepos(collectionPrefixedWithDB string, query bson.M, sort []string, pagingOps map[string]int, result *[]models.RepoRecord) (err error) {

	log.Debugf("Repo Query: %v", query)
	q, session := GetQuery(collectionPrefixedWithDB, query, sort, pagingOps)
	defer session.Close()

	//https://godoc.org/gopkg.in/mgo.v2#Collection.Find
	err = q.All(result)

	if err != nil {
		log.Errorf("query err: %v", err)
	}
	return
}

// FindRoles writes result to []models.Role
func FindRoles(collectionPrefixedWithDB string, query bson.M, sort []string, pagingOps map[string]int, result *[]models.Role) (err error) {

	log.Debugf("Role Query: %v", query)

	q, session := GetQuery(collectionPrefixedWithDB, query, sort, pagingOps)
	defer session.Close()

	//https://godoc.org/gopkg.in/mgo.v2#Collection.Find
	err = q.All(result)

	log.Debugf("Result: %v", *result)
	log.Debugf("Result len: %v", len(*result))

	if err != nil {
		log.Errorf("query err: %v", err)
	}
	return
}

// FindUsers writes result to []models.User
func FindUsers(collectionPrefixedWithDB string, query bson.M, sort []string, pagingOps map[string]int, result *[]models.User) (err error) {

	log.Debugf("User Query: %v", query)

	// if query["username"] != nil {
	// 	debug.PrintStack()
	// }
	q, session := GetQuery(collectionPrefixedWithDB, query, sort, pagingOps)
	defer session.Close()

	//https://godoc.org/gopkg.in/mgo.v2#Collection.Find
	err = q.All(result)

	// log.Debugf("Result: %v", *result)
	// log.Debugf("Result len: %v", len(*result))

	if err != nil {
		log.Errorf("query err: %v", err)
	}
	return
}

// FindProjects writes result to []models.Project
func FindProjects(collectionPrefixedWithDB string, query bson.M, sort []string, pagingOps map[string]int, result *[]models.Project) (err error) {

	log.Debugf("Project Query: %v", query)

	q, session := GetQuery(collectionPrefixedWithDB, query, sort, pagingOps)
	defer session.Close()

	//https://godoc.org/gopkg.in/mgo.v2#Collection.Find
	err = q.All(result)

	log.Debugf("Result: %v", *result)
	log.Debugf("Result len: %v", len(*result))

	if err != nil {
		log.Errorf("query err: %v", err)
	}
	return
}

// FindDocuments writes result to []bson.M
func FindDocuments(collectionPrefixedWithDB string, query bson.M, sort []string, pagingOps map[string]int, result *[]bson.M) (err error) {

	log.Debugf("%s Query: %v", collectionPrefixedWithDB, query)

	q, session := GetQuery(collectionPrefixedWithDB, query, sort, pagingOps)
	defer session.Close()

	//https://godoc.org/gopkg.in/mgo.v2#Collection.Find
	err = q.All(result)

	log.Debugf("Result: %v", *result)
	log.Debugf("Result len: %v", len(*result))

	if err != nil {
		log.Errorf("query err: %v", err)
	}
	return
}

// CountDocuments returns number of documents based on query parameters
func CountDocuments(collectionPrefixedWithDB string, query bson.M, sort []string, pagingOps map[string]int) (result int, err error) {

	q, session := GetQuery(collectionPrefixedWithDB, query, sort, pagingOps)
	defer session.Close()

	//https://godoc.org/gopkg.in/mgo.v2#Collection.Find
	result, err = q.Count()

	if err != nil {
		log.Errorf("query err: %v", err)
	}
	return
}

// FindDistinctProjects writes result to []models.Project
func FindDistinctProjects(collectionPrefixedWithDB string, query bson.M, sort []string, pagingOps map[string]int, key string, result *[]models.Project) (err error) {

	q, session := GetQuery(collectionPrefixedWithDB, query, sort, pagingOps)
	defer session.Close()

	//https://godoc.org/gopkg.in/mgo.v2#Collection.Find
	err = q.Distinct(key, result)

	if err != nil {
		log.Errorf("query err: %v", err)
	}
	return
}

// FindDistinctDocuments writes result to []bson.M
func FindDistinctDocuments(collectionPrefixedWithDB string, query bson.M, sort []string, pagingOps map[string]int, key string, result *[]bson.M) (err error) {

	q, session := GetQuery(collectionPrefixedWithDB, query, sort, pagingOps)
	defer session.Close()

	//https://godoc.org/gopkg.in/mgo.v2#Collection.Find
	err = q.Distinct(key, result)

	if err != nil {
		log.Errorf("query err: %v", err)
	}
	return
}

// FindAndModifyDocument updates a document using mongo findAndModify
func FindAndModifyDocument(collectionPrefixedWithDB string, query bson.M, change mgo.Change) (result bson.M, info *mgo.ChangeInfo, err error) {

	log.Debugf("Query: %v", query)
	dBName, collectionName := getDBAndCollectionName(collectionPrefixedWithDB)

	session := GetSession()
	defer session.Close()

	c := session.DB(dBName).C(collectionName)
	q := c.Find(query)

	result = bson.M{}

	info, err = q.Apply(change, &result)
	if err != nil {
		return result, nil, err
	}

	log.Debugf("updated: %v, removed: %v, UpsertedId: %v", info.Updated, info.Removed, info.UpsertedId)
	return result, info, err
}

/*
  {
      "lastErrorObject": {
          "updatedExisting": true,
          "n": 1
      },
      "value": {
          "_id": "57de1cfde15f77045fdaa39a",
          "guid": "132203bc-e6dc-460f-b46d-5c9f34464730",
          "email": "yjin@appcelerator.com",
          "username": "yjin@appcelerator.com",
          "firstname": "Yuping",
          "orgs_360": [
              {
                  "id": "14301",
                  "name": "appcelerator Inc.",
                  "admin": true,
                  "node_acs_admin": true
              }
          ],
          "orgs_360_updated_at": "2016-10-08T08:27:05.520Z"
      },
      "ok": 1
  }
*/
func UpsertDocument(collectionPrefixedWithDB string, query, update interface{}) (result bson.M, info *mgo.ChangeInfo, err error) {

	dBName, collectionName := getDBAndCollectionName(collectionPrefixedWithDB)

	session := GetSession()
	defer session.Close()

	change := mgo.Change{
		Update:    update,
		Upsert:    true,
		ReturnNew: true,
	}

	result = bson.M{}

	c := session.DB(dBName).C(collectionName)

	//https://godoc.org/gopkg.in/mgo.v2#Query.Apply
	info, err = c.Find(query).Apply(change, &result)

	if err != nil {
		log.Errorf("findAndModify err: %v", err)
	} else {
		log.Debugf("updated: %v, removed: %v, UpsertedId: %v", info.Updated, info.Removed, info.UpsertedId)
	}

	return
}

func UpsertUser(collectionPrefixedWithDB string, query, update interface{}) (result models.User, info *mgo.ChangeInfo, err error) {

	dBName, collectionName := getDBAndCollectionName(collectionPrefixedWithDB)

	session := GetSession()
	defer session.Close()

	change := mgo.Change{
		Update:    update,
		Upsert:    true,
		ReturnNew: true,
	}

	result = models.User{}

	c := session.DB(dBName).C(collectionName)

	//https://godoc.org/gopkg.in/mgo.v2#Query.Apply
	info, err = c.Find(query).Apply(change, &result)

	if err != nil {
		log.Errorf("findAndModify err: %v", err)
	} else {
		log.Debugf("updated: %v, removed: %v, UpsertedId: %v", info.Updated, info.Removed, info.UpsertedId)
	}

	return
}

// RemoveDocument removes a document based on query
func RemoveDocument(collectionPrefixedWithDB string, query bson.M) (err error) {

	dBName, collectionName := getDBAndCollectionName(collectionPrefixedWithDB)

	session := GetSession()
	defer session.Close()

	c := session.DB(dBName).C(collectionName)

	//https://godoc.org/gopkg.in/mgo.v2#Query.Apply
	err = c.Remove(query)
	if err != nil {
		log.Errorf("remove err: %v", err)
	}
	return
}

// InsertDocument adds a new document
func InsertDocument(collectionPrefixedWithDB string, doc interface{}) (err error) {

	log.Debugf("Insert: %v, %v", collectionPrefixedWithDB, doc)
	dBName, collectionName := getDBAndCollectionName(collectionPrefixedWithDB)

	session := GetSession()
	defer session.Close()

	c := session.DB(dBName).C(collectionName)

	//https://godoc.org/gopkg.in/mgo.v2#Query.Apply
	err = c.Insert(doc)
	if err != nil {
		log.Errorf("insert err: %v", err)
	}
	return
}
