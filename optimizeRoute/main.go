package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/supabase-community/supabase-go"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type LatLng struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type Location struct {
	LatLng LatLng `json:"latLng"`
}

type RouteModifiers struct {
	AvoidTolls    bool `json:"avoidTolls"`
	AvoidHighways bool `json:"avoidHighways"`
	AvoidFerries  bool `json:"avoidFerries"`
}

type OriginDestination struct {
	Location Location `json:"location"`
}

type RouteRequest struct {
	Origin                   OriginDestination `json:"origin"`
	Destination              OriginDestination `json:"destination"`
	TravelMode               string            `json:"travelMode"`
	RoutingPreference        string            `json:"routingPreference"`
	DepartureTime            time.Time         `json:"departureTime"`
	ComputeAlternativeRoutes bool              `json:"computeAlternativeRoutes"`
	RouteModifiers           RouteModifiers    `json:"routeModifiers"`
	LanguageCode             string            `json:"languageCode"`
	Units                    string            `json:"units"`
}

type Request struct {
	UserID      *int    `json:"user_id"`
	Origin      LatLng  `json:"origin"`
	Destination LatLng  `json:"destination"`
	Route       *int    `json:"route"`
	ToWork      *bool   `json:"to_work"`
	Timezone    *string `json:"timezone"`
}

type Response struct {
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

type Data struct {
	Routes []Route `json:"routes"`
}

type Route struct {
	DistanceMeters int      `json:"distanceMeters"`
	Duration       string   `json:"duration"`
	Polyline       Polyline `json:"polyline"`
}

type Polyline struct {
	EncodedPolyline string `json:"encodedPolyline"`
}

type QueryRecord struct {
	UserID    int       `json:"user_id"`
	QueryTime time.Time `json:"query_time"`
	Duration  int       `json:"duration"`
	Distance  int       `json:"distance"`
	Route     int       `json:"route"`
	RouteHash string    `json:"route_hash"`
	ToWork    bool      `json:"to_work"`
	DayOfWeek string    `json:"day_of_week"`
}

var supabaseURL = os.Getenv("SUPABASE_URL")
var supabaseKey = os.Getenv("SUPABASE_KEY")

func HandleRequest(ctx context.Context, request Request) (Response, error) {
	if request.UserID == nil {
		return Response{}, fmt.Errorf("invalid request. Missing user ID")
	}
	if request.ToWork == nil {
		return Response{}, fmt.Errorf("invalid request. Missing to_work")
	}
	if request.Route == nil {
		return Response{}, fmt.Errorf("invalid request. Missing route ID")
	}
	if request.Timezone == nil {
		return Response{}, fmt.Errorf("invalid request. Missing timezone")
	}

	if request.Origin.Latitude == nil || request.Origin.Longitude == nil ||
		request.Destination.Latitude == nil || request.Destination.Longitude == nil {
		return Response{}, fmt.Errorf("invalid request. Missing origin or destination coordinates")
	}

	databaseClient, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		fmt.Println("cannot initalize client", err)
	}
	loc, err := time.LoadLocation(*request.Timezone)
	if err != nil {
		fmt.Println("Error loading location:", err)
		return Response{}, fmt.Errorf("error laoding location: %v", err)
	}
	departureTime := time.Now().UTC().In(loc).Add(1 * time.Minute)
	originLat := request.Origin.Latitude
	originLng := request.Origin.Longitude
	destLat := request.Destination.Latitude
	destLng := request.Destination.Longitude

	googleRequest := RouteRequest{
		Origin: OriginDestination{
			Location: Location{
				LatLng: LatLng{
					Latitude:  originLat,
					Longitude: originLng,
				},
			},
		},
		Destination: OriginDestination{
			Location: Location{
				LatLng: LatLng{
					Latitude:  destLat,
					Longitude: destLng,
				},
			},
		},
		TravelMode:               "DRIVE",
		RoutingPreference:        "TRAFFIC_AWARE",
		DepartureTime:            departureTime,
		ComputeAlternativeRoutes: false,
		RouteModifiers: RouteModifiers{
			AvoidTolls:    false,
			AvoidHighways: false,
			AvoidFerries:  false,
		},
		LanguageCode: "en-US",
		Units:        "IMPERIAL",
	}

	// Marshal the struct to JSON
	jsonData, err := json.MarshalIndent(googleRequest, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return Response{}, fmt.Errorf("error marshaling JSON: %v", err)
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	url := "https://routes.googleapis.com/directions/v2:computeRoutes"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return Response{}, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", apiKey)
	req.Header.Set("X-Goog-FieldMask", "routes.duration,routes.distanceMeters,routes.polyline.encodedPolyline")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return Response{}, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read and print the response
	var responseData Data
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		fmt.Println("Error decoding response:", err)
		return Response{}, fmt.Errorf("error decoding response: %v", err)
	}

	// Assuming there's at least one route in the response
	if len(responseData.Routes) > 0 {
		// Extract the values into new variables
		distanceMeters := responseData.Routes[0].DistanceMeters
		duration := responseData.Routes[0].Duration
		encodedPolyline := responseData.Routes[0].Polyline.EncodedPolyline

		numericPart := strings.TrimSuffix(duration, "s")
		durationInt, err := strconv.Atoi(numericPart)
		if err != nil {
			fmt.Println("Error:", err)
			return Response{}, fmt.Errorf("error converting duration: %v", err)
		}

		record := QueryRecord{
			UserID:    *request.UserID,
			QueryTime: departureTime,
			Duration:  durationInt,
			Distance:  distanceMeters,
			Route:     *request.Route,
			RouteHash: encodedPolyline,
			ToWork:    *request.ToWork,
			DayOfWeek: departureTime.Weekday().String(),
		}
		upsert := false //Update if there is data with the same key

		response, rowsEffected, err := databaseClient.From("commutes").Insert(record, upsert, "", "*", "exact").Execute()
		if err != nil {
			fmt.Printf("Failed to insert data: %v", err)
			return Response{}, fmt.Errorf("failed to insert data: %v", err)
		}

		if rowsEffected != 1 {
			fmt.Printf("Incorrect of rows effected, rows effected: %d", rowsEffected)
			return Response{}, fmt.Errorf("incorrect of rows effected, rows effected: %d", rowsEffected)
		}
		fmt.Println("Query Succesful:", response)
	} else {
		fmt.Println("No routes found in the response")
		return Response{Message: "No routes found in the response", Data: responseData}, nil
	}

	return Response{Message: "Request successful", Data: responseData}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
