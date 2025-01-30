package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

type contactData struct {
	Entry      models.Contact
	Debts      []models.GetDebtsRow
	Activities []models.GetActivitiesRow
}

func (b *Controller) HandleContacts(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

		return
	}

	contacts, err := b.persister.GetContacts(r.Context(), email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(contacts); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleCreateContact(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

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

	newEmail := r.FormValue("email")
	if _, err := mail.ParseAddress(newEmail); err != nil {
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
		newEmail,
		pronouns,
		email,
	)
	if err != nil {
		log.Println(errCouldNotInsertIntoDB, err)

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(models.Contact{
		ID: id,

		FirstName: firstName,
		LastName:  lastName,
		Nickname:  nickname,
		Email:     newEmail,
		Pronouns:  pronouns,
	}); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleDeleteContact(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

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

	if err := b.persister.DeleteContact(r.Context(), int32(id), email); err != nil {
		log.Println(errCouldNotDeleteFromDB, err)

		http.Error(w, errCouldNotDeleteFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(id); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleViewContact(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

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

	contact, err := b.persister.GetContact(r.Context(), int32(id), email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	debts, err := b.persister.GetDebts(r.Context(), int32(id), email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	activities, err := b.persister.GetActivities(r.Context(), int32(id), email)
	if err != nil {
		log.Println(errCouldNotFetchFromDB, err)

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(contactData{
		Entry:      contact,
		Debts:      debts,
		Activities: activities,
	}); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}

func (b *Controller) HandleUpdateContact(w http.ResponseWriter, r *http.Request) {
	email, err := b.authorize(r)
	if err != nil {
		log.Println(err)

		http.Error(w, errCouldNotLogin.Error(), http.StatusUnauthorized)

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

	newEmail := r.FormValue("email")
	if _, err := mail.ParseAddress(newEmail); err != nil {
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
		newEmail,
		pronouns,
		email,
		birthday,
		address,
		notes,
	); err != nil {
		log.Println(errCouldNotUpdateInDB, err)

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	var birthdayDate sql.NullTime
	if birthday != nil {
		birthdayDate = sql.NullTime{
			Time:  *birthday,
			Valid: true,
		}
	}

	if err := json.NewEncoder(w).Encode(models.Contact{
		ID: int32(id),

		FirstName: firstName,
		LastName:  lastName,
		Nickname:  nickname,
		Email:     newEmail,
		Pronouns:  pronouns,
		Birthday:  birthdayDate,
		Address:   address,
		Notes:     notes,
	}); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}
