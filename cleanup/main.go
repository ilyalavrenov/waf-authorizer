package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ilyalavrenov/waf-authorizer/meta"
)

func Handler(ctx context.Context) (string, error) {
	sess, err := session.NewSession()

	ddb := dynamodb.New(sess)
	table := os.Getenv("DYNAMODB_TABLE")
	wafipset := os.Getenv("WAF_IPSET")

	timeout, err := strconv.Atoi(os.Getenv("AUTH_TIMEOUT_HOURS"))
	if err != nil {
		fmt.Println("Invalid env var AUTH_TIMEOUT_HOURS, must be an integer:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	ipsetid := meta.FindIPSetID(wafipset, meta.GetIPSets().IPSets)
	if ipsetid == "" {
		fmt.Println("unable to determine IP Set ID")
		os.Exit(1)
	}

	result, err := ddb.Scan(&dynamodb.ScanInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":a": {
				BOOL: aws.Bool(true),
			},
		},
		FilterExpression: aws.String("Active = :a"),
		TableName:        aws.String(table),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	for _, i := range result.Items {
		item := meta.Record{}

		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			fmt.Println("Got error unmarshalling:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		expired := item.DateRedeemed.Add(time.Duration(timeout) * time.Hour).Before(time.Now())

		if expired {
			fmt.Printf("expired: %s %s\n", item.AccessCode, item.IPAddress)

			_, err := meta.ChangeIPSet(*meta.GetWAFToken().ChangeToken, ipsetid, item.IPAddress, "DELETE")
			if err != nil {
				fmt.Println("Got error editing WAF IP set:")
				fmt.Println(err.Error())
			}

			item.Active = false
			item.DateDisabled = time.Now()

			av, err := dynamodbattribute.MarshalMap(item)
			if err != nil {
				fmt.Println("Got error marshalling:")
				fmt.Println(err.Error())
			}

			_, err = ddb.PutItem(&dynamodb.PutItemInput{
				Item:      av,
				TableName: aws.String(table),
			})
			if err != nil {
				fmt.Println("Got error putting item:")
				fmt.Println(err.Error())
			}
		}
	}

	return "Done", nil
}

func main() {
	lambda.Start(Handler)
}
