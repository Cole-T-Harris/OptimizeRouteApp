package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/supabase-community/supabase-go"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	UserID      *int    `json:"user_id"`
	Origin      *string `json:"origin_address"`
	Destination *string `json:"destination_address"`
	Timezone    *string `json:"timezone"`
}

type Response struct {
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

type Data struct {
	AddedPlaces AddedPlaces `json:"added_routes"`
}

type Place struct {
	LatLng  Coordinates `json:"coordinates"`
	Address string      `json:"address"`
}

type AddedPlaces struct {
	Origin      Place `json:"origin"`
	Destination Place `json:"destination"`
}

type Coordinates struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Geometry struct {
	Location Location `json:"location"`
}

type GeoCodeResult struct {
	Geometry         Geometry `json:"geometry"`
	FormattedAddress string   `json:"formatted_address"`
}

type GoogleResponse struct {
	GeoCodeResults []GeoCodeResult `json:"results"`
}

type Route struct {
	UserID               int    `json:"user_id"`
	Origin               string `json:"start_address"`
	Destination          string `json:"end_address"`
	Timezone             string `json:"time_zone"`
	Active               bool   `json:"active"`
	StartDate            string `json:"start_date"`
	EndDate              string `json:"end_date"`
	OriginLatitude       string `json:"start_latitude"`
	OriginLongitude      string `json:"start_longitude"`
	DestinationLatitude  string `json:"end_latitude"`
	DestinationLongitude string `json:"end_longitude"`
}

var googleMapsAPIKey = os.Getenv("GOOGLE_API_KEY")
var supabaseURL = os.Getenv("SUPABASE_URL")
var supabaseKey = os.Getenv("SUPABASE_KEY")
var endDateBuffer = 30 //30 days from now route will become inactive

func getCoordinates(address *string) (Place, error) {
	if address == nil {
		return Place{}, fmt.Errorf("invalid address, address not included or nil")
	}
	formattedString := *address
	// Replace commas with plus signs
	formattedString = strings.ReplaceAll(formattedString, ",", "+")
	// Replace spaces with plus signs
	formattedString = strings.ReplaceAll(formattedString, " ", "+")
	getUrl := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?address=%s&key=%s", formattedString, googleMapsAPIKey)
	response, err := http.Get(getUrl)
	if err != nil {
		fmt.Println("Error sending GET request:", err)
		return Place{}, fmt.Errorf("error sending GET request: %s", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return Place{}, fmt.Errorf("error reading response body: %s", err)
	}

	var apiResponse GoogleResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return Place{}, fmt.Errorf("error unmarshalling JSON: %s", err)
	}
	// Verify length of results
	if len(apiResponse.GeoCodeResults) != 1 {
		fmt.Println("Expected 1 result, got:", len(apiResponse.GeoCodeResults))
		return Place{}, fmt.Errorf("expected 1 result, got: %d", len(apiResponse.GeoCodeResults))
	}
	// Check if geometry.location.lat and geometry.location.lng are present
	firstResult := apiResponse.GeoCodeResults[0]
	if firstResult.Geometry.Location.Lat == 0 || firstResult.Geometry.Location.Lng == 0 {
		fmt.Println("Latitude or Longitude is missing or zero.")
		return Place{}, fmt.Errorf("latitude or Longitude is missing or zero")
	}
	resultingPlace := Place{
		LatLng: Coordinates{
			Latitude:  strconv.FormatFloat(firstResult.Geometry.Location.Lat, 'f', -1, 64),
			Longitude: strconv.FormatFloat(firstResult.Geometry.Location.Lng, 'f', -1, 64),
		},
		Address: firstResult.FormattedAddress,
	}
	return resultingPlace, nil
}

func HandleRequest(ctx context.Context, request Request) (Response, error) {
	if request.UserID == nil {
		return Response{}, fmt.Errorf("invalid request. Missing user ID")
	}
	if request.Origin == nil {
		return Response{}, fmt.Errorf("invalid request. Missing origin_address")
	}
	if request.Destination == nil {
		return Response{}, fmt.Errorf("invalid request. Missing destination_address")
	}
	if request.Timezone == nil {
		return Response{}, fmt.Errorf("invalid request. Missing timezone")
	}
	// Validate timezone string
	userLocation, err := time.LoadLocation(*request.Timezone)
	if err != nil {
		return Response{}, fmt.Errorf("invalid timezone: %s", request.Timezone)
	}
	databaseClient, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		fmt.Println("cannot initalize client", err)
	}
	originPlace, err := getCoordinates(request.Origin)
	if err != nil {
		fmt.Println("Error obtaining origin location coordinates:", err)
		return Response{}, fmt.Errorf("error obtaining location coordinates: %s", err)
	}
	destinationPlace, err := getCoordinates(request.Destination)
	if err != nil {
		fmt.Println("Error obtaining destination location coordinates:", err)
		return Response{}, fmt.Errorf("error obtaining location coordinates: %s", err)
	}
	newRoute := Route{
		UserID:               *request.UserID,
		Origin:               originPlace.Address,
		OriginLatitude:       originPlace.LatLng.Latitude,
		OriginLongitude:      originPlace.LatLng.Longitude,
		Destination:          destinationPlace.Address,
		DestinationLatitude:  destinationPlace.LatLng.Latitude,
		DestinationLongitude: destinationPlace.LatLng.Longitude,
		Timezone:             *request.Timezone,
		Active:               true,
		StartDate:            time.Now().In(userLocation).Format("2006-01-02"),
		EndDate:              time.Now().In(userLocation).AddDate(0, 0, endDateBuffer).Format("2006-01-02"),
	}
	upsert := false
	response, rowsEffected, err := databaseClient.From("routes").Insert(newRoute, upsert, "", "*", "exact").Execute()
	if err != nil {
		fmt.Printf("Failed to insert data: %v", err)
		return Response{}, fmt.Errorf("failed to insert data: %v", err)
	}

	if rowsEffected != 1 {
		fmt.Printf("Incorrect of rows effected, rows effected: %d", rowsEffected)
		return Response{}, fmt.Errorf("incorrect of rows effected, rows effected: %d", rowsEffected)
	}
	fmt.Println("Response: ", response)
	addUserRouteResponse := Response{
		Message: "Success",
		Data: Data{
			AddedPlaces: AddedPlaces{
				Origin:      originPlace,
				Destination: destinationPlace,
			},
		},
	}
	return addUserRouteResponse, nil
}

func main() {
	lambda.Start(HandleRequest)
}
