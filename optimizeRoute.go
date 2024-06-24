package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "time"
    "net/http"
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

func main() {
    departureTime := time.Now().Add(1 * time.Minute)

    request := RouteRequest{
        Origin: OriginDestination{
            Location: Location{
                LatLng: LatLng{
                    Latitude:  40.37619,
                    Longitude: -105.52351,
                },
            },
        },
        Destination: OriginDestination{
            Location: Location{
                LatLng: LatLng{
                    Latitude:  40.36646,
                    Longitude: -105.58111,
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
    jsonData, err := json.MarshalIndent(request, "", "  ")
    if err != nil {
        fmt.Println("Error marshaling JSON:", err)
        return
    }

    apiKey := "YOUR_API_KEY_HERE"
    url := "https://routes.googleapis.com/directions/v2:computeRoutes"
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Println("Error creating request:", err)
        return
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
        return
    }
    defer resp.Body.Close()

    // Read and print the response
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        fmt.Println("Error decoding response:", err)
        return
    }

    fmt.Printf("Response: %+v\n", result)
}