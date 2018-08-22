package db

import (
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

type Database struct {
	session *mgo.Session
	latch   sync.Pool
}

var (
	once        sync.Once
	_default_db Database
)

func InitMongo(hosts string, dur time.Duration, concurrent int) {
	once.Do(func() {
		_default_db.Init(hosts, dur, concurrent)
	})
}

func (db *Database) Init(hosts string, dur time.Duration, concurrent int) {
	// connect db
	sess, err := mgo.Dial(hosts)
	if err != nil {
		log.Errorf("mongodb: cannot connect to %v, %v", hosts, err)
		os.Exit(-1)
	}

	// set params
	sess.SetMode(mgo.Strong, true)
	sess.SetSocketTimeout(dur)
	sess.SetCursorTimeout(0)
	db.session = sess

	db.latch = sync.Pool{
		New: func() interface{} {
			return db.session.Copy()
		},
	}
}

func (db *Database) Execute(f func(sess *mgo.Session) error) error {
	// latch control
	sess := db.latch.Get().(*mgo.Session)
	defer func() {
		db.latch.Put(sess)
	}()

	sess.Refresh()
	return f(sess)
}

func Execute(f func(sess *mgo.Session) error) error {
	return _default_db.Execute(f)
}
