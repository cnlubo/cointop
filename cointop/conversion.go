package cointop

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	color "github.com/miguelmota/cointop/pkg/color"
	"github.com/miguelmota/cointop/pkg/pad"
	log "github.com/sirupsen/logrus"
)

// FiatCurrencyNames is a mpa of currency symbols to names.
// Keep these in alphabetical order.
var FiatCurrencyNames = map[string]string{
	"AUD": "Australian Dollar",
	"BGN": "Bulgarian lev",
	"BRL": "Brazilian Real",
	"CAD": "Canadian Dollar",
	"CFH": "Swiss Franc",
	"CLP": "Chilean Peso",
	"CNY": "Chinese Yuan",
	"CZK": "Czech Koruna",
	"DKK": "Danish Krone",
	"EUR": "Euro",
	"GBP": "British Pound",
	"HKD": "Hong Kong Dollar",
	"HRK": "Croatian kuna",
	"HUF": "Hungarian Forint",
	"IDR": "Indonesian Rupiah",
	"ILS": "Israeli New Shekel",
	"INR": "Indian Rupee",
	"ISK": "Icelandic króna",
	"JPY": "Japanese Yen",
	"KRW": "South Korean Won",
	"MXN": "Mexican Peso",
	"MYR": "Malaysian Ringgit",
	"NOK": "Norwegian Krone",
	"NZD": "New Zealand Dollar",
	"PHP": "Philippine Peso",
	"PKR": "Pakistani Rupe",
	"PLN": "Polish złoty",
	"RON": "Romanian leu",
	"RUB": "Russian Ruble",
	"SEK": "Swedish Krona",
	"SGD": "Singapore Dollar",
	"THB": "Thai Baht",
	"TRY": "Turkish lira",
	"TWD": "New Taiwan Dollar",
	"UAH": "Ukrainian hryvnia",
	"USD": "US Dollar",
	"VND": "Vietnamese Dong",
	"ZAR": "South African Rand",
}

// CryptocurrencyNames is a map of cryptocurrency symbols to name
var CryptocurrencyNames = map[string]string{
	"BTC": "Bitcoin",
	"ETH": "Ethereum",
}

// CurrencySymbolMap is map of fiat currency symbols to names.
// Keep these in alphabetical order.
var CurrencySymbolMap = map[string]string{
	"AUD": "$",
	"BGN": "Лв.",
	"BRL": "R$",
	"BTC": "Ƀ",
	"CAD": "$",
	"CFH": "₣",
	"CLP": "$",
	"CNY": "¥",
	"CZK": "Kč",
	"DKK": "Kr",
	"ETH": "Ξ",
	"EUR": "€",
	"GBP": "£",
	"HKD": "$",
	"HRK": "kn",
	"HUF": "Ft",
	"IDR": "Rp.",
	"ILS": "₪",
	"INR": "₹",
	"ISK": "kr",
	"JPY": "¥",
	"KRW": "₩",
	"MXN": "$",
	"MYR": "RM",
	"NOK": "kr",
	"NZD": "$",
	"PHP": "₱",
	"PKR": "₨",
	"PLN": "zł",
	"RON": "lei",
	"RUB": "Ꝑ",
	"SEK": "kr",
	"SGD": "S$",
	"THB": "฿",
	"TRY": "₺",
	"TWD": "NT$",
	"UAH": "₴",
	"USD": "$",
	"VND": "₫",
	"ZAR": "R",
}

var alphanumericcharacters = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}

// IsSupportedCurrencyConversion returns true if it's a supported currency conversion
func (ct *Cointop) IsSupportedCurrencyConversion(convert string) bool {
	conversions := ct.SupportedCurrencyConversions()
	_, ok := conversions[convert]
	return ok
}

// SupportedCurrencyConversions returns a map of all supported currencies for conversion
func (ct *Cointop) SupportedCurrencyConversions() map[string]string {
	all := map[string]string{}
	for _, symbol := range ct.api.SupportedCurrencies() {
		if v, ok := FiatCurrencyNames[symbol]; ok {
			all[symbol] = v
		}
		if v, ok := CryptocurrencyNames[symbol]; ok {
			all[symbol] = v
		}
	}

	return all
}

// SupportedFiatCurrencyConversions returns map of supported fiat currencies for conversion
func (ct *Cointop) SupportedFiatCurrencyConversions() map[string]string {
	return FiatCurrencyNames
}

// SupportedCryptoCurrencyConversions returns map of supported cryptocurrencies for conversion
func (ct *Cointop) SupportedCryptoCurrencyConversions() map[string]string {
	return CryptocurrencyNames
}

// SortedSupportedCurrencyConversions returns sorted list of supported currencies for conversion
func (ct *Cointop) SortedSupportedCurrencyConversions() []string {
	currencies := ct.SupportedCurrencyConversions()
	keys := make([]string, 0, len(currencies))
	for k := range currencies {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// UpdateConvertMenu updates the convert menu
func (ct *Cointop) UpdateConvertMenu() error {
	log.Debug("UpdateConvertMenu()")
	header := ct.colorscheme.MenuHeader(fmt.Sprintf(" Currency Conversion %s\n\n", pad.Left("[q] close ", ct.Width()-24, " ")))
	helpline := " Press the corresponding key to select currency for conversion\n\n"
	cnt := 0
	h := ct.Views.Menu.Height()
	percol := h - 5
	cols := make([][]string, percol)
	for i := range cols {
		cols[i] = make([]string, 20)
	}

	keys := ct.SortedSupportedCurrencyConversions()
	currencies := ct.SupportedCurrencyConversions()
	for i, key := range keys {
		currency := currencies[key]
		symbol := CurrencySymbol(key)
		if cnt%percol == 0 {
			cnt = 0
		}
		shortcut := string(alphanumericcharacters[i])
		if key == ct.State.currencyConversion {
			shortcut = ct.colorscheme.MenuLabelActive(color.Bold("*"))
			key = ct.colorscheme.Menu(color.Bold(key))
			currency = ct.colorscheme.MenuLabelActive(color.Bold(currency))
		} else {
			key = ct.colorscheme.Menu(key)
			currency = ct.colorscheme.MenuLabel(currency)
		}

		item := fmt.Sprintf(" [ %1s ] %4s %-36s", shortcut, key, fmt.Sprintf("%s %s", currency, symbol))
		cols[cnt] = append(cols[cnt], item)
		cnt = cnt + 1
	}

	var body string
	for i := 0; i < percol; i++ {
		var row string
		for j := 0; j < len(cols[i]); j++ {
			item := cols[i][j]
			row = fmt.Sprintf("%s%s", row, item)
		}
		body = fmt.Sprintf("%s%s\n", body, row)
	}

	content := fmt.Sprintf("%s%s%s", header, helpline, body)
	ct.UpdateUI(func() error {
		ct.Views.Menu.SetFrame(true)
		return ct.Views.Menu.Update(content)
	})

	return nil
}

// SetCurrencyConverstion sets the currency conversion
func (ct *Cointop) SetCurrencyConverstion(convert string) error {
	convert = strings.ToUpper(convert)
	if convert == "" {
		return nil
	}

	if !ct.IsSupportedCurrencyConversion(convert) {
		return errors.New("unsupported currency conversion")
	}

	// NOTE: return if the currency selection wasn't changed
	if ct.State.currencyConversion == convert {
		return nil
	}

	ct.State.currencyConversion = convert
	return nil
}

// SetCurrencyConverstionFn sets the currency conversion function
func (ct *Cointop) SetCurrencyConverstionFn(convert string) func() error {
	log.Debug("SetCurrencyConverstionFn()")
	return func() error {
		ct.HideConvertMenu()

		if err := ct.SetCurrencyConverstion(convert); err != nil {
			return err
		}

		if err := ct.Save(); err != nil {
			return err
		}

		go ct.RefreshAll()
		return nil
	}
}

// CurrencySymbol returns the symbol for the currency conversion
func (ct *Cointop) CurrencySymbol() string {
	log.Debug("CurrencySymbol()")
	return CurrencySymbol(ct.State.currencyConversion)
}

// ShowConvertMenu shows the convert menu view
func (ct *Cointop) ShowConvertMenu() error {
	log.Debug("ShowConvertMenu()")
	ct.State.convertMenuVisible = true
	ct.UpdateConvertMenu()
	ct.SetActiveView(ct.Views.Menu.Name())
	return nil
}

// HideConvertMenu hides the convert menu view
func (ct *Cointop) HideConvertMenu() error {
	log.Debug("HideConvertMenu()")
	ct.State.convertMenuVisible = false
	ct.ui.SetViewOnBottom(ct.Views.Menu)
	ct.SetActiveView(ct.Views.Table.Name())
	ct.UpdateUI(func() error {
		ct.Views.Menu.SetFrame(false)
		return ct.Views.Menu.Update("")
	})
	return nil
}

// ToggleConvertMenu toggles the convert menu view
func (ct *Cointop) ToggleConvertMenu() error {
	log.Debug("ToggleConvertMenu()")
	ct.State.convertMenuVisible = !ct.State.convertMenuVisible
	if ct.State.convertMenuVisible {
		return ct.ShowConvertMenu()
	}
	return ct.HideConvertMenu()
}

// CurrencySymbol returns the symbol for the currency name
func CurrencySymbol(currency string) string {
	symbol, ok := CurrencySymbolMap[strings.ToUpper(currency)]
	if ok {
		return symbol
	}

	return "?"
}
