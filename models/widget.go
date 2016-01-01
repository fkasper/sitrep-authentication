package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	widgetDbColumn = "widgets"
)

type Widget struct {
	Id    bson.ObjectId `bson:"_id,omitempty"`
	Title string        `bson:"title"`
	Css   string        `bson:"css,omitempty"`
	Js    string        `bson:"js,omitempty"`
	Html  string        `bson:"html"`
}
type Widgets []Widget

func PrepareQuery(sess *mgo.Database, column string) *mgo.Collection {
	return sess.C(column)
}

func GetWidgets(mongo *mgo.Database) interface{} {

	var widgets Widgets
	iter := PrepareQuery(mongo, widgetDbColumn).Find(nil).Iter()

	if err := iter.All(&widgets); err != nil {
		return err.Error()
	}

	if err := iter.Close(); err != nil {
		return err.Error()
	}
	return widgets
}

func UpsertWidget(mongo *mgo.Database, widget *Widget) interface{} {
	//objId := bson.NewObjectId()
	change, err := PrepareQuery(mongo, widgetDbColumn).Upsert(bson.M{"title": widget.Title}, widget)
	if err != nil {
		return err.Error()
		//return fmt.Errorf("An Error Occured error=%s", err)
	}

	return change
}

func DeleteWidget(mongo *mgo.Database, widgetId string) interface{} {
	objId := bson.ObjectIdHex(widgetId)
	if objId.Valid() {
		err := PrepareQuery(mongo, widgetDbColumn).RemoveId(objId)
		if err != nil {
			return err.Error()
		}
		return []string{}
	}
	return false
}
