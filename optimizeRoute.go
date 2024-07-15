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
    Data    map[string]interface{} `json:"data"`
}

func HandleRequest(ctx context.Context, request Request) (Response, error) {
    if request.Origin.Latitude == 0 || request.Origin.Longitude == 0 ||
        request.Destination.Latitude == 0 || request.Destination.Longitude == 0 {
        return Response{Message: "Invalid request. Missing origin or destination coordinates."}, nil
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
        return Response{Message: "Error marshaling JSON", Data: nil}, err
    }

    apiKey := os.Getenv("GOOGLE_API_KEY")
    url := "https://routes.googleapis.com/directions/v2:computeRoutes"
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Println("Error creating request:", err)
        return Response{Message: "Error creating request", Data: nil}, err
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
        return Response{Message: "Error sending request", Data: nil}, err
    }
    defer resp.Body.Close()

    // Read and print the response
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        fmt.Println("Error decoding response:", err)
        return Response{Message: "Error decoding response", Data: nil}, err
    }

    return Response{Message: "Request successful", Data: result}, nil
}

func main() {
    lambda.Start(HandleRequest)
}