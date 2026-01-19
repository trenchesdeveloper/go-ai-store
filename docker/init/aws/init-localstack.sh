#!/bin/bash

# Create Bucket
awslocal s3 mb s3://ecommerce-uploads

echo "LocalStack S3 bucket created successfully."

# Create Queue
awslocal sqs create-queue --queue-name "ecommerce-events"
echo "LocalStack SQS queue created successfully."
