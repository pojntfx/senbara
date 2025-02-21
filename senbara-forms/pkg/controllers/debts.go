package controllers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/pojntfx/senbara/senbara-common/pkg/models"
)

type debtData struct {
	pageData
	Entry models.GetDebtAndContactRow
}

func (c *Controller) HandleAddDebt(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for add debt page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling add debt page")

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not prepare add debt page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not prepare add debt page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Getting contact for debt addition from DB", "id", id)

	contact, err := c.persister.GetContact(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Warn("Could not get contact from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := c.tpl.ExecuteTemplate(w, "debts_add.html", contactData{
		pageData: pageData{
			userData: userData,

			Page:       userData.Locale.Get("Add a debt"),
			PrivacyURL: c.privacyURL,
			ImprintURL: c.imprintURL,
		},
		Entry: contact,
	}); err != nil {
		log.Warn("Could not render template for adding a debt", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}

func (c *Controller) HandleCreateDebt(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for create debt", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling create debt")

	if err := r.ParseForm(); err != nil {
		log.Warn("Could not create debt", "err", errors.Join(errCouldNotParseForm, err))

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	ryouOwe := r.FormValue("you_owe")
	if strings.TrimSpace(ryouOwe) == "" {
		log.Warn("Could not create debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	youOwe, err := strconv.Atoi(ryouOwe)
	if err != nil {
		log.Warn("Could not create debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rcontactID := r.FormValue("contact_id")
	if strings.TrimSpace(rcontactID) == "" {
		log.Warn("Could not create debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	contactID, err := strconv.Atoi(rcontactID)
	if err != nil {
		log.Warn("Could not create debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	ramount := r.FormValue("amount")
	if strings.TrimSpace(rcontactID) == "" {
		log.Warn("Could not create debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	amount, err := strconv.ParseFloat(ramount, 64)
	if err != nil {
		log.Warn("Could not create debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	if youOwe == 1 {
		amount = -math.Abs(amount)
	} else {
		amount = math.Abs(amount)
	}

	currency := r.FormValue("currency")
	if strings.TrimSpace(currency) == "" {
		log.Warn("Could not create debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	description := r.FormValue("description")

	log.Debug("Creating debt in DB",
		"contactID", contactID,
		"amount", amount,
		"currency", currency,
		"description", description,
	)

	if _, err := c.persister.CreateDebt(
		r.Context(),

		amount,
		currency,
		description,

		int32(contactID),
		userData.Email,
	); err != nil {
		log.Warn("Could not create debt in DB", "err", errors.Join(errCouldNotInsertIntoDB, err))

		http.Error(w, errCouldNotInsertIntoDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/contacts/view?id=%v", contactID), http.StatusFound)
}

func (c *Controller) HandleSettleDebt(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for settle debt", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling settle debt")

	if err := r.ParseForm(); err != nil {
		log.Warn("Could not settle debt", "err", errors.Join(errCouldNotParseForm, err))

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not settle debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not settle debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rcontactID := r.FormValue("contact_id")
	if strings.TrimSpace(rcontactID) == "" {
		log.Warn("Could not settle debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	contactID, err := strconv.Atoi(rcontactID)
	if err != nil {
		log.Warn("Could not settle debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Settling debt in DB",
		"id", id,
		"contactID", contactID,
	)

	if _, err := c.persister.SettleDebt(
		r.Context(),

		int32(id),

		userData.Email,
	); err != nil {
		log.Warn("Could not settle debt in DB", "err", errors.Join(errCouldNotUpdateInDB, err))

		http.Error(w, errCouldNotUpdateInDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/contacts/view?id=%v", contactID), http.StatusFound)
}

func (c *Controller) HandleUpdateDebt(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for update debt", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling update debt")

	if err := r.ParseForm(); err != nil {
		log.Warn("Could not update debt", "err", errors.Join(errCouldNotParseForm, err))

		http.Error(w, errCouldNotParseForm.Error(), http.StatusInternalServerError)

		return
	}

	rid := r.FormValue("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not update debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not update debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	rcontactID := r.FormValue("contact_id")
	if strings.TrimSpace(rcontactID) == "" {
		log.Warn("Could not update debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	contactID, err := strconv.Atoi(rcontactID)
	if err != nil {
		log.Warn("Could not update debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	ryouOwe := r.FormValue("you_owe")
	if strings.TrimSpace(ryouOwe) == "" {
		log.Warn("Could not update debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	youOwe, err := strconv.Atoi(ryouOwe)
	if err != nil {
		log.Warn("Could not update debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	ramount := r.FormValue("amount")
	if strings.TrimSpace(rcontactID) == "" {
		log.Warn("Could not update debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	amount, err := strconv.ParseFloat(ramount, 64)
	if err != nil {
		log.Warn("Could not update debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	if youOwe == 1 {
		amount = -math.Abs(amount)
	} else {
		amount = math.Abs(amount)
	}

	currency := r.FormValue("currency")
	if strings.TrimSpace(currency) == "" {
		log.Warn("Could not update debt", "err", errInvalidForm)

		http.Error(w, errInvalidForm.Error(), http.StatusUnprocessableEntity)

		return
	}

	description := r.FormValue("description")

	log.Debug("Updating debt in DB",
		"id", id,
		"contactID", contactID,
		"amount", amount,
		"currency", currency,
		"description", description,
	)

	if _, err := c.persister.UpdateDebt(
		r.Context(),

		int32(id),

		userData.Email,

		amount,
		currency,
		description,
	); err != nil {
		log.Warn("Could not update debt in DB", "err", errors.Join(errCouldNotUpdateInDB, err))

		http.Error(w, errCouldNotUpdateInDB.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/contacts/view?id=%v", contactID), http.StatusFound)
}

func (c *Controller) HandleEditDebt(w http.ResponseWriter, r *http.Request) {
	redirected, userData, status, err := c.authorize(w, r, true)
	if err != nil {
		c.log.Warn("Could not authorize user for edit debt page", "err", err)

		http.Error(w, err.Error(), status)

		return
	} else if redirected {
		return
	}

	log := c.log.With("namespace", userData.Email)

	log.Debug("Handling edit debt page")

	rid := r.URL.Query().Get("id")
	if strings.TrimSpace(rid) == "" {
		log.Warn("Could not prepare edit debt page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	id, err := strconv.Atoi(rid)
	if err != nil {
		log.Warn("Could not prepare edit debt page", "err", errInvalidQueryParam)

		http.Error(w, errInvalidQueryParam.Error(), http.StatusUnprocessableEntity)

		return
	}

	log.Debug("Getting debt and contact for editing from DB", "id", id)

	debtAndContact, err := c.persister.GetDebtAndContact(r.Context(), int32(id), userData.Email)
	if err != nil {
		log.Warn("Could not get debt and contact from DB", "err", errors.Join(errCouldNotFetchFromDB, err))

		http.Error(w, errCouldNotFetchFromDB.Error(), http.StatusInternalServerError)

		return
	}

	if err := c.tpl.ExecuteTemplate(w, "debts_edit.html", debtData{
		pageData: pageData{
			userData: userData,

			Page:       userData.Locale.Get("Edit debt"),
			PrivacyURL: c.privacyURL,
			ImprintURL: c.imprintURL,
		},
		Entry: debtAndContact,
	}); err != nil {
		log.Warn("Could not render template for editing a debt", "err", errors.Join(errCouldNotRenderTemplate, err))

		http.Error(w, errCouldNotRenderTemplate.Error(), http.StatusInternalServerError)

		return
	}
}
