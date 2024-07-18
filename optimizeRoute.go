package main

import (
    "bytes"
	"context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
	"github.com/aws/aws-lambda-go/lambda"
    // "github.com/supabase-community/postgrest-go"
)

type LatLng struct {
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
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
    Origin                  OriginDestination `json:"origin"`
    Destination             OriginDestination `json:"destination"`
    TravelMode              string            `json:"travelMode"`
    RoutingPreference       string            `json:"routingPreference"`
    DepartureTime           time.Time         `json:"departureTime"`
    ComputeAlternativeRoutes bool             `json:"computeAlternativeRoutes"`
    RouteModifiers          RouteModifiers    `json:"routeModifiers"`
    LanguageCode            string            `json:"languageCode"`
    Units                   string            `json:"units"`
}

type Request struct {
	Origin      LatLng `json:"origin"`
	Destination LatLng `json:"destination"`
}

type Response struct {
    Message string                 `json:"message"`
    Data    Data                   `json:"data"`
}

type Data struct {
	Routes []Route `json:"routes"`
}      

type Route struct {
	DistanceMeters int     `json:"distanceMeters"`
	Duration       string  `json:"duration"`
	Polyline       Polyline `json:"polyline"`
}

type Polyline struct {
	EncodedPolyline string `json:"encodedPolyline"`
}

// var supabaseURL = os.Getenv("SUPABASE_URL")
// var supabaseKey = os.Getenv("SUPABASE_KEY")

func HandleRequest(ctx context.Context, request Request) (Response, error) {
    if request.Origin.Latitude == 0 || request.Origin.Longitude == 0 ||
        request.Destination.Latitude == 0 || request.Destination.Longitude == 0 {
        return Response{}, fmt.Errorf("invalid request. Missing origin or destination coordinates")
    }

    departureTime := time.Now().Add(1 * time.Minute)
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
        TravelMode:              "DRIVE",
        RoutingPreference:       "TRAFFIC_AWARE",
        DepartureTime:           departureTime,
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
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil{
		fmt.Println("Error decoding response:", err)
		return Response{}, fmt.Errorf("error decoding response: %v", err)
	}

	// Assuming there's at least one route in the response
	if len(responseData.Routes) > 0 {
		// Extract the values into new variables
		distanceMeters := responseData.Routes[0].DistanceMeters
		duration := responseData.Routes[0].Duration
		encodedPolyline := responseData.Routes[0].Polyline.EncodedPolyline

		// Print the extracted values
		fmt.Printf("Distance Meters: %d\n", distanceMeters)
		fmt.Printf("Duration: %s\n", duration)
		fmt.Printf("Encoded Polyline: %s\n", encodedPolyline)
	} else {
		fmt.Println("No routes found in the response")
        return Response{Message: "No routes found in the response", Data: responseData}, nil
	}

    //  // Insert data into Supabase
    //  databaseClient := postgrest.NewClient(supabaseURL, supabaseKey, nil)
    //  data := LatLng{
    //      Latitude:  apiResponse.Latitude,
    //      Longitude: apiResponse.Longitude,
    //  }
 
    //  response, err := databaseClient.From("your_table_name").Insert(data, false, "", "").Execute()
    //  if err != nil {
    //      return fmt.Errorf("error inserting data into Supabase: %v", err)
    //  }
 
    //  if response.StatusCode != http.StatusCreated {
    //      return fmt.Errorf("unexpected Supabase response status: %s", response.Status)
    //  }

    return Response{Message: "Request successful", Data: responseData}, nil
}

func main() {
    lambda.Start(HandleRequest)
}