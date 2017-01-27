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
    Radius int // in km (be carefully, radius > 5km will result in empty location info: https://twittercommunity.com/t/twitter-search-api-always-return-geo-null/66166/6)
    Maxid string // max id to continue the search from this id

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

type Location struct {
    Coordinates [2]float64
    Type string
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

func searchTweets(query url.Values) (*twittergo.SearchResults, *twittergo.APIResponse, error){
    
    url := fmt.Sprintf("/1.1/search/tweets.json?%v", query.Encode())
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, nil, err
    }
    resp, err := client.SendRequest(req)
    if err != nil {
        return nil, nil, err
    }
    results := &twittergo.SearchResults{}
    err = resp.Parse(results)
    if err != nil {
        return nil, nil, err
    }

    return results, resp, err
}


func checkRateLimits(resp *twittergo.APIResponse){
    log.Printf("Rate Limit: %d/%d, Rate Limit Reset: %d (%s)", resp.RateLimitRemaining(), resp.RateLimit(), resp.RateLimitReset().Unix(), resp.RateLimitReset().Format(time.RFC1123))
}

// Save recent tweets about something in a specific location with id greater than maxid until YYYY-MM-DD
// to save tweets later..
func saveTweets(q string, location [2]float64, radius int, seconds int, since, until, maxid string){
    geocode := fmt.Sprintf("%g,%g,%dkm", location[0], location[1], radius)

    query := url.Values{}
    query.Set("q", fmt.Sprintf("%s since:%s until:%s", q, since, until))
    query.Set("geocode", geocode)
    query.Set("count", "110")
    query.Set("result_type", "recent")

    if maxid != "" || maxid != "0"{
        query.Set("max_id", maxid)
    }

    for{
        // log.Printf("\n\nQuery: %v", query.Encode())
        search, resp, err := searchTweets(query)
        if err != nil{
            log.Fatal(err)
        }
        checkRateLimits(resp)

        tweets := search.Statuses()
        log.Printf("Got %d tweets", len(tweets))
        for _, t := range tweets{
            //fmt.Printf("\n\t%s", t.Text())

            if t["coordinates"] == nil{
                // sometimes it comes nil, so we put the location info
                t["coordinates"] = Location{
                    Coordinates: location,
                    Type: "Point",
                }
            }
            // Unix Timestamp for time
            t["created_at_unix"] = t.CreatedAt().Unix()

            db.AddTweet(t)
        }
        log.Printf(" -> Saved!")

        metadata := search.SearchMetadata()
        log.Printf("MaxID: %s", metadata["max_id_str"])

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
        search, resp, err := searchTweets(query)
        if err != nil{
            log.Fatal(err)
        }
        checkRateLimits(resp)

        tweets := search.Statuses()
        log.Printf("Got %d tweets", len(tweets))
        for _, t := range tweets{

            if t["coordinates"] == nil{
                // sometimes it comes nil, so we put the location info
                t["coordinates"] = Location{
                    Coordinates: location,
                    Type: "Point",
                }
            }
            // Unix Timestamp for time
            t["created_at_unix"] = t.CreatedAt().Unix()

            db.AddTweet(t)
        }
        log.Printf(" -> Saved!")
        
        metadata := search.SearchMetadata()
        log.Printf("MaxID: %s", metadata["max_id_str"])

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
        saveTweets(config.Query, config.Location, config.Radius, config.Seconds, config.Since, config.Until, config.Maxid)   
    }

}















