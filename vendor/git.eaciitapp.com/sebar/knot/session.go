package knot

import (
	"time"

	"git.eaciitapp.com/sebar/dbflex"
	"github.com/eaciit/toolkit"
)

var sessionTableName = "sessions"

func (s *sessionStore) setSessionDuration(d time.Duration) {
	s.duration = d
}

func (s *sessionStore) sessionDuration() time.Duration {
	if s.duration == 0 {
		s.duration = 90 * time.Minute
	}
	return s.duration
}

type sessionStore struct {
	connection dbflex.IConnection
	data       *SharedData
	duration   time.Duration
}

type sessionItem struct {
	key        string
	data       *SharedData
	lastAccess time.Time
}

type sessionItemModel struct {
	Key        string `bson:"_id" json:"_id" sql:"_id"`
	Value      interface{}
	LastAccess time.Time
}

func newSessionStore() *sessionStore {
	s := new(sessionStore)
	s.initStorage()
	return s
}

func (s *sessionStore) initStorage() {
	s.data = NewSharedData()
}

func (s *sessionStore) close() {
	if s.connection != nil {
		s.connection.Close()
	}
}

func (s *sessionStore) set(cookieid, key string, value interface{}) error {
	if s.connection == nil {
		var sessData *sessionItem
		sessObj := s.data.Get(cookieid, nil)
		if sessObj == nil {
			sessData = new(sessionItem)
			sessData.key = cookieid
			sessData.data = NewSharedData()
		} else {
			sessData = sessObj.(*sessionItem)
		}
		sessData.data.Set(key, value)
		sessData.lastAccess = time.Now()
		s.data.Set(cookieid, sessData)
		return nil
	}

	cmd := dbflex.From(sessionTableName).Save()

	item := sessionItemModel{}
	item.Key = toolkit.Sprintf("%s_%s", cookieid, key)
	item.Value = value
	item.LastAccess = time.Now()

	_, err := s.connection.Execute(cmd, toolkit.M{}.Set("data", item))
	if err != nil {
		return err
	}
	return nil
}

func (s *sessionStore) get(cookieid, key string, def interface{}) interface{} {
	if s.connection == nil {
		sessObj := s.data.Get(cookieid, nil)

		if sessObj == nil {
			return def
		}

		sessData := sessObj.(*sessionItem)
		if time.Now().Sub(sessData.lastAccess) > s.sessionDuration() {
			sessData.data.Remove(key)
			return def
		}

		sessData.lastAccess = time.Now()
		return sessData.data.Get(key, def)
	}

	_id := toolkit.Sprintf("%s_%s", cookieid, key)
	cmd := dbflex.From(sessionTableName).Select().Where(dbflex.Eq("_id", _id))
	cursor := s.connection.Cursor(cmd, nil)
	if cursor.Error() != nil {
		return def
	}

	var item sessionItemModel
	err := cursor.Fetch(&item)
	if err != nil {
		return def
	}

	if item.Value == nil {
		return def
	}

	if time.Now().Sub(item.LastAccess) > s.sessionDuration() {
		cmd = dbflex.From(sessionTableName).Delete().Where(dbflex.Eq("_id", _id))
		s.connection.Execute(cmd, nil)
		return def
	}

	return item.Value
}
