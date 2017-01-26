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
)

// MyConfig struct: for the user config to save tweets
// This is the struct that the config.json must have
type MyConfig struct {
    ConsumerKey string `json:"consumerKey"`
    ConsumerSecret string

    AccessToken string
    AccessSecret string

    Query string `json:"query"`
    Since string // YYYY-MM-DD
    Until string // YYYY-MM-DD
    Location [2]float64 // longitude, latitude
    Radius int // in km

    Seconds int // seconds between requests (remember you have 350 request per hour limit!)

    SaveType int
        // 0: saveRecentTweets (search and save every X seconds)
        // 1: saveTweets (search and save ALL tweets from SINCE until UNTIL making requests every X seconds)
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
    fmt.Printf("%s\n", string(bytes))
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
        fmt.Printf("Query: %v", query.Encode())
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
            //fmt.Printf("\n%s", t["coordinates"])
            //fmt.Printf("at %v\n\n", t.CreatedAt().Format(time.RFC1123))
        }

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
        fmt.Printf("Query: %v", query.Encode())
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
            //fmt.Printf("\n%s", t["coordinates"])
            //fmt.Printf("at %v\n\n", t.CreatedAt().Format(time.RFC1123))
        }

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
    fmt.Printf("%v", config)

    oauthConfig := &oauth1a.ClientConfig{
        ConsumerKey:    config.ConsumerKey,
        ConsumerSecret: config.ConsumerSecret,
    }
    user := oauth1a.NewAuthorizedConfig(config.AccessToken, config.AccessSecret)
    //oauth1.NewConfig(config["consumerKey"], config["consumerSecret"])
    //token := oauth1.NewToken(config["accessToken"], config["accessSecret"])
    //httpClient := oauthConfig.Client(oauth1.NoContext, token)

    // Twitter client
    client = twittergo.NewClient(oauthConfig, user) //twitter.NewClient(httpClient)


    if config.SaveType == 0{
        saveRecentTweets(config.Query, config.Location, config.Radius, config.Seconds)
    }else if config.SaveType == 0{
        saveTweets(config.Query, config.Location, config.Radius, config.Seconds, config.Since, config.Until)   
    }

}















