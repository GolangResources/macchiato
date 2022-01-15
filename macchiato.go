package macchiato

import (
	"os"
	"bytes"
	"errors"
	"strings"
	"context"
	"reflect"
	"encoding/gob"
        "go.mongodb.org/mongo-driver/bson"
        "go.mongodb.org/mongo-driver/mongo"
        "go.mongodb.org/mongo-driver/mongo/options"
)

type Cache struct {
	client *mongo.Client
	collection *mongo.Collection
}

type Config struct {
	MongoURI string
	Database string
	Collection string
}

type CacheCast struct {
	Interface interface{}
}

type CacheDB struct {
        ID string `bson:"id"`
        Content []byte `bson:"content"`
        Type string `bson:"type"`
}

func NewCache(config *Config) (Cache, error) {
	var err error
	var database string
	var collection string
	var cache Cache

	//We will register in gob our fake struct
	gob.Register(CacheCast{})

	if (config.MongoURI == "") {
		err = errors.New("Mongo URI cannot be empty")
	}

	//If Database is null, we will use "macchiato" as default
	if (config.Database == "") {
		database = "macchiato"
	} else {
		database = config.Database
	}

	//If Collection is null, we will use "cache" as default
	if (config.Collection == "") {
		collection = "cache"
	} else {
		collection = config.Collection
	}

	//Preparing client
        clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
        cache.client, err = mongo.Connect(context.TODO(), clientOptions)
        if err != nil {
		return Cache{}, err
        }

	//We will test now the DB
        err = cache.client.Ping(context.TODO(), nil)
        if err != nil {
		return Cache{}, err
        }

	//Now we will set the collection
        cache.collection = cache.client.Database(database).Collection(collection)

	return &cache, nil
}

func (c *Cache) Disconnect() error {
	err := c.client.Disconnect(context.TODO())

	return err
}

func (c *Cache) Get(s string) (interface{}, bool) {
        var found bool
        var err error
        var b bytes.Buffer
        var result interface{}
        var resultDB CacheDB

        err = c.collection.FindOne(context.TODO(), bson.M{ "id": s }).Decode(&resultDB)
        if err != nil {
                if ! (strings.Contains(err.Error(), "no documents")) {
			return nil, false
                }
        }

        if (resultDB.ID != "") {
                found = true

                b.Write(resultDB.Content)
                d := gob.NewDecoder(&b)

		var resultTmp CacheCast

		err = d.Decode(&resultTmp)
                if err != nil {
			return nil, false
                }

		result = resultTmp.Interface
        }

        return result, found
}

func (c *Cache) Set(s string, i interface{}, n int) (error) {
        var b bytes.Buffer
        var resultDB CacheDB
        var err error

        e := gob.NewEncoder(&b)
        e.Encode(CacheCast{ Interface: i})

        if err != nil {
		return err
        }

        resultDB.ID = s
        resultDB.Type = reflect.TypeOf(i).String()
	resultDB.Content = b.Bytes()

        upFilter := bson.M{
                "$and": bson.A{
                        bson.M{
                                "id": bson.M{
                                        "$eq": resultDB.ID,
                                },
                        },
                },
        }

        upMsg := bson.M{
                "$set": resultDB,
        }

        _, err = c.collection.UpdateMany(context.TODO(), upFilter, upMsg, options.Update().SetUpsert(true))

        return err
}

func (c *Cache) Del(s string) (error) {
        _, err := c.collection.DeleteMany(context.TODO(), bson.M{ "id": s })
        if err != nil {
		return err
        }

        return err
}
