package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// SnsActions encapsulates SNS actions
type SnsActions struct {
	snsClient *sns.Client
}

// Returns a new SnsActions object
func NewSnsActions(cfg aws.Config, region string) *SnsActions {
	snsClient := sns.NewFromConfig(cfg, func(opt *sns.Options) {
		opt.Region = region
	})

	return &SnsActions{snsClient}
}

// Publishes the provided message to the SNS topic defined by topicArn
func (actor *SnsActions) Publish(topicArn string, message string) (*sns.PublishOutput, error) {
	input := sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(topicArn),
	}
	output, err := actor.snsClient.Publish(context.TODO(), &input)
	if err != nil {
		log.Printf("Error calling 'sns:Publish' - %v", err)
	}

	return output, err
}
