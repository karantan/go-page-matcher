// Package naviga provides utilities for browser interactions, specifically
// handling web navigations, login processes, screenshots, cookie management,
// and plugin installation for WordPress platform.
package naviga

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"page_matcher/storage"
	"page_matcher/utils"
	"path"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const (
	navigateTimeout    = 5 * time.Second
	navigationTimeout  = 5 * time.Second
	requestIdleTimeout = 10 * time.Second
	htmlTimeout        = 5 * time.Second
)

// Naviga struct holds all necessary elements for premorming browsaer actions
type Naviga struct {
	screenshotPath string
	launcher       *launcher.Launcher
	browser        *rod.Browser
	page           *rod.Page
	site           PageInfo
}

// PageInfo ...
type PageInfo struct {
	URL      string
	domain   string
	serverIP string
}

func launchInLambda(serverIP, domain string) *launcher.Launcher {
	if os.Getenv("APP_ENV") == "dev" {
		l := launcher.New().Headless(false).Devtools(true)
		// set custom DNS resolver
		// see https://datacadamia.com/web/browser/chrome#dns_resolver
		if serverIP != "" {
			log.Infow("Setting custom DNS resolver", domain, serverIP)
			l.Set("host-rules", fmt.Sprintf("MAP %s %s", domain, serverIP))
		} else {
			log.Info("Using default DNS resolver")
		}
		return l
	}

	l := launcher.New().
		// Lambda extracts the layer contents into the /opt directory in your
		// function’s execution environment
		// Ref: https://docs.aws.amazon.com/lambda/latest/dg/chapter-layers.html
		Bin("/opt/chromium").

		// recommended flags to run in serverless environments
		// see https://github.com/alixaxel/chrome-aws-lambda/blob/master/source/index.ts
		Set("allow-running-insecure-content").
		Set("autoplay-policy", "user-gesture-required").
		Set("disable-component-update").
		Set("disable-domain-reliability").
		Set("disable-features", "AudioServiceOutOfProcess", "IsolateOrigins", "site-per-process").
		Set("disable-print-preview").
		Set("disable-setuid-sandbox").
		Set("disable-site-isolation-trials").
		Set("disable-speech-api").
		Set("disable-web-security").
		Set("disk-cache-size", "33554432").
		Set("enable-features", "SharedArrayBuffer").
		Set("hide-scrollbars").
		Set("ignore-gpu-blocklist").
		Set("in-process-gpu").
		Set("mute-audio").
		Set("no-default-browser-check").
		Set("no-pings").
		Set("no-sandbox").
		Set("no-zygote").
		Set("single-process").
		Set("use-gl", "swiftshader").
		Set("window-size", "1920", "1080")

	if serverIP != "" {
		log.Infow("Setting custom DNS resolver", domain, serverIP)
		l.Set("host-rules", fmt.Sprintf("MAP %s %s", domain, serverIP))
	} else {
		log.Info("Using default DNS resolver")
	}

	return l
}

// NewBrowser creates and returns a new Naviga instance for given website credentials.
// It initializes a browser, connects to a page, and sets up paths for screenshots.
func NewBrowser(serverIP, website string) *Naviga {
	log.Infow("NewBrowser", "server", serverIP, "website", website)
	domain, err := utils.GetDomainFromURL(website)
	if err != nil {
		log.Error(err)
		log.Panicf("Can't extract domain from %s", website)
	}
	screenshotPath := path.Join(os.Getenv("TEMPORARY_STORAGE"), domain+".jpg")

	// instantiate the chromium launcher
	launcher := launchInLambda(serverIP, domain)
	u := launcher.MustLaunch()

	// create a browser instance
	browser := rod.New().ControlURL(u).MustConnect()
	// ignore cert errors because we set custom DNS resolver
	browser.IgnoreCertErrors(true)

	// open a page
	page := browser.MustPage()

	return &Naviga{
		screenshotPath: screenshotPath,
		launcher:       launcher,
		browser:        browser,
		page:           page,
		site: PageInfo{
			URL:      website,
			domain:   domain,
			serverIP: serverIP,
		},
	}
}

// Screenshot takes a screenshot of the current page in Naviga and uploads it to storage.
// Returns the pre-signed URL of the screenshot from the storage.
func (n *Naviga) Screenshot() (string, error) {
	log.Info("Trying to capture the screenshot")
	n.page.MustScreenshot(n.screenshotPath)
	objectKey := n.site.domain + ".jpg"
	b := storage.NewR2Client()
	storage.Upload(b.S3Client, objectKey, n.screenshotPath)
	url, err := storage.GetPresignURL(b.PresignClient, objectKey)
	log.Infow("Screenshot captured", "url", url)
	return url, err
}

// Navigate navigates the browser to `URL` and waits appropriate time for navigation,
// wait navigate and wait idle requests
func (n *Naviga) Navigate(URL string) error {
	err := rod.Try(func() {
		// close the page and open it again to kill all connections otherwise
		// waitNetworkResponse will wait forever
		n.page.MustClose()
		n.page = n.browser.MustPage()

		log.Infof("Navigating to %s", URL)

		e := proto.NetworkResponseReceived{}
		waitNetworkResponse := n.page.WaitEvent(&e)

		log.Debug("Waiting for navigation")
		n.page.Timeout(navigateTimeout).MustNavigate(URL)

		log.Debug("Waiting for network response event")
		waitNetworkResponse()

		log.Debug("Waiting for a page lifecycle events")
		waitNavigation := n.page.Timeout(navigationTimeout).MustWaitNavigation()
		waitNavigation()

		log.Debug("Waiting all requests to be idle")
		waitRequestIdle := n.page.Timeout(requestIdleTimeout).MustWaitRequestIdle()
		waitRequestIdle()

		log.Infof("Waiting done. Page %s loaded with status: %v", URL, e.Response.Status)
	})
	if err != nil {
		log.Errorw("Error in browser navigation", URL, err.Error())
	}
	return err
}

// GetHTML attempts to get HTML source code of the site
func (n *Naviga) GetHTML(URL string) (string, error) {
	statusCode, err := HTTPGetStatusCode(n.site.serverIP, n.site.domain)
	if err != nil {
		log.Errorw("Pre-flight check", "status code", statusCode, "error", err.Error())
		return "The system couldn't connect to the page.", err
	}
	if statusCode != 200 {
		log.Errorw("Pre-flight check", "status code", statusCode)
		return fmt.Sprintf("The system couldn't connect to the page (HTTP Status code: %v).", statusCode), err
	}

	err = n.Navigate(URL)
	if err != nil {
		return fmt.Sprintf("Error while trying to navigate to %s", URL), err
	}

	return n.page.MustHTML(), nil
}

// HTTPGetStatusCode function creates a request to the `url` and returns status code.
func HTTPGetStatusCode(ip, domain string) (int, error) {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: htmlTimeout,
		}).Dial,
		TLSHandshakeTimeout: htmlTimeout,
	}
	var client = &http.Client{
		Timeout:   htmlTimeout,
		Transport: netTransport,
	}
	if ip == "" {
		ip = domain
	}
	url := fmt.Sprintf("http://%s", ip)
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return 0, err
	}
	// This is the way you set "Host" in Go. `req.Header.Add("Host", domain)`
	// doesn't work because it get's ignored. See http/request.go:write
	req.Host = domain

	resp, err := client.Do(req)
	if err != nil {
		log.Errorw(err.Error(), "site", domain, "server", ip)
	}
	log.Infow("HTTPGetStatusCode", "status code", resp.StatusCode)
	return resp.StatusCode, err

}
