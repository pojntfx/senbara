package controllers

import (
	"net/http"
	"strings"

	"github.com/leonelquinteros/gotext"
	"github.com/pojntfx/senbara/senbara-forms/pkg/locales"
	"golang.org/x/text/language"
)

func (c *Controller) localize(r *http.Request) (*gotext.Locale, error) {
	var locale *gotext.Locale
	tags, _, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
	if err != nil {
		return nil, err
	} else if len(tags) == 0 {
		locale = gotext.NewLocaleFS("en_US", locales.FS)
	} else {
		locale = gotext.NewLocaleFS(
			strings.ReplaceAll(tags[0].String(), "-", "_"),
			locales.FS,
		)
	}

	locale.AddDomain("default")

	return locale, nil
}
