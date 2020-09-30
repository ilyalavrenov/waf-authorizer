package meta

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/waf"
)

type Record struct {
	AccessCode   string    `json:"AccessCode"`
	CreatedBy    string    `json:"CreatedBy"`
	DateCreated  time.Time `json:"DateCreated"`
	DateRedeemed time.Time `json:"DateRedeemed"`
	DateDisabled time.Time `json:"DateDisabled"`
	IPAddress    string    `json:"IPAddress"`
	Redeemed     bool      `json:"Redeemed"`
	Active       bool      `json:"Active"`
}

func GetWAFToken() *waf.GetChangeTokenOutput {
	svc := waf.New(session.New())
	input := &waf.GetChangeTokenInput{}
	result, err := svc.GetChangeToken(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case waf.ErrCodeInternalErrorException:
				fmt.Println(waf.ErrCodeInternalErrorException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}
	return result
}

func GetIPSets() *waf.ListIPSetsOutput {
	svc := waf.New(session.New())
	input := &waf.ListIPSetsInput{
		Limit: aws.Int64(100),
	}
	result, err := svc.ListIPSets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case waf.ErrCodeInternalErrorException:
				fmt.Println(waf.ErrCodeInternalErrorException, aerr.Error())
			case waf.ErrCodeInvalidAccountException:
				fmt.Println(waf.ErrCodeInvalidAccountException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}
	return result
}

func FindIPSetID(name string, list []*waf.IPSetSummary) (id string) {
	for _, i := range list {
		if *i.Name == name {
			return *i.IPSetId
		}
	}
	return ""
}

func ChangeIPSet(changetoken, ipsetid, ipaddr, action string) (*waf.UpdateIPSetOutput, error) {
	svc := waf.New(session.New())
	ipcidr := strings.Join([]string{ipaddr, "/32"}, "")
	input := &waf.UpdateIPSetInput{
		ChangeToken: aws.String(changetoken),
		IPSetId:     aws.String(ipsetid),
		Updates: []*waf.IPSetUpdate{
			{
				Action: aws.String(action),
				IPSetDescriptor: &waf.IPSetDescriptor{
					Type:  aws.String("IPV4"),
					Value: aws.String(ipcidr),
				},
			},
		},
	}

	result, error := svc.UpdateIPSet(input)

	return result, error
}
