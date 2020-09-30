package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ilyalavrenov/waf-authorizer/meta"
)

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	path := strings.Split(request.Path, "/")
	token := path[len(path)-1]

	ipaddr := request.RequestContext.Identity.SourceIP

	ddbresult, err := ddb.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]*dynamodb.AttributeValue{
			"AccessCode": {
				S: aws.String(token),
			},
		},
	})
	if err != nil {
		fmt.Println("Unable to get item:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	item := meta.Record{}

	err = dynamodbattribute.UnmarshalMap(ddbresult.Item, &item)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, nil
	}

	if item.AccessCode == "" {
		return events.APIGatewayProxyResponse{
			Body:       "access code invalid",
			StatusCode: 401,
		}, nil
	}

	expired := item.DateRedeemed.Add(time.Duration(timeout) * time.Hour).Before(time.Now())

	if item.Redeemed && item.IPAddress != ipaddr || item.Redeemed && expired {
		return events.APIGatewayProxyResponse{
			Body:       "access code has already been used or is expired",
			StatusCode: 401,
		}, nil
	}

	ipsetid := meta.FindIPSetID(wafipset, meta.GetIPSets().IPSets)

	if ipsetid == "" {
		return events.APIGatewayProxyResponse{
			Body:       "unable to determine IP set ID",
			StatusCode: 401,
		}, nil
	}

	item.Active = true
	item.Redeemed = true
	item.DateRedeemed = time.Now()
	item.IPAddress = ipaddr

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		fmt.Println("Got error marshalling:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(table),
	}

	if _, err := ddb.PutItem(input); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, nil
	}

	if _, err := meta.ChangeIPSet(*meta.GetWAFToken().ChangeToken, ipsetid, ipaddr, "INSERT"); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, nil
	}

	b := struct {
		Timeout   int
		IPAddress string
	}{
		Timeout:   timeout,
		IPAddress: ipaddr,
	}

	const html = `
<!DOCTYPE html>
<html>
    <head>
        <title>Authorization Successful</title>
    </head>
    <body>
        <p>Authorization successful for IP address: {{.IPAddress}}</p>
        <p>You have been authorized for {{.Timeout}} hours.</p>
    </body>
</html>`

	t, err := template.New("webpage").Parse(html)
	if err != nil {
		fmt.Println("Got error rendering html template:")
		fmt.Println(err.Error())
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, b); err != nil {
		fmt.Println(err)
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, nil
	}
	body := tpl.String()

	// pause briefly as WAF rules take a moment to go into effect
	time.Sleep(5000)

	return events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: 200,
		Headers:    map[string]string{"content-type": "text/html"},
	}, nil
}

func main() {
	lambda.Start(Handler)
}
