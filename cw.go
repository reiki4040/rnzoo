package main

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type Billing struct {
	Label string
	Price float64
}

func GetBillingEstimatedCharges() (*Billing, error) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return nil, err
	}

	svc := cloudwatch.NewFromConfig(cfg)

	startTime := time.Now().Add(time.Hour * -8)
	endTime := time.Now()

	// billing register every 4h
	// get -8 hour to now
	// priod(grouping) 24h / point
	params := &cloudwatch.GetMetricStatisticsInput{
		EndTime:    aws.Time(endTime),
		MetricName: aws.String("EstimatedCharges"),
		Namespace:  aws.String("AWS/Billing"),
		Period:     aws.Int32(86400),
		StartTime:  aws.Time(startTime),
		Dimensions: []types.Dimension{
			types.Dimension{
				Name:  aws.String("Currency"),
				Value: aws.String("USD"),
			},
		},
		Statistics: []types.Statistic{
			types.StatisticMaximum,
		},
	}
	resp, err := svc.GetMetricStatistics(ctx, params)
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
