package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"io"
	"os"
	"strings"
	"time"
)

type Request struct {
	UserID      *int    `json:"user_id"`
	Origin      *string  `json:"origin_address"`
	Destination *string  `json:"destination_address"`
	Timezone    *string `json:"timezone"` 
}

type Response struct {
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

type Data struct {
	Status    string `json:"status"`
	AddedRoutes  int `json:"added_routes"`
}

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Location struct {
    Lat float64 `json:"lat"`
    Lng float64 `json:"lng"`
}

type Geometry struct {
    Location Location `json:"location"`
}

type GeoCodeResult struct {
    Geometry Geometry `json:"geometry"`
}

type GoogleResponse struct {
    GeoCodeResults []GeoCodeResult `json:"results"`
}

var googleMapsAPIKey = os.Getenv("GOOGLE_API_KEY")

func getCoordinates(address *string) (Coordinates, error) {
	if address == nil {
		return Coordinates{}, fmt.Errorf("invalid address, address not included or nil")
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
        return Coordinates{}, fmt.Errorf("error sending GET request: %s", err)
    }
    defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
    if err != nil {
        fmt.Println("Error reading response body:", err)
        return Coordinates{}, fmt.Errorf("error reading response body: %s", err)
    }

	var apiResponse GoogleResponse
    err = json.Unmarshal(body, &apiResponse)
    if err != nil {
        fmt.Println("Error unmarshalling JSON:", err)
        return Coordinates{}, fmt.Errorf("error unmarshalling JSON: %s", err)
    }
    // Verify length of results
    if len(apiResponse.GeoCodeResults) != 1 {
        fmt.Println("Expected 1 result, got:", len(apiResponse.GeoCodeResults))
        return Coordinates{}, fmt.Errorf("expected 1 result, got: %d", len(apiResponse.GeoCodeResults))
    }
    // Check if geometry.location.lat and geometry.location.lng are present
    firstResult := apiResponse.GeoCodeResults[0]
    if firstResult.Geometry.Location.Lat == 0 || firstResult.Geometry.Location.Lng == 0 {
        fmt.Println("Latitude or Longitude is missing or zero.")
        return Coordinates{}, fmt.Errorf("latitude or Longitude is missing or zero")
    }
	resultingCoordinates := Coordinates{
		Latitude: firstResult.Geometry.Location.Lat,
		Longitude: firstResult.Geometry.Location.Lng,
	}
	return resultingCoordinates, nil
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
	if _, err := time.LoadLocation(*request.Timezone); err != nil {
		return Response{}, fmt.Errorf("invalid timezone: %s", request.Timezone)
	}
	// databaseClient, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	// if err != nil {
	// 	fmt.Println("cannot initalize client", err)
	// }
	originCoordinates, err := getCoordinates(request.Origin)
	if err != nil {
        fmt.Println("Error obtaining origin location coordinates:", err)
        return Response{}, fmt.Errorf("error obtaining location coordinates: %s", err)
    }
	destinationCoordinates, err := getCoordinates(request.Destination)
	if err != nil {
        fmt.Println("Error obtaining destination location coordinates:", err)
        return Response{}, fmt.Errorf("error obtaining location coordinates: %s", err)
    }
	fmt.Printf("Origin Coordinates. latitude = %f, longitude = %f\n", originCoordinates.Latitude, originCoordinates.Longitude)
	fmt.Printf("Destination Coordinates. latitude = %f, longitude = %f\n", destinationCoordinates.Latitude, destinationCoordinates.Longitude)
	return Response{}, nil
}

func main() {
	lambda.Start(HandleRequest)
}