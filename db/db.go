package db


import (
    //"log"
    //"time"
    //"os"
    //"io"
    //"strings"
    //"fmt"
    //"net/http"
    //"html"
    //"encoding/json"

    "gopkg.in/mgo.v2"
    //"gopkg.in/mgo.v2/bson"

)

type DB struct{
	MongoSession *mgo.Session
	Name string
	Host string
}

var instance *DB = nil

func CreateInstance(host, name, user, pass string) *DB {
    if instance == nil {
        instance = &DB{Name:name, Host:host}

        sess, err := mgo.Dial(instance.Host)
        if err != nil{
            panic(err)
        }

        instance.MongoSession = sess

        if user != "" && pass != ""{
            instance.MongoSession.DB(instance.Name).Login(user, pass)
            
            if err != nil{
                panic(err)
            }
        }

    }
    return instance
}

func GetInstance() *DB {
    return instance
}

func GetDB() (*mgo.Database){
	db := GetInstance()
	return db.MongoSession.DB(db.Name)
}

func EnsureIndex(){
    db := GetDB()
    
    db.C("tweet").EnsureIndex(mgo.Index{Key: []string{"id_str"}, Unique: true, DropDups: true})
    db.C("tweet").EnsureIndex(mgo.Index{Key: []string{"created_at_unix"}})
}


// Tweets
func AddTweet(u map[string]interface{}) (map[string]interface{}, error){
	db := GetDB()

    err := db.C("tweet").Insert(u)

	return u, err
}


