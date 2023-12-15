// main page matcher lambda function
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"page_matcher/logger"
	"page_matcher/naviga"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var log = logger.New("main", false)

// OldPage ...
type OldPage struct {
	ServerIP string `json:"server_ip"`
	URL      string `json:"url"`
}

// NewPage ...
type NewPage struct {
	ServerIP string `json:"server_ip"`
	URL      string `json:"url"`
}

// Message is a struct for holding SNS message
type Message struct {
	OldPage OldPage `json:"old_page"`
	NewPage NewPage `json:"new_page"`
}

// ResponsePayload struct holds information we send back to SQS.
// Empty message means everything went OK and we have nothing to report, otherwise
// we populate Message with the reason where and why things didn't go as planed.
type ResponsePayload struct {
	Similarity    float64
	Message       string `json:"message"`
	OldScreenshot string `json:"old_screenshot_url"`
	NewScreenshot string `json:"new_screenshot_url"`
}

// handler lambda function
func handler(_ context.Context, snsEvent events.SNSEvent) (ResponsePayload, error) {
	// all notification messages will contain a single published message.
	// Ref: https://aws.amazon.com/sns/faqs/ (Reliability section)
	snsRecord := snsEvent.Records[0].SNS
	log.Infof("Message = %s", snsRecord.Message)
	r := ResponsePayload{}

	message := Message{}
	err := json.Unmarshal([]byte(snsRecord.Message), &message)
	if err != nil {
		return r, fmt.Errorf("unexpected error parsing SNS message: %v", err)
	}

	oldPage := message.OldPage
	o := naviga.NewBrowser(oldPage.ServerIP, oldPage.URL)
	oldHTML, err := o.GetHTML(oldPage.URL)
	if err != nil {
		log.Errorw("Get HTML", "page", "old", "err", err)
		r.Message = oldHTML
		screenshotURL, err := o.Screenshot()
		if err == nil {
			r.OldScreenshot = screenshotURL
		}
		return r, err
	}

	newPage := message.NewPage
	n := naviga.NewBrowser(newPage.ServerIP, newPage.URL)
	newHTML, err := n.GetHTML(newPage.URL)
	if err != nil {
		log.Errorw("Get HTML", "page", "new", "err", err)
		r.Message = newHTML
		screenshotURL, err := n.Screenshot()
		if err == nil {
			r.NewScreenshot = screenshotURL
		}
		return r, err
	}

	// Sorensen-Dice
	sd := metrics.NewSorensenDice()
	sdSimilarity := strutil.Similarity(oldHTML, newHTML, sd)
	log.Infow("Sorensen-Dice", "similarity", sdSimilarity)
	// r.Similarity = sdSimilarity

	// New Jaccard
	nj := metrics.NewJaccard()
	njSimilarity := strutil.Similarity(oldHTML, newHTML, nj)
	log.Infow("Jaccard", "similarity", njSimilarity)
	r.Similarity = njSimilarity

	if njSimilarity < 0.95 {
		oldScreenshotURL, oerr := o.Screenshot()
		newScreenshotURL, nerr := n.Screenshot()
		if oerr == nil && nerr == nil {
			r.OldScreenshot = oldScreenshotURL
			r.NewScreenshot = newScreenshotURL
		} else {
			log.Error(oerr)
			log.Error(nerr)
		}
	}

	return r, nil
}

func main() {
	lambda.Start(handler)
}
