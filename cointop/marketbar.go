package cointop

import (
	"fmt"
	"math"
	"strings"
	"time"

	types "github.com/miguelmota/cointop/pkg/api/types"
	"github.com/miguelmota/cointop/pkg/color"
	"github.com/miguelmota/cointop/pkg/humanize"
	"github.com/miguelmota/cointop/pkg/pad"
	"github.com/miguelmota/cointop/pkg/ui"
	log "github.com/sirupsen/logrus"
)

// MarketbarView is structure for marketbar view
type MarketbarView = ui.View

// NewMarketbarView returns a new marketbar view
func NewMarketbarView() *MarketbarView {
	var view *MarketbarView = ui.NewView("marketbar")
	return view
}

// UpdateMarketbar updates the market bar view
func (ct *Cointop) UpdateMarketbar() error {
	log.Debug("UpdateMarketbar()")
	maxX := ct.Width()
	logo := "❯❯❯cointop"
	if ct.colorschemeName == "cointop" {
		logo = fmt.Sprintf("%s%s%s%s", color.Green("❯"), color.Cyan("❯"), color.Green("❯"), color.Cyan("cointop"))
	}
	var content string

	if ct.IsPortfolioVisible() {
		ct.State.marketBarHeight = 1
		total := ct.GetPortfolioTotal()
		totalstr := humanize.Monetaryf(total, 2)
		if !(ct.State.currencyConversion == "BTC" || ct.State.currencyConversion == "ETH" || total < 1) {
			total = math.Round(total*1e2) / 1e2
			totalstr = humanize.Monetaryf(total, 2)
		}

		timeframe := ct.State.selectedChartRange
		chartname := ct.SelectedCoinName()
		var charttitle string
		if chartname == "" {
			chartname = "Portfolio"
			charttitle = ct.colorscheme.MarketBarLabelActive(chartname)
		} else {
			charttitle = fmt.Sprintf("Portfolio - %s", ct.colorscheme.MarketBarLabelActive(chartname))
		}

		var percentChange24H float64
		for _, p := range ct.GetPortfolioSlice() {
			n := ((p.Balance / total) * p.PercentChange24H)
			if math.IsNaN(n) {
				continue
			}
			percentChange24H += n
		}

		color24h := ct.colorscheme.MarketbarSprintf()
		arrow := ""
		if percentChange24H > 0 {
			color24h = ct.colorscheme.MarketbarChangeUpSprintf()
			arrow = "▲"
		}
		if percentChange24H < 0 {
			color24h = ct.colorscheme.MarketbarChangeDownSprintf()
			arrow = "▼"
		}

		chartInfo := ""
		if !ct.State.hideChart {
			chartInfo = fmt.Sprintf(
				"[ Chart: %s %s ] ",
				charttitle,
				timeframe,
			)
		}

		totalstr = fmt.Sprintf("%s%s", ct.CurrencySymbol(), totalstr)
		if ct.State.hidePortfolioBalances {
			totalstr = HiddenBalanceChars
		}

		content = fmt.Sprintf(
			"%sTotal Portfolio Value: %s • 24H: %s",
			chartInfo,
			ct.colorscheme.MarketBarLabelActive(totalstr),
			color24h(fmt.Sprintf("%.2f%%%s", percentChange24H, arrow)),
		)
	} else {
		ct.State.marketBarHeight = 1
		if ct.Width() < 125 {
			ct.State.marketBarHeight = 2
		}

		var market types.GlobalMarketData
		var err error
		cachekey := ct.CacheKey("market")
		cached, found := ct.cache.Get(cachekey)

		if found {
			// cache hit
			var ok bool
			market, ok = cached.(types.GlobalMarketData)
			if ok {
				log.Debug("UpdateMarketbar() soft cache hit")
			}
		}

		if market.TotalMarketCapUSD == 0 {
			market, err = ct.api.GetGlobalMarketData(ct.State.currencyConversion)
			if err != nil {
				if ct.filecache != nil {
					ct.filecache.Get(cachekey, &market)
				}
			}

			ct.cache.Set(cachekey, market, 10*time.Second)
			if ct.filecache != nil {
				go func() {
					ct.filecache.Set(cachekey, market, 24*time.Hour)
				}()
			}
		}

		timeframe := ct.State.selectedChartRange
		chartname := ct.SelectedCoinName()
		if chartname == "" {
			chartname = "Global"
		}

		chartInfo := ""
		if !ct.State.hideChart {
			chartInfo = fmt.Sprintf(
				"[ Chart: %s %s ] ",
				ct.colorscheme.MarketBarLabelActive(chartname),
				timeframe,
			)
		}

		separator1 := "•"
		separator2 := "•"
		offset := strings.Repeat(" ", 12)
		if ct.Width() < 105 {
			separator1 = "\n" + offset
		} else if ct.Width() < 125 {
			separator2 = "\n" + offset
		}

		content = fmt.Sprintf(
			"%sGlobal ▶ Market Cap: %s %s 24H Volume: %s %s BTC Dominance: %.2f%%",
			chartInfo,
			fmt.Sprintf("%s%s", ct.CurrencySymbol(), humanize.Monetaryf(market.TotalMarketCapUSD, 0)),
			separator1,
			fmt.Sprintf("%s%s", ct.CurrencySymbol(), humanize.Monetaryf(market.Total24HVolumeUSD, 0)),
			separator2,
			market.BitcoinPercentageOfMarketCap,
		)
	}

	content = fmt.Sprintf("%s %s", logo, content)
	content = pad.Right(content, maxX, " ")
	content = ct.colorscheme.Marketbar(content)

	ct.UpdateUI(func() error {
		return ct.Views.Marketbar.Update(content)
	})

	return nil
}
