// Command screenshot is a chromedp example demonstrating how to take a
// screenshot of a specific element and of the entire browser viewport.
package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

var (
	headless  = flag.Bool("headless", true, "run browser in headless mode")
	outputDir = flag.String("out", ".", "directory where to save screenshots")
	selector  = flag.String("selector", "", "take screenshot of element matching selector")
	timeout   = flag.Duration("timeout", 10*time.Second, "limit program execution")
)

func main() {
	flag.Parse()

	url := flag.Arg(0)
	if url == "" {
		log.Fatal("Missing URL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	ctx, cancel = chromedp.NewExecAllocator(
		ctx,
		append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", *headless),
		)...,
	)
	defer cancel()

	// create context
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var buf []byte
	if *selector != "" {
		// capture screenshot of an element
		if err := chromedp.Run(ctx, elementScreenshot(url, *selector, &buf)); err != nil {
			log.Fatal(err)
		}
		if err := ioutil.WriteFile(filepath.Join(*outputDir, "elementScreenshot.png"), buf, 0644); err != nil {
			log.Fatal(err)
		}
	} else {
		// capture entire browser viewport, returning png with quality=90
		if err := chromedp.Run(ctx, fullScreenshot(url, 90, &buf)); err != nil {
			log.Fatal(err)
		}
		if err := ioutil.WriteFile(filepath.Join(*outputDir, "fullScreenshot.png"), buf, 0644); err != nil {
			log.Fatal(err)
		}
	}
}

// elementScreenshot takes a screenshot of a specific element.
func elementScreenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.WaitVisible(sel),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible),
	}
}

// fullScreenshot takes a screenshot of the entire browser viewport.
//
// Liberally copied from puppeteer's source.
//
// Note: this will override the viewport emulation settings.
func fullScreenshot(urlstr string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
	}
}
