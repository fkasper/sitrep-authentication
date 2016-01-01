package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	appsDbColumn = "apps"
)

type App struct {
	Id        bson.ObjectId   `bson:"_id,omitempty"`
	Settings  string          `bson:"settings"`
	Name      string          `bson:"name,omitempty"`
	Customers []bson.ObjectId `bson:"customers,omitempty"`
	Domains   []bson.ObjectId `bson:"domains"`
}
type Apps []App

func GetApps(mongo *mgo.Database) interface{} {

	var apps Apps
	iter := PrepareQuery(mongo, appsDbColumn).Find(nil).Iter()

	if err := iter.All(&apps); err != nil {
		return err.Error()
	}

	if err := iter.Close(); err != nil {
		return err.Error()
	}
	return apps
}
