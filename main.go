package main

import (
    "log"
    "fmt"
    "encoding/json"
    "io/ioutil"
    "time"
    "net/url"
    "net/http"

    "github.com/kurrik/oauth1a"
    "github.com/kurrik/twittergo"

    "TweetsSaver/db"
)

// MyConfig struct: for the user config to save tweets
// This is the struct that the config.json must have
type MyConfig struct {
    ConsumerKey string
    ConsumerSecret string

    AccessToken string
    AccessSecret string

    Query string
    Since string // YYYY-MM-DD
    Until string // YYYY-MM-DD
    Location [2]float64 // longitude, latitude
    Radius int // in km

    Seconds int // seconds between requests (remember you have 350 request per hour limit!)

    SaveType int
        // 0: saveRecentTweets (search and save every X seconds)
        // 1: saveTweets (search and save ALL tweets from SINCE until UNTIL making requests every X seconds)

    // DB info
    DbConfig MyDBConfig

}
type MyDBConfig struct {
    Host string
    User string
    Pass string
    Name string // DB name
}

var (
    client *twittergo.Client
)

func LoadConfig(filename string) (MyConfig, error){
    var s MyConfig

    bytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return s, err
    }
    // Unmarshal json
    err = json.Unmarshal(bytes, &s)
    return s, err
}

func searchTweets(query url.Values) (*twittergo.SearchResults, error){
    
    url := fmt.Sprintf("/1.1/search/tweets.json?%v", query.Encode())
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    resp, err := client.SendRequest(req)
    if err != nil {
        return nil, err
    }
    results := &twittergo.SearchResults{}
    err = resp.Parse(results)
    if err != nil {
        return nil, err
    }

    return results, err
}


// Save recent tweets about something in a specific location with id greater than maxid until YYYY-MM-DD
// to save tweets later..
func saveTweets(q string, location [2]float64, radius int, seconds int, since, until string){
    geocode := fmt.Sprintf("%g,%g,%dkm", location[0], location[1], radius)

    query := url.Values{}
    query.Set("q", fmt.Sprintf("%s since:%s until:%s", q, since, until))
    query.Set("geocode", geocode)
    query.Set("count", "110")
    query.Set("result_type", "recent")

    for{
        fmt.Printf("\n\nQuery: %v", query.Encode())
        search, err := searchTweets(query)
        if err != nil{
            log.Fatal(err)
        }

        tweets := search.Statuses()
        fmt.Printf("\nGot %d tweets", len(tweets))
        for _, t := range tweets{
            //fmt.Printf("\n\t%s", t.Text())

            if t["coordinates"] == nil{
                // sometimes it comes nil, so we put the location info
                t["coordinates"] = [2]float64{location[0], location[1]}
            }
            // Unix Timestamp for time
            t["created_at_unix"] = t.CreatedAt().Unix()

            //fmt.Printf("\n%s", t["coordinates"])
            //fmt.Printf("at %v\n\n", t.CreatedAt().Unix())

            db.AddTweet(t)
        }
        fmt.Printf(" -> Saved!")

        query, err = search.NextQuery() // next page
        if err != nil {
            log.Fatal(err)
        }

        // sleep
        time.Sleep(time.Duration(seconds) * time.Second)
    }
}

// Save recent tweets about something in a specific location every X seconds
func saveRecentTweets(q string, location [2]float64, radius int, seconds int){
    geocode := fmt.Sprintf("%g,%g,%dkm", location[0], location[1], radius)

    query := url.Values{}
    query.Set("q", q)
    query.Set("geocode", geocode)
    query.Set("count", "110")
    query.Set("result_type", "recent")

    for{
        search, err := searchTweets(query)
        if err != nil{
            log.Fatal(err)
        }

        tweets := search.Statuses()
        fmt.Printf("\nGot %d tweets", len(tweets))
        for _, t := range tweets{
            //fmt.Printf("\n\t%s", t.Text())

            if t["coordinates"] == nil{
                // sometimes it comes nil, so we put the location info
                t["coordinates"] = [2]float64{location[0], location[1]}
            }
            // Unix Timestamp for time
            t["created_at_unix"] = t.CreatedAt().Unix()

            //fmt.Printf("\n%s", t["coordinates"])
            //fmt.Printf("at %v\n\n", t.CreatedAt().Unix())

            db.AddTweet(t)
        }
        fmt.Printf(" -> Saved!")

        // sleep
        time.Sleep(time.Duration(seconds) * time.Second)
    }
}

func main(){

    config, err := LoadConfig("./config.json")
    if err != nil{
        fmt.Printf("Error loading config [%s]\n", err)
        return
    }

    oauthConfig := &oauth1a.ClientConfig{
        ConsumerKey:    config.ConsumerKey,
        ConsumerSecret: config.ConsumerSecret,
    }
    user := oauth1a.NewAuthorizedConfig(config.AccessToken, config.AccessSecret)

    // Twitter client
    client = twittergo.NewClient(oauthConfig, user)

    // Prepare DB connection
    db.CreateInstance(config.DbConfig.Host, config.DbConfig.Name, config.DbConfig.User, config.DbConfig.Pass)
    db.EnsureIndex()

    if config.SaveType == 0{
        saveRecentTweets(config.Query, config.Location, config.Radius, config.Seconds)
    }else if config.SaveType == 1{
        saveTweets(config.Query, config.Location, config.Radius, config.Seconds, config.Since, config.Until)   
    }

}















