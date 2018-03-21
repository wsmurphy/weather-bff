package main

import (
	"log"
	"net/http"
	"os"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "time"

	"github.com/gin-gonic/gin"
    "github.com/russross/blackfriday"
)

const apiKey = "b4608d4fcb4accac0a8cc2ea6949eeb5"
var netClient = &http.Client{
    Timeout: time.Second * 20,
}

// Coordinates struct holds longitude and latitude data in returned
// JSON or as parameter data for requests using longitude and latitude.
type Coordinates struct {
    Longitude float64 `json:"lon"`
    Latitude  float64 `json:"lat"`
}

// Weather struct holds high-level, basic info on the returned
// data.
type Weather struct {
    ID          int    `json:"id"`
    Main        string `json:"main"`
    Description string `json:"description"`
    Icon        string `json:"icon"`
}

// Main struct contains the temperates, humidity, pressure for the request.
type Main struct {
    Temp      float64 `json:"temp"`
    TempMin   float64 `json:"temp_min"`
    TempMax   float64 `json:"temp_max"`
    Pressure  float64 `json:"pressure"`
    SeaLevel  float64 `json:"sea_level"`
    GrndLevel float64 `json:"grnd_level"`
    Humidity  int     `json:"humidity"`
}

type CurrentWeatherData struct {
    GeoPos  Coordinates `json:"coord"`
    Base    string      `json:"base"`
    Weather []Weather   `json:"weather"`
    Main    Main        `json:"main"`
    Dt      int         `json:"dt"`
    ID      int         `json:"id"`
    Name    string      `json:"name"`
    Cod     int         `json:"cod"`
    Unit    string
    Lang    string
    Key     string
}

type Dashboard struct {
    weatherConditions Weather
}

type Fact struct {
    IconUrl string
    Id      string
    Url     string
    Value   string
}

type AirQuality struct {
    Latitude    float64 `json:"lat"`
    Longitude   float64 `json:"long"`
    Date_iso    string  `json:"date_iso"`
    Date        int     `json:"date"`
    Value       float64 `json:"value"`
}

func GetWeather(ch chan<-string) {

    var weatherResponse CurrentWeatherData

    //Todo: take as parms
    location := "27013,US"
    units := "imperial"

    url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?zip=%s&units=%s&APPID=%s", location, units, apiKey)

    resp, _ := netClient.Get(url)

  body, _ := ioutil.ReadAll(resp.Body)
  err := json.Unmarshal(body, &weatherResponse)
  if err == nil {
      ch <- fmt.Sprintf("Temp: %f \nConditions: %s", weatherResponse.Main.Temp, weatherResponse.Weather[0].Description)
  } else {
    log.Output(1, "Error " + err.Error())
      ch <- fmt.Sprintf("Error unmarshalling response")
  }
}

func GetFact(ch chan<-string) {

    var factResponse Fact

    url := "https://api.chucknorris.io/jokes/random?category=dev"

    resp, _ := netClient.Get(url)

   body, _ := ioutil.ReadAll(resp.Body)
   err := json.Unmarshal(body, &factResponse)
   if err == nil {
       ch <- fmt.Sprintf("Fact : %s", factResponse.Value)
   } else {
    log.Output(1, "Error " + err.Error())
    ch <- fmt.Sprintf("Error unmarshalling response")
   }
}

func GetUVIndex(ch chan<-string) {
    var qualityResponse AirQuality

    //TODO: Get from weather call
    lat := 35.61
    long := -80.42    

    url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/uvi?lat=%f&lon=%f&APPID=%s", lat, long, apiKey)

    resp, _ := netClient.Get(url)

    body, _ := ioutil.ReadAll(resp.Body)
    err := json.Unmarshal(body, &qualityResponse)
    if err == nil {
        log.Output(1, fmt.Sprintf("Response: %f %f", lat, long))

        switch {
        case qualityResponse.Value < 3.0:
           ch <- fmt.Sprintf("UV Index : Green")
        case qualityResponse.Value < 6.0:
            ch <- fmt.Sprintf("UV Index : Yellow")
        case qualityResponse.Value < 8.0:
            ch <- fmt.Sprintf("UV Index : Orange")
        case qualityResponse.Value < 11.0:
            ch <- fmt.Sprintf("UV Index : Red")
        default:
            ch <- fmt.Sprintf("UV Index : Violet")
        }
    } else {
       log.Output(1, "Error " + err.Error())
       ch <- fmt.Sprintf("Error unmarshalling response")
    }
}

func dashboardHandler(c *gin.Context) {

    ch := make(chan string)
    ch2 := make(chan string)
    ch3 := make(chan string)

    go GetWeather(ch)
    go GetFact(ch2)
    go GetUVIndex(ch3)

    c.String(http.StatusOK, string(blackfriday.MarkdownBasic([]byte(fmt.Sprintf("%s\n%s\n%s", <-ch, <-ch2, <-ch3)))))
}

func getIndex(c *gin.Context) {
        c.HTML(http.StatusOK, "index.tmpl.html", nil)
    }

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", getIndex)
    router.GET("/dashboard", dashboardHandler)

	router.Run(":" + port)
}
