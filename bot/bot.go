package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"worstTwitter/user"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

var ReportChan = make(chan user.UserID, 100)

func RunWorker() {
	go func() {
		for {
			err := processQueue()
			if err != nil {
				fmt.Println(err)
			}
		}
	}()
}

func processQueue() error {
	flag := os.Getenv("FLAG")
	if flag == "" {
		flag = "FLAG{DUMMY}"
	}

	staff_id := <-ReportChan

	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var denyButtons = []*cdp.Node{}
	err := chromedp.Run(ctx,
		emulation.SetUserAgentOverride("WebScraper 1.0"),
		SetCookiesAction(&network.CookieParam{
			Name:     "SESSION_USERID",
			Value:    strconv.Itoa(staff_id),
			Domain:   "localhost",
			SameSite: network.CookieSameSiteLax,
			HTTPOnly: true,
		}),

		chromedp.Navigate(`http://localhost:8080/home`),
	)
	if err != nil {
		return err
	}

	tctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	chromedp.Run(tctx, chromedp.Nodes(".deny", &denyButtons, chromedp.ByQueryAll))

	for _, button := range denyButtons {
		err = chromedp.Run(ctx,
			chromedp.MouseClickNode(button, chromedp.ButtonType(input.Left)),
		)
		if err != nil {
			return err
		}
	}
	err = chromedp.Run(ctx,
		chromedp.Click(`#accept_all`, chromedp.NodeVisible),
		chromedp.SendKeys(`input[name="tweet_text"]`, flag, chromedp.NodeVisible),
		chromedp.Click(`input[value="tweet"]`, chromedp.NodeVisible),
		chromedp.Sleep(1*time.Second),
	)
	return err
}

func SetCookiesAction(cookie *network.CookieParam) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		return network.SetCookies([]*network.CookieParam{cookie}).Do(ctx)
	})
}
