package controllers

import (
	"net/http"
	"strings"

	"github.com/leonelquinteros/gotext"
	"github.com/pojntfx/senbara/senbara-forms/config/locales"
	"golang.org/x/text/language"
)

func (c *Controller) localize(r *http.Request) (*gotext.Locale, error) {
	acceptLanguageHeader := r.Header.Get("Accept-Language")

	c.log.Debug("Setting up locale", "acceptLanguageHeader", acceptLanguageHeader)

	var locale *gotext.Locale
	tags, _, err := language.ParseAcceptLanguage(acceptLanguageHeader)
	if err != nil {
		return nil, err
	} else if len(tags) == 0 {
		c.log.Debug("Could not find locale, falling back to en_US")

		locale = gotext.NewLocaleFS("en_US", locales.FS)
	} else {
		localeCode := strings.ReplaceAll(tags[0].String(), "-", "_")

		c.log.Debug("Found matching locale", "localeCode", localeCode)

		locale = gotext.NewLocaleFS(
			localeCode,
			locales.FS,
		)
	}

	locale.AddDomain("default")

	return locale, nil
}
