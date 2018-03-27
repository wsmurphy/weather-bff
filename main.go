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
)

const openWeatherAPIKey = "b4608d4fcb4accac0a8cc2ea6949eeb5"
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
    Humidity  int     `json:"humidity"`
}

type CurrentWeatherData struct {
    GeoPos  Coordinates `json:"coord"`
    Weather []Weather   `json:"weather"`
    Main    Main        `json:"main"`
    Dt      int         `json:"dt"`
    ID      int         `json:"id"`
    Name    string      `json:"name"`
    Cod     int         `json:"cod"`
}

type Fact struct {
    Value   string  `json:"value"`
}

type AirQuality struct {
    Value       float64 `json:"value"`
		StringValue string  `json:"stringValue"`
		ColorValue  string  `json:"colorValue"`
}

//JSON struct for response to the dashboard endpoint
type DashboardResponse struct {
    WeatherConditions CurrentWeatherData
		Fact Fact
		AirQuality AirQuality
}

//TODO: determine what info is needed and restruct the response to only the necessary info
func GetWeather(ch chan<-CurrentWeatherData, zip string) {

    var weatherResponse CurrentWeatherData

    //Todo: take as parms
    location := zip + ",US"
    units := "imperial"

    url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?zip=%s&units=%s&APPID=%s", location, units, openWeatherAPIKey)

    resp, _ := netClient.Get(url)

  body, _ := ioutil.ReadAll(resp.Body)
  err := json.Unmarshal(body, &weatherResponse)
  if err == nil {
      ch <- weatherResponse
  } else {
    log.Output(1, "Error " + err.Error())
  }
}

//TODO: Only pass back the info from this response that is actually needed
func GetFact(ch chan<-Fact) {

    var factResponse Fact

    url := "https://api.chucknorris.io/jokes/random?category=dev"

    resp, _ := netClient.Get(url)

   body, _ := ioutil.ReadAll(resp.Body)
   err := json.Unmarshal(body, &factResponse)
   if err == nil {
       ch <- factResponse
   } else {
    log.Output(1, "Error " + err.Error())
   }
}

func GetUVIndex(ch chan<-AirQuality, lat float64, long float64) {
    var qualityResponse AirQuality

    url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/uvi?lat=%f&lon=%f&APPID=%s", lat, long, openWeatherAPIKey)

    resp, _ := netClient.Get(url)

    body, _ := ioutil.ReadAll(resp.Body)
    err := json.Unmarshal(body, &qualityResponse)
    if err == nil {

		//Map color into response. Business logic should be in this layer, not in the app that calls it
        switch {
        case qualityResponse.Value < 3.0:
						qualityResponse.StringValue = "Low"
						qualityResponse.ColorValue = "Green"
        case qualityResponse.Value < 6.0:
					qualityResponse.StringValue = "Moderate"
					qualityResponse.ColorValue = "Yellow"
        case qualityResponse.Value < 8.0:
					qualityResponse.StringValue = "High"
					qualityResponse.ColorValue = "Orange"
        case qualityResponse.Value < 11.0:
					qualityResponse.StringValue = "Very High"
					qualityResponse.ColorValue = "Red"
				case qualityResponse.Value >= 11.0:
					qualityResponse.StringValue = "Extreme"
					qualityResponse.ColorValue = "Violet"
        default:
					qualityResponse.StringValue = "Unknown"
					qualityResponse.ColorValue = "Blue"
        }

				ch <- qualityResponse
    } else {
       log.Output(1, "Error " + err.Error())
    }
}

func dashboardHandler(c *gin.Context) {
		zip := c.Query("zip")

    ch := make(chan CurrentWeatherData)
    ch2 := make(chan Fact)
    ch3 := make(chan AirQuality)

    go GetWeather(ch, zip)
    go GetFact(ch2)

		var weatherResponse = <-ch
		var lat = weatherResponse.GeoPos.Latitude
		var long = weatherResponse.GeoPos.Longitude

    go GetUVIndex(ch3, lat, long)


		var factResponse = <-ch2
		var uviResponse = <-ch3

    //TODO: Refine the DashboardResponse to only what the UI needs
		respJSON := DashboardResponse{WeatherConditions: weatherResponse,
																	Fact: factResponse,
																	AirQuality: uviResponse}

	 //TODO: Error handling if one of the responses is nil
    c.JSON(http.StatusOK, respJSON)
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

	router.GET("/", getIndex)
  router.GET("/dashboard", dashboardHandler)

	router.Run(":" + port)
}
