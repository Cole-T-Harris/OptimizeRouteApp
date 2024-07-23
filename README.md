# OptimizeRouteApp

### Building and testing in Local:

Build binary and zip
```
./build.sh
```

Testing in Local For the OptimizeRouteFunction:
```
sam local invoke OptimizeRouteFunction --event event.json
```

Example event.json
```
{
  "user_id": 1,
  "route": 1,
  "origin": {
    "latitude": START_LATITUDE,
    "longitude": START_LONGITUDE
  },
  "destination": {
    "latitude": STOP_LATITUDE,
    "longitude": STOP_LONGITUDE
  }
}
```

Example template.yaml for sam cli:
```
AWSTemplateFormatVersion: '2010-09-09'
Transform: 'AWS::Serverless-2016-10-31'

Resources:
  OptimizeRouteFunction:
    Type: 'AWS::Serverless::Function'
    Properties:
      Handler: optimizeRoute
      Runtime: go1.x
      CodeUri: ./dist/optimizeRoute.zip
      Timeout: 10  
      MemorySize: 128
      Description: 'A Lambda function to calculate the commute distance and time'  
      Environment:
        Variables:
          GOOGLE_API_KEY: "YOUR_API_KEY"
          SUPABASE_URL: "YOUR_DATABASE_URL"
          SUPABASE_KEY: "YOUR_DATABASE_SECRET"

  CommutesQueueFunction:
    Type: 'AWS::Serverless::Function'
    Properties:
      Handler: commutesQueue
      Runtime: go1.x
      CodeUri: ./dist/commutesQueue.zip
      Timeout: 10  
      MemorySize: 128
      Description: 'A Lambda function to find the routes to send API requests'  
      Environment:
        Variables:
          SUPABASE_URL: "YOUR_DATABASE_URL"
          SUPABASE_KEY: "YOUR_DATABASE_SECRET"
```