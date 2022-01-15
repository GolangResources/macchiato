# macchiato

Ristretto compatible library designed to use mongo as backend. Is not designed for high performance, just to persist data when you work with ephemeral environments (like Heroku or Google App Engine).

In the future can be great include ristretto to get both functions, but for now it's up to you do the right use :D.

## Example

```
package main

import (
        "os"
        "log"
        "github.com/GolangResources/macchiato"
)

var cache macchiato.Cache

func main() {
        cache, err := macchiato.NewCache(&macchiato.Config{
                MongoURI: os.Getenv("MONGO_URI"),
                Database: "macchiato",
                Collection: "cache",
        })
        if err != nil {
                panic(err)
        }
	defer cache.Disconnect()

        log.Println("SET: ", cache.Set("TEST", "a", 1))

        result, found := cache.Get("TEST")
        if found {
                log.Println("GET: " + result.(string))
        }

        log.Println("DEL: ", cache.Del("TEST"))
}
```

## How to set TTL

```
db.cache.createIndex( { "id": 1 }, { unique:true, expireAfterSeconds: 172800 } )
```
