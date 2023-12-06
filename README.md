# go-page-matcher

## Introduction
**go-page-matcher** is an AWS Lambda tool designed for comparing the similarity of two web pages, particularly useful in site migration processes.

## Tools Used

- **devenv (devenv.sh)**: Enables fast, declarative, reproducible, and composable developer
environments using Nix.
- **Go Language**: The backbone of our application, providing efficiency and concurrency.
- **AWS Lambda**: Our serverless compute service where our function resides and gets executed.
- **Serverless Framework**: Facilitates deploying and managing applications on cloud platforms
without worrying about infrastructure.

## Prerequisites

Before diving into `go-screenshoter`, ensure you have the following installed:

- **Nix Language**: Essential for our `devenv` tool.
- **devenv tool**: Install it by following guidelines [here](devenv.sh).

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

With `devenv` and `Nix` already installed:

```bash
direnv shell
```

This will setup a consistent and reproducible developer environment for you.


## Deploy to AWS Lambda

You'll need to download the chromium binary compatible with x86_64.
E.g. [alixaxel/chrome-aws-lambda](https://raw.githubusercontent.com/alixaxel/chrome-aws-lambda/master/bin/chromium.br)

Extract it in the `layers/chromium` folder and make sure it has executable permissions.

```bash
wget -P layer https://raw.githubusercontent.com/alixaxel/chrome-aws-lambda/master/bin/chromium.br
brotli --decompress --rm --output=layer/chromium layers/chromium/chromium.br
chmod 777 layers/chromium/chromium
```

Follow Serverless framework guidelines to deploy the function to AWS Lambda. Ensure
your AWS credentials are properly set up.

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

## License
Licensed under BSD-3-Clause. See `LICENSE` file for details.

## Contact
Open an issue on GitHub for questions or feedback.
