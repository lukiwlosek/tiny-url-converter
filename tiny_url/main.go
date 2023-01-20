package main

import (
	"context"
	"encoding/json"
	"fmt"

	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"crypto/md5"
	"encoding/hex"
)

type MyEvent struct {
	Url string `json:"url"`
}

type Item struct {
	Id	string	`json:"id"`
	Url	string	`json:"url"`
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	
	//setup 
	tableName := "TinyUrlLukasz"
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create DynamoDB client
	svc := dynamodb.New(sess)


	switch request.HTTPMethod {
	case "GET": 
		idQuery := request.QueryStringParameters["id"]
		result, err := svc.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
					"id": {
							S: aws.String(idQuery),
					},
			},
		})
		if err != nil {
				log.Fatalf("Got error calling GetItem: %s", err)
		}
		if result.Item == nil {
			response := events.APIGatewayProxyResponse{
				StatusCode: 404,
			}
			return response, nil
		}
		item := Item{}

		err = dynamodbattribute.UnmarshalMap(result.Item, &item)
		if err != nil {
				panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
		}

		if err != nil {
			log.Fatalf("Got error marshall: %s", err)
		}
		responseHeader := map[string]string{
			"Content-Type":"application/json",
			"Location": item.Url,
		}
		response := events.APIGatewayProxyResponse{
			StatusCode: 301,
			Headers: responseHeader,
		}
		return response, nil

	case "POST":
		var event MyEvent
		json.Unmarshal([]byte(request.Body), &event)

		// hash the url and shorten it
		id := GetMD5Hash(event.Url)[0:7]
		fullPath := "https://" + request.Headers["Host"] + request.Path + "?=" + id
		item := Item{
				Id:   id,
				Url:  event.Url,
		}
	
		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
				log.Fatalf("Got error marshalling new movie item: %s", err)
		}
		input := &dynamodb.PutItemInput{
				Item:      av,
				TableName: aws.String(tableName),
		}
	
		_, err = svc.PutItem(input)
		if err != nil {
				log.Fatalf("Got error calling PutItem: %s", err)
		}
	
	
		fmt.Println("Successfully added '" + item.Id + "' (" + item.Url + ") to table " + tableName)
		
		// convert item for resposne
		itemMarshalled, err := json.Marshal(MyEvent{Url: fullPath})
		if err != nil {
			log.Fatalf("Got error marshall: %s", err)
		}
		responseHeader := map[string]string{"Content-Type":"application/json"}
		response := events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body: string(itemMarshalled),
			Headers: responseHeader,
		}
		return response, nil
	}
	response := events.APIGatewayProxyResponse{
		StatusCode: 501,
	}
	return response, nil 
}

func main() {
	lambda.Start(HandleRequest)
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}