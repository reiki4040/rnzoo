package main

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type Billing struct {
	Label string
	Price float64
}

func GetBillingEstimatedCharges() (*Billing, error) {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
	svc := cloudwatch.New(sess)

	startTime := time.Now().Add(time.Hour * -8)
	endTime := time.Now()

	// billing register every 4h
	// get -8 hour to now
	// priod(grouping) 24h / point
	params := &cloudwatch.GetMetricStatisticsInput{
		EndTime:    aws.Time(endTime),
		MetricName: aws.String("EstimatedCharges"),
		Namespace:  aws.String("AWS/Billing"),
		Period:     aws.Int64(86400),
		StartTime:  aws.Time(startTime),
		Dimensions: []*cloudwatch.Dimension{
			&cloudwatch.Dimension{
				Name:  aws.String("Currency"),
				Value: aws.String("USD"),
			},
		},
		Statistics: []*string{
			aws.String("Maximum"),
		},
	}
	resp, err := svc.GetMetricStatistics(params)
	if err != nil {
		return nil, err
	}

	if resp.Label == nil {
		return nil, errors.New("Label is empty. failed get billing.")
	}

	if len(resp.Datapoints) == 0 {
		return nil, errors.New("Datapoints is empty... failed get billing.")
	}

	if resp.Datapoints[0].Maximum == nil {
		return nil, errors.New("Datapoint Maximum is empty... failed get billing.")
	}

	b := &Billing{
		Label: *resp.Label,
		Price: *resp.Datapoints[0].Maximum,
	}

	return b, nil
}
