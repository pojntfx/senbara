package controllers

import (
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/pojntfx/senbara/senbara-forms/pkg/models"
)

type contactsData struct {
	pageData
	Entries []models.Contact
}

type contactData struct {
	pageData
	Entry      models.Contact
	Debts      []models.GetDebtsRow
	Activities []models.GetActivitiesRow
}

func (b *Controller) HandleContacts(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	contacts, err := b.persister.GetContacts(r.Context(), userData.Email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := b.tpl.ExecuteTemplate(w, "contacts.html", contactsData{
		pageData: pageData{
			userData: userData,

			Page:       userData.Locale.Get("Contacts"),
			PrivacyURL: b.privacyURL,
			ImprintURL: b.imprintURL,
		},
		Entries: contacts,
	}); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleAddContact(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	if err := b.tpl.ExecuteTemplate(w, "contacts_add.html", pageData{
		userData: userData,

		Page:       userData.Locale.Get("Add a contact"),
		PrivacyURL: b.privacyURL,
		ImprintURL: b.imprintURL,
	}); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleCreateContact(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println(errCouldNotParseForm, err)

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	firstName := r.FormValue("first_name")
	if strings.TrimSpace(firstName) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	lastName := r.FormValue("last_name")
	if strings.TrimSpace(lastName) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	email := r.FormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		log.Println(err)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	nickname := r.FormValue("nickname")

	pronouns := r.FormValue("pronouns")
	if strings.TrimSpace(pronouns) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := b.persister.CreateContact(
		r.Context(),
		firstName,
		lastName,
		nickname,
		email,
		pronouns,
		userData.Email,
	)
	if err != nil {
		log.Println(errCouldNotInsertIntoDB, err)

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/contacts/view?id=%v", id), http.StatusFound)
}

func (b *Controller) HandleDeleteContact(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println(errCouldNotParseForm, err)

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	if err := b.persister.DeleteContact(r.Context(), int32(id), userData.Email); err != nil {
		log.Println(errCouldNotDeleteFromDB, err)

		http.Error(w, errCouldNotDeleteFromDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/contacts", http.StatusFound)
}

func (b *Controller) HandleViewContact(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Println(errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Println(errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	contact, err := b.persister.GetContact(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	debts, err := b.persister.GetDebts(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	activities, err := b.persister.GetActivities(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := b.tpl.ExecuteTemplate(w, "contacts_view.html", contactData{
		pageData: pageData{
			userData: userData,

			Page:       contact.FirstName + " " + contact.LastName,
			PrivacyURL: b.privacyURL,
			ImprintURL: b.imprintURL,

			BackURL: "/contacts",
		},
		Entry:      contact,
		Debts:      debts,
		Activities: activities,
	}); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleUpdateContact(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println(errCouldNotParseForm, err)

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	firstName := r.FormValue("first_name")
	if strings.TrimSpace(firstName) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	lastName := r.FormValue("last_name")
	if strings.TrimSpace(lastName) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	email := r.FormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		log.Println(err)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	nickname := r.FormValue("nickname")

	pronouns := r.FormValue("pronouns")
	if strings.TrimSpace(pronouns) == "" {
		log.Println(errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rbirthday := r.FormValue("birthday")

	var birthday *time.Time
	if strings.TrimSpace(rbirthday) != "" {
		b, err := time.Parse("2006-01-02", rbirthday)
		if err != nil {
			log.Println(errInvalidForm)

			http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

			return
		}

		birthday = &b
	}

	address := r.FormValue("address")

	notes := r.FormValue("notes")

	if err := b.persister.UpdateContact(
		r.Context(),
		int32(id),
		firstName,
		lastName,
		nickname,
		email,
		pronouns,
		userData.Email,
		birthday,
		address,
		notes,
	); err != nil {
		log.Println(errCouldNotUpdateInDB, err)

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/contacts/view?id="+rid, http.StatusFound)
}

func (b *Controller) HandleEditContact(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := b.authorize(w, r)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Println(errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Println(errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	contact, err := b.persister.GetContact(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := b.tpl.ExecuteTemplate(w, "contacts_edit.html", contactData{
		pageData: pageData{
			userData: userData,

			Page:       userData.Locale.Get("Edit contact"),
			PrivacyURL: b.privacyURL,
			ImprintURL: b.imprintURL,
		},
		Entry: contact,
	}); err != nil {
		log.Println(errCouldNotRenderTemplate, err)

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}
