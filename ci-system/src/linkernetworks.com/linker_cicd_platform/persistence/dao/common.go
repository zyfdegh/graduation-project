package dao

import (
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/compose/mejson"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_cicd_platform/persistence/session"
)

const ParamID = "_id"                 // mongo id parameter
const ParamTimeCreate = "time_create" //time document created
const ParamTimeUpdate = "time_update" //time document last updated

type Dao struct {
	SessMng    *session.SessionManager
	MongoAlias string
}

func (d *Dao) ParamID() string {
	return ParamID
}

func (d *Dao) getCollection(collName string, session *mgo.Session) *mgo.Collection {
	config, err := d.SessMng.GetConfig(d.MongoAlias)
	if err != nil {
		logrus.Errorf("get db config error %+v, use default database ", err)
		return session.DB("").C(collName)
	}
	db := config.GetString("database", "")
	return session.DB(db).C(collName)
}

//
// Composes a query for finding documents
//
func (d *Dao) composeQuery(col *mgo.Collection, selector bson.M, one bool, fields bson.M, skip int, limit int, sort string) (query *mgo.Query, err error) {
	logrus.Debugf("compose query with params: selector=%v, one=%v, fields=%v, skip=%d, limit=%d, sort=%s", selector, one, fields, skip, limit, sort)
	// Create a Mongo Query
	query = col.Find(selector)

	// Fields of document to select
	if len(fields) > 0 {
		query.Select(fields)
	}

	// If selects one from `_id` parameter that's all
	if one {
		return
	}

	// Number of documents to skip in result set
	query.Skip(skip)

	// Maximum number of documents in the result set
	query.Limit(limit)

	// Compose sort from comma separated list in request query
	if len(sort) > 0 {
		query.Sort(strings.Split(sort, ",")...)
	}

	return
}

func (d *Dao) HandleInsert(collName string, selector bson.M, document bson.M) (id string, newDoc interface{}, err error) {
	// Mongo session
	session, needclose, err := d.SessMng.GetDefault()
	if err != nil {
		return
	}

	// Close session if it's needed
	if needclose {
		defer session.Close()
	}

	// Mongo Collection
	col := d.getCollection(collName, session)
	// Set document _id if not set
	if document[ParamID] == nil {
		// If id in selector use it
		if selector[ParamID] != nil {
			// Set document id from selector
			document[ParamID] = selector[ParamID]
			// Get string ID for content-location
			if hexid, ok := document[ParamID].(bson.ObjectId); ok {
				id = hexid.Hex()
			} else {
				id, _ = document[ParamID].(string)
			}
		} else {
			// Create new ObjectId
			docid := bson.NewObjectId()
			// Set new ID for document
			document[ParamID] = docid
			// Get string ID for content-location
			id = docid.Hex()
		}
	}

	//Add time_create if not exist
	document[ParamTimeCreate] = getCurrentTime()
	//Add time_update same as time_create at the first time
	document[ParamTimeUpdate] = document[ParamTimeCreate].(string)

	newDoc = document

	// Insert document to collection
	if err = col.Insert(newDoc); err != nil {
		logrus.Errorf("Error inserting document: %v", err)
		return
	}

	return
}

func (d *Dao) HandleUpdateById(collName string, selector bson.M, document bson.M) (docid string, newDoc interface{}, created bool, err error) {
	// Mongo session
	session, needclose, err := d.SessMng.GetDefault()
	if err != nil {
		return
	}

	// Close session if it's needed
	if needclose {
		defer session.Close()
	}

	// Mongo Collection
	col := d.getCollection(collName, session)

	// Update document(/s)
	var (
		info *mgo.ChangeInfo
	)

	// Trasform id to ObjectId if needed
	if id, _ := document[ParamID].(string); id != "" && bson.IsObjectIdHex(id) {
		document[ParamID] = bson.ObjectIdHex(id)
	}

	//Set time_update
	document[ParamTimeUpdate] = getCurrentTime()

	newDoc = document

	// Update document by id
	info, err = col.UpsertId(selector[ParamID], newDoc)
	if err != nil {
		return
	}

	// Get id from mongo
	if info != nil && info.UpsertedId != nil {
		docid, _ = info.UpsertedId.(string)
		created = (info.Updated == 0)
	}
	// Otherwise from selector
	if docid == "" {
		if id, ok := selector[ParamID].(string); ok {
			docid = id
		} else if id, ok := selector[ParamID].(bson.ObjectId); ok {
			docid = id.Hex()
		}
	}

	return
}

func (d *Dao) HandleUpdateByQuery(collName string, selector bson.M, document bson.M) (err error) {
	// Mongo session
	session, needclose, err := d.SessMng.GetDefault()
	if err != nil {
		return
	}

	// Close session if it's needed
	if needclose {
		defer session.Close()
	}

	// Mongo Collection
	col := d.getCollection(collName, session)

	//Set time_update
	document[ParamTimeUpdate] = getCurrentTime()

	// Trasform id to ObjectId if needed
	if id, _ := document[ParamID].(string); id != "" && bson.IsObjectIdHex(id) {
		document[ParamID] = bson.ObjectIdHex(id)
	}

	// Update all matching selector
	_, err = col.UpdateAll(selector, document)

	return
}

func (d *Dao) HandleUpdateByQueryPartial(collName string, selector bson.M, document bson.M) (err error) {
	// Mongo session
	session, needclose, err := d.SessMng.GetDefault()
	if err != nil {
		return
	}

	// Close session if it's needed
	if needclose {
		defer session.Close()
	}

	// Mongo Collection
	col := d.getCollection(collName, session)

	//Set time_update
	document[ParamTimeUpdate] = getCurrentTime()
	change := bson.M{"$set": document}
	err = col.Update(selector, change)

	return
}

func (d *Dao) HandleQuery(collName string, selector bson.M, one bool, fields bson.M, skip int, limit int, sort string, extended_json string) (total int, lenth int, jsonDocuments interface{}, err error) {
	// Mongo session
	session, needclose, err := d.SessMng.GetDefault()
	if err != nil {
		return
	}

	// Close session if it's needed
	if needclose {
		defer session.Close()
	}

	// Mongo Collection
	col := d.getCollection(collName, session)

	// Compose a query from request
	query, err := d.composeQuery(col, selector, one, fields, skip, limit, sort)
	if err != nil {
		return
	}

//	logrus.Debugf("collection=%v, query=%v", collName, query)

	// If _id parameter is included in path
	// 	queries only one document.
	// Get documents from database
	if one {
		// Get one document
		document := bson.M{}
		err = query.One(&document)
		if err != nil {
			return
		}
		if extended_json == "true" {
			jsonDocuments, err = mejson.Marshal(document)
			if err != nil {
				return
			}
		} else {
			jsonDocuments = document
		}
		total = 1
		lenth = 1
		return
	}

	// Get all documents
	documents := []bson.M{}
	err = query.All(&documents)
	if err != nil {
		return
	}

	if extended_json == "true" {
		jsonDocuments, err = mejson.Marshal(documents)
		if err != nil {
			return
		}
	} else {
		jsonDocuments = documents
	}

	// return founded result's length
	lenth = len(documents)

	// Count documents if count parameter is included in query
	query.Skip(0)
	query.Limit(0)
	if n, err := query.Count(); err == nil {
		total = n
	}

	return
}

func (d Dao) HandleDelete(collName string, one bool, selector bson.M) (err error) {
	// Mongo session
	session, needclose, err := d.SessMng.GetDefault()
	if err != nil {
		return
	}

	// Close session if it's needed
	if needclose {
		defer session.Close()
	}

	// Mongo Collection
	col := d.getCollection(collName, session)

	// If no selector at all - drop entire collection
	if len(selector) == 0 {
		err = col.DropCollection()
		if err != nil {
			return
		}
		return
	}

	// Remove one document if no query, otherwise remove all matching query
	if one {
		err = col.Remove(selector)
	} else {
		_, err = col.RemoveAll(selector)
	}

	return
}

//get current time as string formatted to RFC3339
func getCurrentTime() (t string) {
	t = time.Now().Format(time.RFC3339)
	return
}
