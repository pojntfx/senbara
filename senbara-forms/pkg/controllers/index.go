package controllers

import (
	"log"
	"net/http"
)

func (b *Controller) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/" {
		_, userData, status, err := b.authorize(nil, r)
		if err != nil {
			log.Println(err)

			http.Error(w, err.Error(), status)

			return
		}

		if err := b.tpl.ExecuteTemplate(w, "index.html", indexData{
			pageData: pageData{
				userData: userData,

				Page:       userData.Locale.Get("Home"),
				PrivacyURL: b.privacyURL,
				ImprintURL: b.imprintURL,
			},
		}); err != nil {
			log.Println(errCouldNotRenderTemplate, err)

			http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

			return
		}

		return
	}

	redirected, userData, status, err := b.authorize(w, r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	if err := b.tpl.ExecuteTemplate(w, "404.html", pageData{
		userData: userData,

		Page:       userData.Locale.Get("Page not found"),
		PrivacyURL: b.privacyURL,
		ImprintURL: b.imprintURL,
	}); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}
