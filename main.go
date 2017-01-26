package main

import (
    "log"
    "fmt"
    "encoding/json"
    "io/ioutil"
    "time"

    "github.com/dghubble/go-twitter/twitter"
    "github.com/dghubble/oauth1"
)

var (
    client *twitter.Client
)

func LoadJSONTemplate(filename string) (map[string]string, error){
    s := make(map[string]string, 0)
 
    bytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return s, err
    }
    // Unmarshal json
    err = json.Unmarshal(bytes, &s)
    return s, err
}

// Save recent tweets about something in a specific location with id greater than maxid until YYYY-MM-DD
// to save tweets later..
func saveTweets(query string, location [2]float64, radius int, seconds int, until string){
    geocode := fmt.Sprintf("%g,%g,%dkm", location[0], location[1], radius)

    // Get last tweet of the previous day
    search, _, _ := client.Search.Tweets(&twitter.SearchTweetParams{
        Query: query,
        Geocode: geocode,
        Count: 5,
        ResultType: "recent",
        Until: until,
    })

    // use this max id to get tweets from the next day
    var maxid int64 = search.Metadata.MaxID // next page
    for{
        search, _, err := client.Search.Tweets(&twitter.SearchTweetParams{
            Query: query,
            Geocode: geocode,
            Count: 500,
            ResultType: "recent",
            Until: until,
            MaxID: maxid,
        })

        if err != nil{
            log.Fatal(err)
        }

        tweets := search.Statuses
        metadata := search.Metadata
        fmt.Printf("\nGot %d tweets", len(tweets))
        for _, t := range tweets{
            fmt.Printf("\n\t%s", t.Text)

            if t.Coordinates == nil{
                // sometimes it comes nil, so we put the location info
                t.Coordinates = &twitter.Coordinates{Coordinates:location, Type:"Point"}
            }
        }

        maxid = metadata.MaxID // next page

        // sleep
        time.Sleep(time.Duration(seconds) * time.Second)
    }
}

// Save recent tweets about something in a specific location every X seconds
func saveRecentTweets(query string, location [2]float64, radius int, seconds int){
    geocode := fmt.Sprintf("%g,%g,%dkm", location[0], location[1], radius)

    for{
        search, _, err := client.Search.Tweets(&twitter.SearchTweetParams{
            Query: query,
            Geocode: geocode,
            Count: 500,
            ResultType: "recent",
        })

        if err != nil{
            log.Fatal(err)
        }

        tweets := search.Statuses
        fmt.Printf("\nGot %d tweets", len(tweets))
        for _, t := range tweets{
            fmt.Printf("\n\t%s", t.Text)

            if t.Coordinates == nil{
                // sometimes it comes nil, so we put the location info
                t.Coordinates = &twitter.Coordinates{Coordinates:location, Type:"Point"}
            }
        }

        // sleep
        time.Sleep(time.Duration(seconds) * time.Second)
    }
}


func main(){

    config, err := LoadJSONTemplate("./config.json")
    if err != nil{
        fmt.Printf("Error loading config [%s]\n", err)
        return
    }

    oauthConfig := oauth1.NewConfig(config["consumerKey"], config["consumerSecret"])
    token := oauth1.NewToken(config["accessToken"], config["accessSecret"])
    httpClient := oauthConfig.Client(oauth1.NoContext, token)

    // Twitter client
    client = twitter.NewClient(httpClient)

    //saveRecentTweets("#hi", [2]float64{40.415178, -3.703697}, 50, 15)
    saveTweets("#halamadrid", [2]float64{40.415178, -3.703697}, 50, 15, "2017-01-25")

}















