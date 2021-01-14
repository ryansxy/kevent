package store

import (
	"errors"
	"fmt"
	"log"
	"upper.io/db.v3"
	"upper.io/db.v3/mongo"
)

type MongoConfig = mongo.ConnectionURL

type MongodbStore struct {
	Config *MongoConfig
}

var DefaultMongoStore *MongodbStore

func InitMongodbStore(config *MongoConfig) (err error) {
	if config == nil || config.Host == "" || config.Database == "" {
		err = errors.New("invalid mongodb configuration")
		return
	}
	DefaultMongoStore = &MongodbStore{
		Config: config,
	}
	return
}

func (m *MongodbStore) Insert(colletionname string, obj interface{}) error {

	sess, err := mongo.Open(m.Config)
	if err != nil {
		log.Fatalf("db.Open(): %q\n", err)
		return errors.New(fmt.Sprintf("db.Open(): %q\n", err))
	}
	defer sess.Close() // Remember to close the database session.

	// Pointing to the "birthday" table.
	Collection := sess.Collection(colletionname)

	_, err = Collection.Insert(obj)
	if nil != err {
		return err
	}
	return nil
}

// according to condmap  to find the res.
func (m *MongodbStore) Delete(colletionname string, condmap map[interface{}]interface{}) error {
	sess, err := mongo.Open(m.Config)
	if err != nil {
		log.Fatalf("db.Open(): %q\n", err)
		return errors.New(fmt.Sprintf("db.Open(): %q\n", err))
	}
	defer sess.Close() // Remember to close the database session.

	// Pointing to the "birthday" table.
	Collection := sess.Collection(colletionname)

	res := Collection.Find(condmap)
	// Trying to remove the row.
	err = res.Delete()

	if err != nil {
		log.Printf("db.delete() error: %s\n", err)
	}

	return nil
}

func (m *MongodbStore) Find(colletionname string, condmap map[interface{}]interface{}, dest interface{}) (res db.Result, err error) {
	/*	if condmap == nil || len(condmap) == 0 {
		err = errors.New("invalid query conditions")
		return
	}*/
	sess, err := mongo.Open(m.Config)
	if err != nil {
		//log.Fatalf("db.Open(): %q\n", err)
		log.Printf("db.Open(): %s\n", err)
		err = errors.New(fmt.Sprintf("db.Open(): %q\n", err))
		return
	}
	defer sess.Close() // Remember to close the database session.

	// Pointing to the colletionname table.
	Collection := sess.Collection(colletionname)

	// Let's query for the results we've just inserted.

	query := make(map[interface{}]interface{})
	for k, v := range condmap {
		query[k] = v
	}
	res = Collection.Find(db.Cond(query))
	err = res.All(dest)
	if err != nil {
		log.Printf("res.All(): %q\n", err)
		err = errors.New(fmt.Sprintf("res.All(): %q\n", err))
		return
	}
	return
}
