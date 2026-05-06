#!/bin/bash
set -euo pipefail

ENDPOINT="${AWS_ENDPOINT_URL:-http://localhost:4566}"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-us-east-1}"

awslocal sns create-topic --name campaign-events >/dev/null 2>&1 || true
awslocal sqs create-queue --queue-name bidder-cache >/dev/null 2>&1 || true
awslocal sqs create-queue --queue-name analytics-in >/dev/null 2>&1 || true

TOPIC_ARN=$(awslocal sns list-topics --query "Topics[?contains(TopicArn, 'campaign-events')].TopicArn | [0]" --output text)
if [[ -z "$TOPIC_ARN" || "$TOPIC_ARN" == "None" ]]; then
  echo "bootstrap: could not resolve campaign-events topic ARN" >&2
  exit 1
fi

BIDDER_ARN=$(awslocal sqs get-queue-attributes --queue-url "$ENDPOINT/000000000000/bidder-cache" --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)
ANALYTICS_ARN=$(awslocal sqs get-queue-attributes --queue-url "$ENDPOINT/000000000000/analytics-in" --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)

awslocal sns subscribe --topic-arn "$TOPIC_ARN" --protocol sqs --notification-endpoint "$BIDDER_ARN" >/dev/null 2>&1 || true
awslocal sns subscribe --topic-arn "$TOPIC_ARN" --protocol sqs --notification-endpoint "$ANALYTICS_ARN" >/dev/null 2>&1 || true
