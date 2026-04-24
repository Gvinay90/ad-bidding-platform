#!/bin/bash
set -e
 
ENDPOINT=http://localhost:4566
REGION=us-east-1
 
awslocal sns create-topic --name campaign-events
awslocal sqs create-queue --queue-name bidder-cache
awslocal sqs create-queue --queue-name analytics-in
 
TOPIC_ARN=$(awslocal sns list-topics --query 'Topics[0].TopicArn' --output text)
BIDDER_ARN=$(awslocal sqs get-queue-attributes --queue-url $ENDPOINT/000000000000/bidder-cache --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)
ANALYTICS_ARN=$(awslocal sqs get-queue-attributes --queue-url $ENDPOINT/000000000000/analytics-in --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)
 
awslocal sns subscribe --topic-arn $TOPIC_ARN --protocol sqs --notification-endpoint $BIDDER_ARN
awslocal sns subscribe --topic-arn $TOPIC_ARN --protocol sqs --notification-endpoint $ANALYTICS_ARN
