# go-page-matcher

## Introduction
**go-page-matcher** is an AWS Lambda tool designed for comparing the similarity of two web pages, particularly useful in site migration processes.

## Features
- Compares web pages for similarity
- Deployed and managed through AWS Lambda
- Triggered via SNS messages
- Outputs results to an SQS queue with detailed similarity analysis

## Installation and deployment
To install, clone the repository:
```bash
git clone https://github.com/karantan/go-page-matcher.git
cd go-page-matcher
go mod tidy
```

To deploy on AWS run:
```bash
make deploy
```


## Usage
The tool is triggered by sending an SNS message to the "page-matcher" topic. The message should have the following JSON structure:
```json
{
    "old_page": {"server_ip": "", "url": "https://foo.com.si"},
    "new_page": {"server_ip": "85.90.246.138", "url": "http://bar.com"}
}
```
Upon comparison, the lambda function sends a success message to the "success-matche" SQS queue with the similarity index and, if applicable, screenshots of both pages.

Example of a message in the "success-matche" SQS queue:
```json
{
    "version": "1.0",
    "timestamp": "2023-12-01T18:59:06.962Z",
    "requestContext": {
        "requestId": "48796cb7-cb58-4439-ae03-629ba63ea46c",
        "functionArn": "arn:aws:lambda:us-east-1:...:function:page-matcher-v1-page_matcher:$LATEST",
        "condition": "Success",
        "approximateInvokeCount": 1
    },
    "requestPayload": {
        "Records": [
            {
                "EventSource": "aws:sns",
                "EventVersion": "1.0",
                "EventSubscriptionArn": "arn:aws:sns:us-east-1:...:page-matcher:8ce455a6-5b86-458e-bd56-dc0b42de2bc6",
                "Sns": {
                    "Type": "Notification",
                    "MessageId": "ad6bfc03-0933-57d9-8b7a-a1e7261b3a7b",
                    "TopicArn": "arn:aws:sns:us-east-1:...:page-matcher",
                    "Subject": null,
                    "Message": '{\n    "old_page": {\n\t    "server_ip": "",\n\t    "url": "https://karantan.si"\n    },\n    "new_page": {\n\t    "server_ip": "85.90.246.138",\n\t    "url": "http://testtist.fun"\n    }\n}\n',
                    "Timestamp": "2023-12-01T18:58:48.168Z",
                    "SignatureVersion": "1",
                    "Signature": "...",
                    "SigningCertUrl": "https://sns.us-east-1.amazonaws.com/SimpleNotificationService-....pem",
                    "UnsubscribeUrl": "https://sns.us-east-1.amazonaws.com/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:us-east-1:...:page-matcher:8ce455a6-5b86-458e-bd56-dc0b42de2bc6",
                    "MessageAttributes": {}
                }
            }
        ]
    },
    "responseContext": {"statusCode": 200, "executedVersion": "$LATEST"},
    "responsePayload": {
        "Similarity": 0.4469139917095424,
        "message": "",
        "old_screenshot_url": "https://...",
        "new_screenshot_url": "https://..."
    }
}
```

## Requirements
- AWS Lambda setup
- SNS and SQS configuration
- GoLang environment for development


## License
Licensed under BSD-3-Clause. See `LICENSE` file for details.

## Contact
Open an issue on GitHub for questions or feedback.
