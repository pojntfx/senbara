package main

import "errors"

var (
	errCouldNotLogin            = errors.New("could not login")
	errCouldNotWriteSettingsKey = errors.New("could not write settings key")
	errMissingPrivacyURL        = errors.New("missing privacy policy URL")
	errMissingContactID         = errors.New("missing contact ID")
	errInvalidContactID         = errors.New("invalid contact ID")
	errMissingActivityID        = errors.New("missing activity ID")
	errInvalidActivityID        = errors.New("invalid activity ID")
	errDebtDoesNotExist         = errors.New("debt does not exist")
	errMissingJournalEntryID    = errors.New("missing journal entry ID")
	errInvalidJournaEntrylID    = errors.New("invalid journal entry ID")
)
