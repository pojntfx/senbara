package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/jwijenbergh/puregotk/v4/webkit"
	"github.com/oapi-codegen/runtime/types"
	. "github.com/pojntfx/go-gettext/pkg/i18n"
	"github.com/pojntfx/senbara/senbara-common/pkg/authn"
	"github.com/pojntfx/senbara/senbara-gnome/assets/resources"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/zalando/go-keyring"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	level := new(slog.LevelVar)
	log := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))

	if err := initEmbeddedI18n(); err != nil {
		panic(err)
	}

	settings, err := initEmbeddedSettings()
	if err != nil {
		panic(err)
	}

	a := adw.NewApplication(resources.AppID, gio.GApplicationHandlesOpenValue)

	var (
		w       adw.ApplicationWindow
		nv      adw.NavigationView
		mto     adw.ToastOverlay
		authner *authn.Authner
		u       userData
	)

	// TODO: Re-use the main authorizer
	authorize := func(
		ctx context.Context,

		loginIfSignedOut bool,
	) (
		redirected bool,

		client *api.ClientWithResponses,
		status int,

		err error,
	) {
		authorizer := newAuthorizer(
			log,

			authner,
			settings,

			func() string {
				return nv.GetVisiblePage().GetTag()
			},
			func(tags []string, position int) {
				nv.ReplaceWithTags(tags, position)
			},
			func(ud userData) {
				u = ud
			},
			func() string {
				return settings.GetString(resources.SettingServerURLKey)
			},
		)

		return authorizer.authorize(ctx, loginIfSignedOut)
	}

	var rawError string
	onPanic := func(err error) {
		rawError = err.Error()
		i18nErr := L(rawError)

		log.Error(
			"An unexpected error occured, showing error message to user",
			"rawError", rawError,
			"i18nErr", i18nErr,
		)

		toast := adw.NewToast(i18nErr)
		toast.SetButtonLabel(L("Copy details"))
		toast.SetActionName("app.copyErrorToClipboard")

		mto.AddToast(toast)
	}

	onActivate := func(_ gio.Application) {
		aboutDialog := adw.NewAboutDialogFromAppdata(resources.ResourceMetainfoPath, resources.AppVersion)
		aboutDialog.SetDevelopers(resources.AppDevelopers)
		aboutDialog.SetArtists(resources.AppArtists)
		aboutDialog.SetCopyright(resources.AppCopyright)

		b := gtk.NewBuilderFromResource(resources.ResourceWindowUIPath)

		preferencesDialogBuilder := gtk.NewBuilderFromResource(resources.ResourcePreferencesDialogUIPath)
		contactsCreateDialogBuilder := gtk.NewBuilderFromResource(resources.ResourceContactsCreateDialogUIPath)
		debtsCreateDialogBuilder := gtk.NewBuilderFromResource(resources.ResourceDebtsCreateDialogUIPath)
		activitiesCreateDialogBuilder := gtk.NewBuilderFromResource(resources.ActivitiesDebtsCreateDialogUIPath)
		journalEntriesCreateDialogBuilder := gtk.NewBuilderFromResource(resources.JournalEntriesCreateDialogUIPath)

		b.GetObject("main_window").Cast(&w)

		b.GetObject("main_navigation").Cast(&nv)

		b.GetObject("main_toasts_overlay").Cast(&mto)

		var (
			preferencesDialog              adw.PreferencesDialog
			preferencesDialogVerboseSwitch gtk.Switch

			welcomeGetStartedButton  gtk.Button
			welcomeGetStartedSpinner adw.Spinner

			configServerURLInput           adw.EntryRow
			configServerURLContinueButton  gtk.Button
			configServerURLContinueSpinner adw.Spinner

			spec openapi3.T

			previewLoginButton  gtk.Button
			previewLoginSpinner adw.Spinner

			previewContactsCountLabel   gtk.Label
			previewContactsCountSpinner adw.Spinner

			previewJournalEntriesCountLabel   gtk.Label
			previewJournalEntriesCountSpinner adw.Spinner

			oidcDcrInitialAccessTokenPortalUrl string

			registerRegisterButton gtk.Button

			configInitialAccessTokenInput        adw.PasswordEntryRow
			configInitialAccessTokenLoginButton  gtk.Button
			configInitialAccessTokenLoginSpinner adw.Spinner

			exchangeLoginCancelButton  gtk.Button
			exchangeLogoutCancelButton gtk.Button

			homeSplitView      adw.NavigationSplitView
			homeNavigation     adw.NavigationView
			homeSidebarListbox gtk.ListBox
			homeContentPage    adw.NavigationPage

			homeUserMenuButton  gtk.MenuButton
			homeUserMenuAvatar  adw.Avatar
			homeUserMenuSpinner adw.Spinner

			homeHamburgerMenuButton  gtk.MenuButton
			homeHamburgerMenuIcon    gtk.Image
			homeHamburgerMenuSpinner adw.Spinner

			homeSidebarContactsCountLabel   gtk.Label
			homeSidebarContactsCountSpinner adw.Spinner

			homeSidebarJournalEntriesCountLabel   gtk.Label
			homeSidebarJournalEntriesCountSpinner adw.Spinner

			contactsStack       gtk.Stack
			contactsListBox     gtk.ListBox
			contactsSearchEntry gtk.SearchEntry

			contactsAddButton    gtk.Button
			contactsSearchButton gtk.ToggleButton

			contactsEmptyAddButton gtk.Button

			contactsCreateDialog adw.Dialog

			contactsCreateDialogAddButton  gtk.Button
			contactsCreateDialogAddSpinner adw.Spinner

			contactsCreateDialogFirstNameInput adw.EntryRow
			contactsCreateDialogLastNameInput  adw.EntryRow
			contactsCreateDialogNicknameInput  adw.EntryRow
			contactsCreateDialogEmailInput     adw.EntryRow
			contactsCreateDialogPronounsInput  adw.EntryRow

			contactsCreateDialogEmailWarningButton gtk.MenuButton

			contactsErrorStatusPage        adw.StatusPage
			contactsErrorRefreshButton     gtk.Button
			contactsErrorCopyDetailsButton gtk.Button

			contactsViewPageTitle              adw.WindowTitle
			contactsViewStack                  gtk.Stack
			contactsViewErrorStatusPage        adw.StatusPage
			contactsViewErrorRefreshButton     gtk.Button
			contactsViewErrorCopyDetailsButton gtk.Button

			contactsViewEditButton   gtk.Button
			contactsViewDeleteButton gtk.Button

			contactsViewOptionalFieldsPreferencesGroup adw.PreferencesGroup

			contactsViewBirthdayRow adw.ActionRow
			contactsViewAddressRow  adw.ActionRow
			contactsViewNotesRow    adw.ActionRow

			contactsViewDebtsListBox      gtk.ListBox
			contactsViewActivitiesListBox gtk.ListBox

			activitiesViewPageTitle              adw.WindowTitle
			activitiesViewStack                  gtk.Stack
			activitiesViewErrorStatusPage        adw.StatusPage
			activitiesViewErrorRefreshButton     gtk.Button
			activitiesViewErrorCopyDetailsButton gtk.Button

			activitiesViewEditButton   gtk.Button
			activitiesViewDeleteButton gtk.Button

			activitiesViewPageBodyWebView webkit.WebView

			activitiesEditPageTitle              adw.WindowTitle
			activitiesEditStack                  gtk.Stack
			activitiesEditErrorStatusPage        adw.StatusPage
			activitiesEditErrorRefreshButton     gtk.Button
			activitiesEditErrorCopyDetailsButton gtk.Button

			activitiesEditPageSaveButton  gtk.Button
			activitiesEditPageSaveSpinner adw.Spinner

			activitiesEditPageNameInput           adw.EntryRow
			activitiesEditPageDateInput           adw.EntryRow
			activitiesEditPageDescriptionExpander adw.ExpanderRow
			activitiesEditPageDescriptionInput    gtk.TextView

			activitiesEditPageDateWarningButton gtk.MenuButton

			activitiesEditPagePopoverLabel gtk.Label

			debtsEditPageTitle              adw.WindowTitle
			debtsEditStack                  gtk.Stack
			debtsEditErrorStatusPage        adw.StatusPage
			debtsEditErrorRefreshButton     gtk.Button
			debtsEditErrorCopyDetailsButton gtk.Button

			debtsEditPageSaveButton  gtk.Button
			debtsEditPageSaveSpinner adw.Spinner

			debtsEditPageYouOweRadio         gtk.CheckButton
			debtsEditPageAmountInput         adw.SpinRow
			debtsEditPageCurrencyInput       adw.EntryRow
			debtsEditPageDescriptionExpander adw.ExpanderRow
			debtsEditPageDescriptionInput    gtk.TextView

			debtsEditPageYouOweActionRow  adw.ActionRow
			debtsEditPageTheyOweActionRow adw.ActionRow

			contactsEditPageTitle              adw.WindowTitle
			contactsEditStack                  gtk.Stack
			contactsEditErrorStatusPage        adw.StatusPage
			contactsEditErrorRefreshButton     gtk.Button
			contactsEditErrorCopyDetailsButton gtk.Button

			contactsEditPageSaveButton  gtk.Button
			contactsEditPageSaveSpinner adw.Spinner

			contactsEditPageFirstNameInput adw.EntryRow
			contactsEditPageLastNameInput  adw.EntryRow
			contactsEditPageNicknameInput  adw.EntryRow
			contactsEditPageEmailInput     adw.EntryRow
			contactsEditPagePronounsInput  adw.EntryRow

			contactsEditPageBirthdayInput   adw.EntryRow
			contactsEditPageAddressExpander adw.ExpanderRow
			contactsEditPageAddressInput    gtk.TextView
			contactsEditPageNotesExpander   adw.ExpanderRow
			contactsEditPageNotesInput      gtk.TextView

			contactsEditPageEmailWarningButton    gtk.MenuButton
			contactsEditPageBirthdayWarningButton gtk.MenuButton

			contactsEditPagePopoverLabel gtk.Label

			debtsCreateDialog adw.Dialog

			debtsCreateDialogAddButton  gtk.Button
			debtsCreateDialogAddSpinner adw.Spinner

			debtsCreateDialogTitle adw.WindowTitle

			debtsCreateDialogYouOweRadio         gtk.CheckButton
			debtsCreateDialogAmountInput         adw.SpinRow
			debtsCreateDialogCurrencyInput       adw.EntryRow
			debtsCreateDialogDescriptionExpander adw.ExpanderRow
			debtsCreateDialogDescriptionInput    gtk.TextView

			debtsCreateDialogYouOweActionRow  adw.ActionRow
			debtsCreateDialogTheyOweActionRow adw.ActionRow

			activitiesCreateDialog adw.Dialog

			activitiesCreateDialogAddButton  gtk.Button
			activitiesCreateDialogAddSpinner adw.Spinner

			activitiesCreateDialogTitle adw.WindowTitle

			activitiesCreateDialogNameInput           adw.EntryRow
			activitiesCreateDialogDateInput           adw.EntryRow
			activitiesCreateDialogDescriptionExpander adw.ExpanderRow
			activitiesCreateDialogDescriptionInput    gtk.TextView

			activitiesCreateDialogDateWarningButton gtk.MenuButton

			activitiesCreateDialogPopoverLabel gtk.Label

			journalEntriesStack       gtk.Stack
			journalEntriesListBox     gtk.ListBox
			journalEntriesSearchEntry gtk.SearchEntry

			journalEntriesAddButton    gtk.Button
			journalEntriesSearchButton gtk.ToggleButton

			journalEntriesEmptyAddButton gtk.Button

			journalEntriesErrorStatusPage        adw.StatusPage
			journalEntriesErrorRefreshButton     gtk.Button
			journalEntriesErrorCopyDetailsButton gtk.Button

			journalEntriesCreateDialog adw.Dialog

			journalEntriesCreateDialogAddButton  gtk.Button
			journalEntriesCreateDialogAddSpinner adw.Spinner

			journalEntriesCreateDialogRatingToggleGroup adw.ToggleGroup
			journalEntriesCreateDialogTitleInput        adw.EntryRow
			journalEntriesCreateDialogBodyExpander      adw.ExpanderRow
			journalEntriesCreateDialogBodyInput         gtk.TextView

			journalEntriesViewPageTitle              adw.WindowTitle
			journalEntriesViewStack                  gtk.Stack
			journalEntriesViewErrorStatusPage        adw.StatusPage
			journalEntriesViewErrorRefreshButton     gtk.Button
			journalEntriesViewErrorCopyDetailsButton gtk.Button

			journalEntriesViewEditButton   gtk.Button
			journalEntriesViewDeleteButton gtk.Button

			journalEntriesViewPageBodyWebView webkit.WebView

			journalEntriesEditPageTitle              adw.WindowTitle
			journalEntriesEditStack                  gtk.Stack
			journalEntriesEditErrorStatusPage        adw.StatusPage
			journalEntriesEditErrorRefreshButton     gtk.Button
			journalEntriesEditErrorCopyDetailsButton gtk.Button

			journalEntriesEditPageSaveButton  gtk.Button
			journalEntriesEditPageSaveSpinner adw.Spinner

			journalEntriesEditPageRatingToggleGroup adw.ToggleGroup
			journalEntriesEditPageTitleInput        adw.EntryRow
			journalEntriesEditPageBodyExpander      adw.ExpanderRow
			journalEntriesEditPageBodyInput         gtk.TextView
		)

		preferencesDialogBuilder.GetObject("preferences_dialog").Cast(&preferencesDialog)
		preferencesDialogBuilder.GetObject("preferences_dialog_verbose_switch").Cast(&preferencesDialogVerboseSwitch)
		b.GetObject("welcome_get_started_button").Cast(&welcomeGetStartedButton)
		b.GetObject("welcome_get_started_spinner").Cast(&welcomeGetStartedSpinner)
		b.GetObject("config_server_url_input").Cast(&configServerURLInput)
		b.GetObject("config_server_url_continue_button").Cast(&configServerURLContinueButton)
		b.GetObject("config_server_url_continue_spinner").Cast(&configServerURLContinueSpinner)
		b.GetObject("preview_login_button").Cast(&previewLoginButton)
		b.GetObject("preview_login_spinner").Cast(&previewLoginSpinner)
		b.GetObject("preview_contacts_count_label").Cast(&previewContactsCountLabel)
		b.GetObject("preview_contacts_count_spinner").Cast(&previewContactsCountSpinner)
		b.GetObject("preview_journal_entries_count_label").Cast(&previewJournalEntriesCountLabel)
		b.GetObject("preview_journal_entries_count_spinner").Cast(&previewJournalEntriesCountSpinner)
		b.GetObject("register_register_button").Cast(&registerRegisterButton)
		b.GetObject("config_initial_access_token_input").Cast(&configInitialAccessTokenInput)
		b.GetObject("config_initial_access_token_login_button").Cast(&configInitialAccessTokenLoginButton)
		b.GetObject("config_initial_access_token_login_spinner").Cast(&configInitialAccessTokenLoginSpinner)
		b.GetObject("exchange_login_cancel_button").Cast(&exchangeLoginCancelButton)
		b.GetObject("exchange_logout_cancel_button").Cast(&exchangeLogoutCancelButton)
		b.GetObject("home_split_view").Cast(&homeSplitView)
		b.GetObject("home_navigation").Cast(&homeNavigation)
		b.GetObject("home_sidebar_listbox").Cast(&homeSidebarListbox)
		b.GetObject("home_content_page").Cast(&homeContentPage)
		b.GetObject("home_user_menu_button").Cast(&homeUserMenuButton)
		b.GetObject("home_user_menu_avatar").Cast(&homeUserMenuAvatar)
		b.GetObject("home_user_menu_spinner").Cast(&homeUserMenuSpinner)
		b.GetObject("home_hamburger_menu_button").Cast(&homeHamburgerMenuButton)
		b.GetObject("home_hamburger_menu_icon").Cast(&homeHamburgerMenuIcon)
		b.GetObject("home_hamburger_menu_spinner").Cast(&homeHamburgerMenuSpinner)
		b.GetObject("home_sidebar_contacts_count_label").Cast(&homeSidebarContactsCountLabel)
		b.GetObject("home_sidebar_contacts_count_spinner").Cast(&homeSidebarContactsCountSpinner)
		b.GetObject("home_sidebar_journal_entries_count_label").Cast(&homeSidebarJournalEntriesCountLabel)
		b.GetObject("home_sidebar_journal_entries_count_spinner").Cast(&homeSidebarJournalEntriesCountSpinner)
		b.GetObject("contacts_stack").Cast(&contactsStack)
		b.GetObject("contacts_list").Cast(&contactsListBox)
		b.GetObject("contacts_searchentry").Cast(&contactsSearchEntry)
		b.GetObject("contacts_add_button").Cast(&contactsAddButton)
		b.GetObject("contacts_search_button").Cast(&contactsSearchButton)
		b.GetObject("contacts_empty_add_button").Cast(&contactsEmptyAddButton)
		contactsCreateDialogBuilder.GetObject("contacts_create_dialog").Cast(&contactsCreateDialog)
		contactsCreateDialogBuilder.GetObject("contacts_create_dialog_add_button").Cast(&contactsCreateDialogAddButton)
		contactsCreateDialogBuilder.GetObject("contacts_create_dialog_add_spinner").Cast(&contactsCreateDialogAddSpinner)
		contactsCreateDialogBuilder.GetObject("contacts_create_dialog_first_name_input").Cast(&contactsCreateDialogFirstNameInput)
		contactsCreateDialogBuilder.GetObject("contacts_create_dialog_last_name_input").Cast(&contactsCreateDialogLastNameInput)
		contactsCreateDialogBuilder.GetObject("contacts_create_dialog_nickname_input").Cast(&contactsCreateDialogNicknameInput)
		contactsCreateDialogBuilder.GetObject("contacts_create_dialog_email_input").Cast(&contactsCreateDialogEmailInput)
		contactsCreateDialogBuilder.GetObject("contacts_create_dialog_pronouns_input").Cast(&contactsCreateDialogPronounsInput)
		contactsCreateDialogBuilder.GetObject("contacts_create_dialog_email_warning_button").Cast(&contactsCreateDialogEmailWarningButton)
		b.GetObject("contacts_error_status_page").Cast(&contactsErrorStatusPage)
		b.GetObject("contacts_error_refresh_button").Cast(&contactsErrorRefreshButton)
		b.GetObject("contacts_error_copy_details").Cast(&contactsErrorCopyDetailsButton)
		b.GetObject("contacts_view_page_title").Cast(&contactsViewPageTitle)
		b.GetObject("contacts_view_stack").Cast(&contactsViewStack)
		b.GetObject("contacts_view_error_status_page").Cast(&contactsViewErrorStatusPage)
		b.GetObject("contacts_view_error_refresh_button").Cast(&contactsViewErrorRefreshButton)
		b.GetObject("contacts_view_error_copy_details").Cast(&contactsViewErrorCopyDetailsButton)
		b.GetObject("contacts_view_edit_button").Cast(&contactsViewEditButton)
		contactsViewEditButton.SetActionTargetValue(glib.NewVariantInt64(0))
		b.GetObject("contacts_view_delete_button").Cast(&contactsViewDeleteButton)
		contactsViewDeleteButton.SetActionTargetValue(glib.NewVariantInt64(0))
		b.GetObject("contacts_view_optional_fields").Cast(&contactsViewOptionalFieldsPreferencesGroup)
		b.GetObject("contacts_view_birthday").Cast(&contactsViewBirthdayRow)
		b.GetObject("contacts_view_address").Cast(&contactsViewAddressRow)
		b.GetObject("contacts_view_notes").Cast(&contactsViewNotesRow)
		b.GetObject("contacts_view_debts").Cast(&contactsViewDebtsListBox)
		b.GetObject("contacts_view_activities").Cast(&contactsViewActivitiesListBox)
		b.GetObject("activities_view_page_title").Cast(&activitiesViewPageTitle)
		b.GetObject("activities_view_stack").Cast(&activitiesViewStack)
		b.GetObject("activities_view_error_status_page").Cast(&activitiesViewErrorStatusPage)
		b.GetObject("activities_view_error_refresh_button").Cast(&activitiesViewErrorRefreshButton)
		b.GetObject("activities_view_error_copy_details").Cast(&activitiesViewErrorCopyDetailsButton)
		b.GetObject("activities_view_edit_button").Cast(&activitiesViewEditButton)
		activitiesViewEditButton.SetActionTargetValue(glib.NewVariantInt64(0))
		b.GetObject("activities_view_delete_button").Cast(&activitiesViewDeleteButton)
		activitiesViewDeleteButton.SetActionTargetValue(glib.NewVariantInt64(0))
		b.GetObject("activities_view_body").Cast(&activitiesViewPageBodyWebView)
		{
			// Set transparent background for WebView
			bg := gdk.RGBA{Alpha: 0}
			activitiesViewPageBodyWebView.SetBackgroundColor(&bg)

			// Handle navigation policy - open external links in browser
			onDecidePolicy := func(_ webkit.WebView, decisionPtr uintptr, decisionType webkit.PolicyDecisionType) bool {
				if decisionType == webkit.PolicyDecisionTypeNavigationActionValue {
					decision := webkit.NavigationPolicyDecisionNewFromInternalPtr(decisionPtr)
					u, err := url.Parse(decision.GetNavigationAction().GetRequest().GetUri())
					if err != nil {
						log.Warn("Could not parse activity view WebView URL", "err", err)
						return true
					}

					openExternally := u.Scheme != "about"
					log.Debug("Handling navigation in activity view WebView", "openExternally", openExternally, "url", u.String())

					if openExternally {
						gtk.ShowUri(&w.ApplicationWindow.Window, u.String(), uint32(gdk.CURRENT_TIME))
						return true
					}

					return false
				}
				return false
			}
			activitiesViewPageBodyWebView.ConnectDecidePolicy(&onDecidePolicy)
		}
		b.GetObject("activities_edit_page_title").Cast(&activitiesEditPageTitle)
		b.GetObject("activities_edit_stack").Cast(&activitiesEditStack)
		b.GetObject("activities_edit_error_status_page").Cast(&activitiesEditErrorStatusPage)
		b.GetObject("activities_edit_error_refresh_button").Cast(&activitiesEditErrorRefreshButton)
		b.GetObject("activities_edit_error_copy_details").Cast(&activitiesEditErrorCopyDetailsButton)
		b.GetObject("activities_edit_save_button").Cast(&activitiesEditPageSaveButton)
		b.GetObject("activities_edit_save_spinner").Cast(&activitiesEditPageSaveSpinner)
		b.GetObject("activities_edit_page_name_input").Cast(&activitiesEditPageNameInput)
		b.GetObject("activities_edit_page_date_input").Cast(&activitiesEditPageDateInput)
		b.GetObject("activities_edit_page_description_expander").Cast(&activitiesEditPageDescriptionExpander)
		b.GetObject("activities_edit_page_description_input").Cast(&activitiesEditPageDescriptionInput)
		b.GetObject("activities_edit_page_date_warning_button").Cast(&activitiesEditPageDateWarningButton)
		b.GetObject("activities_edit_page_date_popover_label").Cast(&activitiesEditPagePopoverLabel)
		b.GetObject("debts_edit_page_title").Cast(&debtsEditPageTitle)
		b.GetObject("debts_edit_stack").Cast(&debtsEditStack)
		b.GetObject("debts_edit_error_status_page").Cast(&debtsEditErrorStatusPage)
		b.GetObject("debts_edit_error_refresh_button").Cast(&debtsEditErrorRefreshButton)
		b.GetObject("debts_edit_error_copy_details").Cast(&debtsEditErrorCopyDetailsButton)
		b.GetObject("debts_edit_save_button").Cast(&debtsEditPageSaveButton)
		b.GetObject("debts_edit_save_spinner").Cast(&debtsEditPageSaveSpinner)
		b.GetObject("debts_edit_page_you_owe_radio").Cast(&debtsEditPageYouOweRadio)
		b.GetObject("debts_edit_page_amount_input").Cast(&debtsEditPageAmountInput)
		b.GetObject("debts_edit_page_currency_input").Cast(&debtsEditPageCurrencyInput)
		b.GetObject("debts_edit_page_description_expander").Cast(&debtsEditPageDescriptionExpander)
		b.GetObject("debts_edit_page_description_input").Cast(&debtsEditPageDescriptionInput)
		b.GetObject("debts_edit_page_debt_type_you_owe_row").Cast(&debtsEditPageYouOweActionRow)
		b.GetObject("debts_edit_page_debt_type_they_owe_row").Cast(&debtsEditPageTheyOweActionRow)
		b.GetObject("contacts_edit_page_title").Cast(&contactsEditPageTitle)
		b.GetObject("contacts_edit_stack").Cast(&contactsEditStack)
		b.GetObject("contacts_edit_error_status_page").Cast(&contactsEditErrorStatusPage)
		b.GetObject("contacts_edit_error_refresh_button").Cast(&contactsEditErrorRefreshButton)
		b.GetObject("contacts_edit_error_copy_details").Cast(&contactsEditErrorCopyDetailsButton)
		b.GetObject("contacts_edit_save_button").Cast(&contactsEditPageSaveButton)
		b.GetObject("contacts_edit_save_spinner").Cast(&contactsEditPageSaveSpinner)
		b.GetObject("contacts_edit_page_first_name_input").Cast(&contactsEditPageFirstNameInput)
		b.GetObject("contacts_edit_page_last_name_input").Cast(&contactsEditPageLastNameInput)
		b.GetObject("contacts_edit_page_nickname_input").Cast(&contactsEditPageNicknameInput)
		b.GetObject("contacts_edit_page_email_input").Cast(&contactsEditPageEmailInput)
		b.GetObject("contacts_edit_page_pronouns_input").Cast(&contactsEditPagePronounsInput)
		b.GetObject("contacts_edit_page_birthday_input").Cast(&contactsEditPageBirthdayInput)
		b.GetObject("contacts_edit_page_address_expander").Cast(&contactsEditPageAddressExpander)
		b.GetObject("contacts_edit_page_address_input").Cast(&contactsEditPageAddressInput)
		b.GetObject("contacts_edit_page_notes_expander").Cast(&contactsEditPageNotesExpander)
		b.GetObject("contacts_edit_page_notes_input").Cast(&contactsEditPageNotesInput)
		b.GetObject("contacts_edit_page_email_warning_button").Cast(&contactsEditPageEmailWarningButton)
		b.GetObject("contacts_edit_page_birthday_warning_button").Cast(&contactsEditPageBirthdayWarningButton)
		b.GetObject("contacts_edit_page_birthday_popover_label").Cast(&contactsEditPagePopoverLabel)
		debtsCreateDialogBuilder.GetObject("debts_create_dialog").Cast(&debtsCreateDialog)
		debtsCreateDialogBuilder.GetObject("debts_create_dialog_add_button").Cast(&debtsCreateDialogAddButton)
		debtsCreateDialogBuilder.GetObject("debts_create_dialog_add_spinner").Cast(&debtsCreateDialogAddSpinner)
		debtsCreateDialogBuilder.GetObject("debts_create_dialog_title").Cast(&debtsCreateDialogTitle)
		debtsCreateDialogBuilder.GetObject("debts_create_dialog_you_owe_radio").Cast(&debtsCreateDialogYouOweRadio)
		debtsCreateDialogBuilder.GetObject("debts_create_dialog_amount_input").Cast(&debtsCreateDialogAmountInput)
		debtsCreateDialogBuilder.GetObject("debts_create_dialog_currency_input").Cast(&debtsCreateDialogCurrencyInput)
		debtsCreateDialogBuilder.GetObject("debts_create_dialog_description_expander").Cast(&debtsCreateDialogDescriptionExpander)
		debtsCreateDialogBuilder.GetObject("debts_create_dialog_description_input").Cast(&debtsCreateDialogDescriptionInput)
		debtsCreateDialogBuilder.GetObject("debts_create_dialog_debt_type_you_owe_row").Cast(&debtsCreateDialogYouOweActionRow)
		debtsCreateDialogBuilder.GetObject("debts_create_dialog_debt_type_they_owe_row").Cast(&debtsCreateDialogTheyOweActionRow)
		activitiesCreateDialogBuilder.GetObject("activities_create_dialog").Cast(&activitiesCreateDialog)
		activitiesCreateDialogBuilder.GetObject("activities_create_dialog_add_button").Cast(&activitiesCreateDialogAddButton)
		activitiesCreateDialogBuilder.GetObject("activities_create_dialog_add_spinner").Cast(&activitiesCreateDialogAddSpinner)
		activitiesCreateDialogBuilder.GetObject("activities_create_dialog_title").Cast(&activitiesCreateDialogTitle)
		activitiesCreateDialogBuilder.GetObject("activities_create_dialog_name_input").Cast(&activitiesCreateDialogNameInput)
		activitiesCreateDialogBuilder.GetObject("activities_create_dialog_date_input").Cast(&activitiesCreateDialogDateInput)
		activitiesCreateDialogBuilder.GetObject("activities_create_dialog_description_expander").Cast(&activitiesCreateDialogDescriptionExpander)
		activitiesCreateDialogBuilder.GetObject("activities_create_dialog_description_input").Cast(&activitiesCreateDialogDescriptionInput)
		activitiesCreateDialogBuilder.GetObject("activities_create_dialog_date_warning_button").Cast(&activitiesCreateDialogDateWarningButton)
		activitiesCreateDialogBuilder.GetObject("activities_create_dialog_date_popover_label").Cast(&activitiesCreateDialogPopoverLabel)
		b.GetObject("journal_entries_stack").Cast(&journalEntriesStack)
		b.GetObject("journal_entries_list").Cast(&journalEntriesListBox)
		b.GetObject("journal_entries_searchentry").Cast(&journalEntriesSearchEntry)
		b.GetObject("journal_entries_add_button").Cast(&journalEntriesAddButton)
		b.GetObject("journal_entries_search_button").Cast(&journalEntriesSearchButton)
		b.GetObject("journal_entries_empty_add_button").Cast(&journalEntriesEmptyAddButton)
		b.GetObject("journal_entries_error_status_page").Cast(&journalEntriesErrorStatusPage)
		b.GetObject("journal_entries_error_refresh_button").Cast(&journalEntriesErrorRefreshButton)
		b.GetObject("journal_entries_error_copy_details").Cast(&journalEntriesErrorCopyDetailsButton)
		journalEntriesCreateDialogBuilder.GetObject("journal_entries_create_dialog").Cast(&journalEntriesCreateDialog)
		journalEntriesCreateDialogBuilder.GetObject("journal_entries_create_dialog_add_button").Cast(&journalEntriesCreateDialogAddButton)
		journalEntriesCreateDialogBuilder.GetObject("journal_entries_create_dialog_add_spinner").Cast(&journalEntriesCreateDialogAddSpinner)
		journalEntriesCreateDialogBuilder.GetObject("journal_entries_create_dialog_rating").Cast(&journalEntriesCreateDialogRatingToggleGroup)
		journalEntriesCreateDialogBuilder.GetObject("journal_entries_create_dialog_title_input").Cast(&journalEntriesCreateDialogTitleInput)
		journalEntriesCreateDialogBuilder.GetObject("journal_entries_create_dialog_body_expander").Cast(&journalEntriesCreateDialogBodyExpander)
		journalEntriesCreateDialogBuilder.GetObject("journal_entries_create_dialog_body_input").Cast(&journalEntriesCreateDialogBodyInput)
		b.GetObject("journal_entries_view_page_title").Cast(&journalEntriesViewPageTitle)
		b.GetObject("journal_entries_view_stack").Cast(&journalEntriesViewStack)
		b.GetObject("journal_entries_view_error_status_page").Cast(&journalEntriesViewErrorStatusPage)
		b.GetObject("journal_entries_view_error_refresh_button").Cast(&journalEntriesViewErrorRefreshButton)
		b.GetObject("journal_entries_view_error_copy_details").Cast(&journalEntriesViewErrorCopyDetailsButton)
		b.GetObject("journal_entries_view_edit_button").Cast(&journalEntriesViewEditButton)
		journalEntriesViewEditButton.SetActionTargetValue(glib.NewVariantInt64(0))
		b.GetObject("journal_entries_view_delete_button").Cast(&journalEntriesViewDeleteButton)
		journalEntriesViewDeleteButton.SetActionTargetValue(glib.NewVariantInt64(0))
		b.GetObject("journal_entries_view_body").Cast(&journalEntriesViewPageBodyWebView)

		bg := gdk.RGBA{Alpha: 0}
		journalEntriesViewPageBodyWebView.SetBackgroundColor(&bg)

		onDecidePolicy := func(_ webkit.WebView, decisionPtr uintptr, decisionType webkit.PolicyDecisionType) bool {
			if decisionType == webkit.PolicyDecisionTypeNavigationActionValue {
				decision := webkit.NavigationPolicyDecisionNewFromInternalPtr(decisionPtr)
				u, err := url.Parse(decision.GetNavigationAction().GetRequest().GetUri())
				if err != nil {
					log.Warn("Could not parse journal entry view WebView URL", "err", err)
					return true
				}

				openExternally := u.Scheme != "about"
				log.Debug("Handling navigation in journal entry view WebView", "openExternally", openExternally, "url", u.String())

				if openExternally {
					gtk.ShowUri(&w.ApplicationWindow.Window, u.String(), uint32(gdk.CURRENT_TIME))
					return true
				}

				return false
			}
			return false
		}
		journalEntriesViewPageBodyWebView.ConnectDecidePolicy(&onDecidePolicy)

		b.GetObject("journal_entries_edit_page_title").Cast(&journalEntriesEditPageTitle)
		b.GetObject("journal_entries_edit_stack").Cast(&journalEntriesEditStack)
		b.GetObject("journal_entries_edit_error_status_page").Cast(&journalEntriesEditErrorStatusPage)
		b.GetObject("journal_entries_edit_error_refresh_button").Cast(&journalEntriesEditErrorRefreshButton)
		b.GetObject("journal_entries_edit_error_copy_details").Cast(&journalEntriesEditErrorCopyDetailsButton)
		b.GetObject("journal_entries_edit_save_button").Cast(&journalEntriesEditPageSaveButton)
		b.GetObject("journal_entries_edit_save_spinner").Cast(&journalEntriesEditPageSaveSpinner)
		b.GetObject("journal_entries_edit_page_rating").Cast(&journalEntriesEditPageRatingToggleGroup)
		b.GetObject("journal_entries_edit_page_title_input").Cast(&journalEntriesEditPageTitleInput)
		b.GetObject("journal_entries_edit_page_body_expander").Cast(&journalEntriesEditPageBodyExpander)
		b.GetObject("journal_entries_edit_page_body_input").Cast(&journalEntriesEditPageBodyInput)

		settings.Bind(resources.SettingVerboseKey, &preferencesDialogVerboseSwitch.Object, "active", gio.GSettingsBindDefaultValue)

		setValidationSuffixVisible := func(input *adw.EntryRow, suffix *gtk.MenuButton, visible bool) {
			if visible && suffix.GetParent() == nil {
				input.AddSuffix(&suffix.Widget)
				input.AddCssClass("error")
			} else if !visible && suffix.GetParent() != nil {
				input.RemoveCssClass("error")
				input.Remove(&suffix.Widget)
			}
		}

		onWelcomeGetStartedClicked := func(_ gtk.Button) {
			nv.PushByTag(resources.PageConfigServerURL)
		}
		welcomeGetStartedButton.ConnectClicked(&onWelcomeGetStartedClicked)

		settings.Bind(resources.SettingServerURLKey, &configServerURLInput.PreferencesRow.ListBoxRow.Widget.InitiallyUnowned.Object, "text", gio.GSettingsBindDefaultValue)

		updateConfigServerURLContinueButtonSensitive := func() {
			if len(settings.GetString(resources.SettingServerURLKey)) > 0 {
				configServerURLContinueButton.SetSensitive(true)
			} else {
				configServerURLContinueButton.SetSensitive(false)
			}
		}

		parseLocaleDate := func(localeDate string) (time.Time, error) {
			return time.Parse(glibDateTimeFromGo(time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC)).Format("%x"), localeDate)
		}

		invalidDateLabel := fmt.Sprintf("Not a valid date (format: %v)", glibDateTimeFromGo(time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC)).Format("%x"))
		activitiesCreateDialogPopoverLabel.SetLabel(L(invalidDateLabel))
		activitiesEditPagePopoverLabel.SetLabel(L(invalidDateLabel))
		contactsEditPagePopoverLabel.SetLabel(L(invalidDateLabel))

		var deregistrationLock sync.Mutex
		deregisterOIDCClient := func() error {
			deregistrationLock.Lock()
			defer deregistrationLock.Unlock()

			if registrationClientURI := settings.GetString(resources.SettingRegistrationClientURIKey); registrationClientURI != "" {
				registrationAccessToken, err := keyring.Get(resources.AppID, resources.SecretRegistrationAccessToken)
				if err != nil {
					return err
				}

				if err := authn.DeregisterOIDCClient(
					ctx,

					slog.New(log.Handler().WithGroup("oidcDeregistration")),

					registrationAccessToken,
					registrationClientURI,
				); err != nil {
					return err
				}

				if ok := settings.SetString(resources.SettingRegistrationClientURIKey, ""); !ok {
					return errCouldNotWriteSettingsKey
				}

				if err := keyring.Delete(resources.AppID, resources.SecretRegistrationAccessToken); err != nil && !errors.Is(err, keyring.ErrNotFound) {
					return err
				}
			}

			// We indiscriminately clear the client ID, even if the client was never registered
			// via OIDC dynamic client registration so that we can switch Senbara servers (which
			// configure different OIDC endpoints and thus expect different OIDC client IDs) properly
			if ok := settings.SetString(resources.SettingOIDCClientIDKey, ""); !ok {
				return errCouldNotWriteSettingsKey
			}

			return nil
		}

		openPreferencesAction := gio.NewSimpleAction("openPreferences", nil)
		connectSimpleActionActivate(openPreferencesAction, func() {
			preferencesDialog.Present(&w.ApplicationWindow.Window.Widget)
		})
		a.SetAccelsForAction("app.openPreferences", []string{`<Primary>comma`})
		a.AddAction(openPreferencesAction)

		deregisterClientAction := gio.NewSimpleAction("deregisterClient", nil)

		updateDeregisterClientActionEnabled := func() {
			deregisterClientAction.SetEnabled(settings.GetString(resources.SettingOIDCClientIDKey) != "")
		}

		connectSimpleActionActivate(deregisterClientAction, func() {
			configServerURLContinueButton.SetSensitive(false)
			welcomeGetStartedButton.SetSensitive(false)
			configServerURLContinueSpinner.SetVisible(true)
			welcomeGetStartedSpinner.SetVisible(true)

			go func() {
				defer welcomeGetStartedButton.SetSensitive(true)
				defer configServerURLContinueSpinner.SetVisible(false)
				defer welcomeGetStartedSpinner.SetVisible(false)

				if err := deregisterOIDCClient(); err != nil {
					onPanic(err)

					return
				}

				updateDeregisterClientActionEnabled()
				updateConfigServerURLContinueButtonSensitive()
			}()
		})
		a.AddAction(deregisterClientAction)

		onSetLogLevel := func(verbose bool) {
			if verbose {
				level.Set(slog.LevelDebug)
			} else {
				level.Set(slog.LevelInfo)
			}
		}
		onSetLogLevel(settings.GetBoolean(resources.SettingVerboseKey))

		connectSettingsChanged(settings, func(key string) {
			switch key {
			case resources.SettingVerboseKey:
				onSetLogLevel(settings.GetBoolean(resources.SettingVerboseKey))

			case resources.SettingServerURLKey:
				configServerURLContinueButton.SetSensitive(false)
				configServerURLContinueSpinner.SetVisible(true)

				go func() {
					defer configServerURLContinueSpinner.SetVisible(false)

					if err := deregisterOIDCClient(); err != nil {
						onPanic(err)

						return
					}

					updateDeregisterClientActionEnabled()
					updateConfigServerURLContinueButtonSensitive()
				}()
			}
		})

		checkSenbaraServerConfiguration := func() error {
			var (
				serverURL = settings.GetString(resources.SettingServerURLKey)
			)

			client, err := api.NewClientWithResponses(serverURL)
			if err != nil {
				return err
			}

			log.Debug("Getting OpenAPI spec")

			res, err := client.GetOpenAPISpec(ctx)
			if err != nil {
				return err
			}

			log.Debug("Got OpenAPI spec", "status", res.StatusCode)

			if res.StatusCode != http.StatusOK {
				return errors.New(res.Status)
			}

			if err := json.NewDecoder(res.Body).Decode(&spec); err != nil {
				return err
			}

			return nil
		}

		setupAuthn := func(registerClient bool) error {
			o, err := authn.DiscoverOIDCProviderConfiguration(
				ctx,

				slog.New(log.Handler().WithGroup("oidcDiscovery")),

				spec.Components.SecuritySchemes["oidc"].Value.OpenIdConnectUrl,
			)
			if err != nil {
				return err
			}

			oidcClientID := settings.GetString(resources.SettingOIDCClientIDKey)
			if oidcClientID == "" && registerClient {
				c, err := authn.RegisterOIDCClient(
					ctx,

					slog.New(log.Handler().WithGroup("oidcRegistration")),

					o,

					"Senbara GNOME",
					redirectURL,

					configInitialAccessTokenInput.GetText(),
				)
				if err != nil {
					return err
				}

				if ok := settings.SetString(resources.SettingOIDCClientIDKey, c.ClientID); !ok {
					return errCouldNotWriteSettingsKey
				}

				if ok := settings.SetString(resources.SettingRegistrationClientURIKey, c.RegistrationClientURI); !ok {
					return errCouldNotWriteSettingsKey
				}

				if err := keyring.Set(resources.AppID, resources.SecretRegistrationAccessToken, c.RegistrationAccessToken); err != nil {
					return err
				}

				oidcClientID = c.ClientID
			}

			updateDeregisterClientActionEnabled()

			authner = authn.NewAuthner(
				slog.New(log.Handler().WithGroup("authner")),

				o.Issuer,
				o.EndSessionEndpoint,

				oidcClientID,
				redirectURL,
			)

			if err := authner.Init(ctx); err != nil {
				return err
			}

			return nil
		}

		connectButtonClicked(&configServerURLContinueButton, func() {
			configServerURLContinueButton.SetSensitive(false)
			configServerURLContinueSpinner.SetVisible(true)

			go func() {
				defer configServerURLContinueSpinner.SetVisible(false)

				if err := checkSenbaraServerConfiguration(); err != nil {
					onPanic(err)

					return
				}

				nv.PushByTag(resources.PagePreview)
			}()
		})

		connectButtonClicked(&previewLoginButton, func() {
			previewLoginButton.SetSensitive(false)
			previewLoginSpinner.SetVisible(true)

			go func() {
				defer previewLoginButton.SetSensitive(true)
				defer previewLoginSpinner.SetVisible(false)

				if err := checkSenbaraServerConfiguration(); err != nil {
					onPanic(err)

					return
				}

				if err := setupAuthn(true); err != nil {
					onPanic(err)

					return
				}

				if v := spec.Components.SecuritySchemes["oidc"].Value.Extensions[api.OidcDcrInitialAccessTokenPortalUrlExtensionKey]; v != nil {
					vv, ok := v.(string)
					if ok {
						oidcDcrInitialAccessTokenPortalUrl = vv

						nv.PushByTag(resources.PageRegister)

						return
					}
				}

				nv.PushByTag(resources.PageHome)
			}()
		})

		connectButtonClicked(&registerRegisterButton, func() {
			go func() {
				if _, err := gio.AppInfoLaunchDefaultForUri(oidcDcrInitialAccessTokenPortalUrl, nil); err != nil {
					onPanic(err)

					return
				}

				nv.PushByTag(resources.PageConfigInitialAccessToken)
			}()
		})

		connectPasswordEntryRowChanged(&configInitialAccessTokenInput, func() {
			if configInitialAccessTokenInput.GetTextLength() > 0 {
				configInitialAccessTokenLoginButton.SetSensitive(true)
			} else {
				configInitialAccessTokenLoginButton.SetSensitive(false)
			}
		})

		connectButtonClicked(&configInitialAccessTokenLoginButton, func() {
			configInitialAccessTokenLoginButton.SetSensitive(false)
			configInitialAccessTokenLoginSpinner.SetVisible(true)

			go func() {
				defer configInitialAccessTokenLoginButton.SetSensitive(true)
				defer configInitialAccessTokenLoginSpinner.SetVisible(false)

				if err := checkSenbaraServerConfiguration(); err != nil {
					onPanic(err)

					return
				}

				if err := setupAuthn(true); err != nil {
					onPanic(err)

					return
				}

				nv.PushByTag(resources.PageHome)
			}()
		})

		selectDifferentServerAction := gio.NewSimpleAction("selectDifferentServer", nil)
		connectSimpleActionActivate(selectDifferentServerAction, func() {
			nv.ReplaceWithTags([]string{resources.PageWelcome}, 1)
		})
		a.AddAction(selectDifferentServerAction)

		connectButtonClicked(&exchangeLoginCancelButton, func() {
			nv.ReplaceWithTags([]string{resources.PageWelcome}, 1)
		})

		connectButtonClicked(&exchangeLogoutCancelButton, func() {
			nv.ReplaceWithTags([]string{resources.PageHome}, 1)
		})

		enableHomeUserMenuLoading := func() {
			homeUserMenuButton.SetSensitive(false)
			homeUserMenuAvatar.SetVisible(false)
			homeUserMenuSpinner.SetVisible(true)
		}

		disableHomeUserMenuLoading := func() {
			homeUserMenuSpinner.SetVisible(false)
			homeUserMenuAvatar.SetVisible(true)
			homeUserMenuButton.SetSensitive(true)
		}

		enableHomeHamburgerMenuLoading := func() {
			homeHamburgerMenuButton.SetSensitive(false)
			homeHamburgerMenuIcon.SetVisible(false)
			homeHamburgerMenuSpinner.SetVisible(true)
		}

		disableHomeHamburgerMenuLoading := func() {
			homeHamburgerMenuSpinner.SetVisible(false)
			homeHamburgerMenuIcon.SetVisible(true)
			homeHamburgerMenuButton.SetSensitive(true)
		}

		enableHomeSidebarLoading := func() {
			homeSidebarContactsCountLabel.SetVisible(false)
			homeSidebarContactsCountSpinner.SetVisible(true)

			homeSidebarJournalEntriesCountLabel.SetVisible(false)
			homeSidebarJournalEntriesCountSpinner.SetVisible(true)
		}

		disableHomeSidebarLoading := func() {
			homeSidebarJournalEntriesCountSpinner.SetVisible(false)
			homeSidebarJournalEntriesCountLabel.SetVisible(true)

			homeSidebarContactsCountSpinner.SetVisible(false)
			homeSidebarContactsCountLabel.SetVisible(true)
		}

		enablePreviewLoading := func() {
			previewContactsCountLabel.SetVisible(false)
			previewContactsCountSpinner.SetVisible(true)

			previewJournalEntriesCountLabel.SetVisible(false)
			previewJournalEntriesCountSpinner.SetVisible(true)
		}

		disablePreviewLoading := func() {
			previewJournalEntriesCountSpinner.SetVisible(false)
			previewJournalEntriesCountLabel.SetVisible(true)

			previewContactsCountSpinner.SetVisible(false)
			previewContactsCountLabel.SetVisible(true)
		}

		var (
			contactsCount        = 0
			visibleContactsCount = 0
		)

		connectSearchEntryChanged(&contactsSearchEntry, func() {
			go func() {
				contactsStack.SetVisibleChildName(resources.PageContactsLoading)

				visibleContactsCount = 0

				contactsListBox.InvalidateFilter()

				if visibleContactsCount > 0 {
					contactsStack.SetVisibleChildName(resources.PageContactsList)
				} else {
					contactsStack.SetVisibleChildName(resources.PageContactsNoResults)
				}
			}()
		})

		var (
			journalEntriesCount        = 0
			visibleJournalEntriesCount = 0
		)

		connectSearchEntryChanged(&journalEntriesSearchEntry, func() {
			go func() {
				journalEntriesStack.SetVisibleChildName(resources.PageJournalEntriesLoading)

				visibleJournalEntriesCount = 0

				journalEntriesListBox.InvalidateFilter()

				if visibleJournalEntriesCount > 0 {
					journalEntriesStack.SetVisibleChildName(resources.PageJournalEntriesList)
				} else {
					journalEntriesStack.SetVisibleChildName(resources.PageJournalEntriesNoResults)
				}
			}()
		})

		setListBoxFilterFunc(&journalEntriesListBox, func(row *gtk.ListBoxRow) bool {
			var r adw.ActionRow
			row.Cast(&r)
			var (
				f = strings.ToLower(journalEntriesSearchEntry.GetText())

				rt = strings.ToLower(r.PreferencesRow.GetTitle())
				rs = strings.ToLower(r.GetSubtitle())
			)

			log.Debug(
				"Filtering journal entry",
				"filter", f,
				"title", rt,
				"subtitle", rs,
			)

			if strings.Contains(rt, f) || strings.Contains(rs, f) {
				visibleJournalEntriesCount++

				return true
			}

			return false
		})

		connectButtonClicked(&contactsAddButton, func() {
			contactsCreateDialog.Present(&w.ApplicationWindow.Window.Widget)
			contactsCreateDialogFirstNameInput.GrabFocus()
		})

		connectButtonClicked(&contactsEmptyAddButton, func() {
			contactsCreateDialog.Present(&w.ApplicationWindow.Window.Widget)
			contactsCreateDialogFirstNameInput.GrabFocus()
		})

		onValidateContactsCreateDialogForm := func() {
			if email := contactsCreateDialogEmailInput.GetText(); email != "" {
				if _, err := mail.ParseAddress(email); err != nil {
					setValidationSuffixVisible(&contactsCreateDialogEmailInput, &contactsCreateDialogEmailWarningButton, true)

					contactsCreateDialogAddButton.SetSensitive(false)

					return
				}
			}

			setValidationSuffixVisible(&contactsCreateDialogEmailInput, &contactsCreateDialogEmailWarningButton, false)

			if contactsCreateDialogFirstNameInput.GetText() != "" &&
				contactsCreateDialogLastNameInput.GetText() != "" &&
				contactsCreateDialogEmailInput.GetText() != "" &&
				contactsCreateDialogPronounsInput.GetText() != "" {
				contactsCreateDialogAddButton.SetSensitive(true)
			} else {
				contactsCreateDialogAddButton.SetSensitive(false)
			}
		}

		connectEntryRowChanged(&contactsCreateDialogFirstNameInput, onValidateContactsCreateDialogForm)
		connectEntryRowChanged(&contactsCreateDialogLastNameInput, onValidateContactsCreateDialogForm)
		connectEntryRowChanged(&contactsCreateDialogNicknameInput, onValidateContactsCreateDialogForm)
		connectEntryRowChanged(&contactsCreateDialogEmailInput, onValidateContactsCreateDialogForm)
		connectEntryRowChanged(&contactsCreateDialogPronounsInput, onValidateContactsCreateDialogForm)

		connectDialogClosed(&contactsCreateDialog, func() {
			contactsCreateDialogFirstNameInput.SetText("")
			contactsCreateDialogLastNameInput.SetText("")
			contactsCreateDialogNicknameInput.SetText("")
			contactsCreateDialogEmailInput.SetText("")
			contactsCreateDialogPronounsInput.SetText("")

			setValidationSuffixVisible(&contactsCreateDialogEmailInput, &contactsCreateDialogEmailWarningButton, false)
		})

		onValidateDebtsCreateDialogForm := func() {
			if debtsCreateDialogCurrencyInput.GetText() != "" {
				debtsCreateDialogAddButton.SetSensitive(true)
			} else {
				debtsCreateDialogAddButton.SetSensitive(false)
			}
		}

		connectEntryRowChanged(&debtsCreateDialogCurrencyInput, onValidateDebtsCreateDialogForm)

		onValidateDebtsEditPageForm := func() {
			if debtsEditPageCurrencyInput.GetText() != "" {
				debtsEditPageSaveButton.SetSensitive(true)
			} else {
				debtsEditPageSaveButton.SetSensitive(false)
			}
		}

		connectEntryRowChanged(&debtsEditPageCurrencyInput, onValidateDebtsEditPageForm)

		connectDialogClosed(&debtsCreateDialog, func() {
			debtsCreateDialogTitle.SetSubtitle("")

			debtsCreateDialogYouOweActionRow.SetTitle("")
			debtsCreateDialogTheyOweActionRow.SetTitle("")

			debtsCreateDialogYouOweRadio.SetActive(true)

			debtsCreateDialogAmountInput.SetValue(0)
			debtsCreateDialogCurrencyInput.SetText("")

			debtsCreateDialogDescriptionExpander.SetExpanded(false)
			debtsCreateDialogDescriptionInput.GetBuffer().SetText("", 0)
		})

		onValidateActivitiesCreateDialogForm := func() {
			if date := activitiesCreateDialogDateInput.GetText(); date != "" {
				if _, err := parseLocaleDate(date); err != nil {
					setValidationSuffixVisible(&activitiesCreateDialogDateInput, &activitiesCreateDialogDateWarningButton, true)

					activitiesCreateDialogAddButton.SetSensitive(false)

					return
				}
			}

			setValidationSuffixVisible(&activitiesCreateDialogDateInput, &activitiesCreateDialogDateWarningButton, false)

			if activitiesCreateDialogNameInput.GetText() != "" &&
				activitiesCreateDialogDateInput.GetText() != "" {
				activitiesCreateDialogAddButton.SetSensitive(true)
			} else {
				activitiesCreateDialogAddButton.SetSensitive(false)
			}
		}

		connectEntryRowChanged(&activitiesCreateDialogNameInput, onValidateActivitiesCreateDialogForm)
		connectEntryRowChanged(&activitiesCreateDialogDateInput, onValidateActivitiesCreateDialogForm)

		connectDialogClosed(&activitiesCreateDialog, func() {
			activitiesCreateDialogNameInput.SetText("")
			activitiesCreateDialogDateInput.SetText("")

			setValidationSuffixVisible(&activitiesCreateDialogDateInput, &activitiesCreateDialogDateWarningButton, false)

			activitiesCreateDialogDescriptionExpander.SetExpanded(false)
			activitiesCreateDialogDescriptionInput.GetBuffer().SetText("", 0)
		})

		onValidateActivitiesEditPageForm := func() {
			if date := activitiesEditPageDateInput.GetText(); date != "" {
				if _, err := parseLocaleDate(date); err != nil {
					setValidationSuffixVisible(&activitiesEditPageDateInput, &activitiesEditPageDateWarningButton, true)

					activitiesEditPageSaveButton.SetSensitive(false)

					return
				}
			}

			setValidationSuffixVisible(&activitiesEditPageDateInput, &activitiesEditPageDateWarningButton, false)

			if activitiesEditPageNameInput.GetText() != "" &&
				activitiesEditPageDateInput.GetText() != "" {
				activitiesEditPageSaveButton.SetSensitive(true)
			} else {
				activitiesEditPageSaveButton.SetSensitive(false)
			}
		}

		connectEntryRowChanged(&activitiesEditPageNameInput, onValidateActivitiesEditPageForm)
		connectEntryRowChanged(&activitiesEditPageDateInput, onValidateActivitiesEditPageForm)

		onValidateContactsEditPageForm := func() {
			if email := contactsEditPageEmailInput.GetText(); email != "" {
				if _, err := mail.ParseAddress(email); err != nil {
					setValidationSuffixVisible(&contactsEditPageEmailInput, &contactsEditPageEmailWarningButton, true)

					contactsEditPageSaveButton.SetSensitive(false)

					return
				}
			}

			setValidationSuffixVisible(&contactsEditPageEmailInput, &contactsEditPageEmailWarningButton, false)

			if date := contactsEditPageBirthdayInput.GetText(); date != "" {
				if _, err := parseLocaleDate(date); err != nil {
					setValidationSuffixVisible(&contactsEditPageBirthdayInput, &contactsEditPageBirthdayWarningButton, true)

					contactsEditPageSaveButton.SetSensitive(false)

					return
				}
			}

			setValidationSuffixVisible(&contactsEditPageBirthdayInput, &contactsEditPageBirthdayWarningButton, false)

			if contactsEditPageFirstNameInput.GetText() != "" &&
				contactsEditPageLastNameInput.GetText() != "" &&
				contactsEditPageEmailInput.GetText() != "" &&
				contactsEditPagePronounsInput.GetText() != "" {
				contactsEditPageSaveButton.SetSensitive(true)
			} else {
				contactsEditPageSaveButton.SetSensitive(false)
			}
		}

		connectEntryRowChanged(&contactsEditPageFirstNameInput, onValidateContactsEditPageForm)
		connectEntryRowChanged(&contactsEditPageLastNameInput, onValidateContactsEditPageForm)
		connectEntryRowChanged(&contactsEditPageNicknameInput, onValidateContactsEditPageForm)
		connectEntryRowChanged(&contactsEditPageEmailInput, onValidateContactsEditPageForm)
		connectEntryRowChanged(&contactsEditPagePronounsInput, onValidateContactsEditPageForm)

		connectEntryRowChanged(&contactsEditPageBirthdayInput, onValidateContactsEditPageForm)

		connectButtonClicked(&journalEntriesAddButton, func() {
			journalEntriesCreateDialog.Present(&w.ApplicationWindow.Window.Widget)
			journalEntriesCreateDialogTitleInput.GrabFocus()
		})

		connectButtonClicked(&journalEntriesEmptyAddButton, func() {
			journalEntriesCreateDialog.Present(&w.ApplicationWindow.Window.Widget)
			journalEntriesCreateDialogTitleInput.GrabFocus()
		})

		onValidateJournalEntriesCreateDialogForm := func() {
			if journalEntriesCreateDialogTitleInput.GetText() != "" &&
				getTextBufferText(journalEntriesCreateDialogBodyInput.GetBuffer()) != "" {
				journalEntriesCreateDialogAddButton.SetSensitive(true)
			} else {
				journalEntriesCreateDialogAddButton.SetSensitive(false)
			}
		}

		connectEntryRowChanged(&journalEntriesCreateDialogTitleInput, onValidateJournalEntriesCreateDialogForm)
		connectTextBufferChanged(journalEntriesCreateDialogBodyInput.GetBuffer(), onValidateJournalEntriesCreateDialogForm)

		connectDialogClosed(&journalEntriesCreateDialog, func() {
			journalEntriesCreateDialogRatingToggleGroup.SetActive(0)

			journalEntriesCreateDialogTitleInput.SetText("")

			journalEntriesCreateDialogBodyExpander.SetExpanded(true)
			journalEntriesCreateDialogBodyInput.GetBuffer().SetText("", 0)
		})

		onValidateJournalEntriesEditPageForm := func() {
			if journalEntriesEditPageTitleInput.GetText() != "" &&
				getTextBufferText(journalEntriesEditPageBodyInput.GetBuffer()) != "" {
				journalEntriesEditPageSaveButton.SetSensitive(true)
			} else {
				journalEntriesEditPageSaveButton.SetSensitive(false)
			}
		}

		connectEntryRowChanged(&journalEntriesEditPageTitleInput, onValidateJournalEntriesEditPageForm)
		connectTextBufferChanged(journalEntriesEditPageBodyInput.GetBuffer(), onValidateJournalEntriesEditPageForm)

		createErrAndLoadingHandlers := func(
			errorStatusPage *adw.StatusPage,
			errorRefreshButton *gtk.Button,
			errorCopyDetailsButton *gtk.Button,

			onRefresh func(),

			onEnableLoading func(),
			onDisableLoading func(err string),
		) (
			onError func(error),
			enableLoading func(),
			disableLoading func(),
			clearError func(),
		) {
			var rawErr string
			onError = func(err error) {
				rawErr = err.Error()
				i18nErr := L(rawErr)

				log.Error(
					"An unexpected error occured, showing error message to user",
					"rawError", rawErr,
					"i18nErr", i18nErr,
				)

				errorStatusPage.SetDescription(i18nErr)
			}

			connectButtonClicked(errorRefreshButton, onRefresh)

			connectButtonClicked(errorCopyDetailsButton, func() {
				w.GetClipboard().SetText(rawErr)
			})

			enableLoading = onEnableLoading

			disableLoading = func() {
				onDisableLoading(rawErr)
			}

			return onError,
				enableLoading,
				disableLoading,
				func() {
					rawErr = ""
				}
		}

		handleContactsError,
			enableContactsLoading,
			disableContactsLoading,
			clearContactsError := createErrAndLoadingHandlers(
			&contactsErrorStatusPage,
			&contactsErrorRefreshButton,
			&contactsErrorCopyDetailsButton,

			func() {
				homeNavigation.ReplaceWithTags([]string{resources.PageContacts}, 1)
			},

			func() {
				homeSidebarContactsCountLabel.SetVisible(false)
				homeSidebarContactsCountSpinner.SetVisible(true)

				contactsStack.SetVisibleChildName(resources.PageContactsLoading)
			},
			func(err string) {
				homeSidebarContactsCountSpinner.SetVisible(false)
				homeSidebarContactsCountLabel.SetVisible(true)

				homeSidebarContactsCountLabel.SetText(fmt.Sprintf("%v", contactsCount))

				if err == "" {
					if contactsCount > 0 {
						contactsStack.SetVisibleChildName(resources.PageContactsList)
					} else {
						contactsStack.SetVisibleChildName(resources.PageContactsEmpty)
					}
				} else {
					contactsStack.SetVisibleChildName(resources.PageContactsError)
				}
			},
		)

		handleContactsViewError,
			enableContactsViewLoading,
			disableContactsViewLoading,
			clearContactsViewError := createErrAndLoadingHandlers(
			&contactsViewErrorStatusPage,
			&contactsViewErrorRefreshButton,
			&contactsViewErrorCopyDetailsButton,

			func() {
				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView}, 2)
			},

			func() {
				contactsViewStack.SetVisibleChildName(resources.PageContactsViewLoading)
			},
			func(err string) {
				if err == "" {
					contactsViewStack.SetVisibleChildName(resources.PageContactsViewData)
				} else {
					contactsViewStack.SetVisibleChildName(resources.PageContactsViewError)
				}
			},
		)

		handleActivitiesViewError,
			enableActivitiesViewLoading,
			disableActivitiesViewLoading,
			clearActivitiesViewError := createErrAndLoadingHandlers(
			&activitiesViewErrorStatusPage,
			&activitiesViewErrorRefreshButton,
			&activitiesViewErrorCopyDetailsButton,

			func() {
				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView, resources.PageActivitiesView}, 3)
			},

			func() {
				activitiesViewStack.SetVisibleChildName(resources.PageActivitiesViewLoading)
			},
			func(err string) {
				if err == "" {
					activitiesViewStack.SetVisibleChildName(resources.PageActivitiesViewData)
				} else {
					activitiesViewStack.SetVisibleChildName(resources.PageActivitiesViewError)
				}
			},
		)

		handleActivitiesEditError,
			enableActivitiesEditLoading,
			disableActivitiesEditLoading,
			clearActivitiesEditError := createErrAndLoadingHandlers(
			&activitiesEditErrorStatusPage,
			&activitiesEditErrorRefreshButton,
			&activitiesEditErrorCopyDetailsButton,

			func() {
				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView, resources.PageActivitiesView, resources.PageActivitiesEdit}, 4)
			},

			func() {
				activitiesEditStack.SetVisibleChildName(resources.PageActivitiesEditLoading)
			},
			func(err string) {
				if err == "" {
					activitiesEditStack.SetVisibleChildName(resources.PageActivitiesEditData)
					activitiesEditPageNameInput.GrabFocus()
				} else {
					activitiesEditStack.SetVisibleChildName(resources.PageActivitiesEditError)
				}
			},
		)

		handleDebtsEditError,
			enableDebtsEditLoading,
			disableDebtsEditLoading,
			clearDebtsEditError := createErrAndLoadingHandlers(
			&debtsEditErrorStatusPage,
			&debtsEditErrorRefreshButton,
			&debtsEditErrorCopyDetailsButton,

			func() {
				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView, resources.PageDebtsEdit}, 3)
			},

			func() {
				debtsEditStack.SetVisibleChildName(resources.PageDebtsEditLoading)
			},
			func(err string) {
				if err == "" {
					debtsEditStack.SetVisibleChildName(resources.PageDebtsEditData)
					debtsEditPageAmountInput.GrabFocus()
				} else {
					debtsEditStack.SetVisibleChildName(resources.PageDebtsEditError)
				}
			},
		)

		handleContactsEditError,
			enableContactsEditLoading,
			disableContactsEditLoading,
			clearContactsEditError := createErrAndLoadingHandlers(
			&contactsEditErrorStatusPage,
			&contactsEditErrorRefreshButton,
			&contactsEditErrorCopyDetailsButton,

			func() {
				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView, resources.PageContactsEdit}, 3)
			},

			func() {
				contactsEditStack.SetVisibleChildName(resources.PageContactsEditLoading)
			},
			func(err string) {
				if err == "" {
					contactsEditStack.SetVisibleChildName(resources.PageContactsEditData)
					contactsEditPageFirstNameInput.GrabFocus()
				} else {
					contactsEditStack.SetVisibleChildName(resources.PageContactsEditError)
				}
			},
		)

		handleJournalEntriesError,
			enableJournalEntriesLoading,
			disableJournalEntriesLoading,
			clearJournalEntriesError := createErrAndLoadingHandlers(
			&journalEntriesErrorStatusPage,
			&journalEntriesErrorRefreshButton,
			&journalEntriesErrorCopyDetailsButton,

			func() {
				homeNavigation.ReplaceWithTags([]string{resources.PageJournalEntries}, 1)
			},

			func() {
				homeSidebarJournalEntriesCountLabel.SetVisible(false)
				homeSidebarJournalEntriesCountSpinner.SetVisible(true)

				journalEntriesStack.SetVisibleChildName(resources.PageJournalEntriesLoading)
			},
			func(err string) {
				homeSidebarJournalEntriesCountSpinner.SetVisible(false)
				homeSidebarJournalEntriesCountLabel.SetVisible(true)

				homeSidebarJournalEntriesCountLabel.SetText(fmt.Sprintf("%v", journalEntriesCount))

				if err == "" {
					if journalEntriesCount > 0 {
						journalEntriesStack.SetVisibleChildName(resources.PageJournalEntriesList)
					} else {
						journalEntriesStack.SetVisibleChildName(resources.PageJournalEntriesEmpty)
					}
				} else {
					journalEntriesStack.SetVisibleChildName(resources.PageJournalEntriesError)
				}
			},
		)

		handleJournalEntriesViewError,
			enableJournalEntriesViewLoading,
			disableJournalEntriesViewLoading,
			clearJournalEntriesViewError := createErrAndLoadingHandlers(
			&journalEntriesViewErrorStatusPage,
			&journalEntriesViewErrorRefreshButton,
			&journalEntriesViewErrorCopyDetailsButton,

			func() {
				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView, resources.PageJournalEntriesView}, 3)
			},

			func() {
				journalEntriesViewStack.SetVisibleChildName(resources.PageJournalEntriesViewLoading)
			},
			func(err string) {
				if err == "" {
					journalEntriesViewStack.SetVisibleChildName(resources.PageJournalEntriesViewData)
				} else {
					journalEntriesViewStack.SetVisibleChildName(resources.PageJournalEntriesViewError)
				}
			},
		)

		handleJournalEntriesEditError,
			enableJournalEntriesEditLoading,
			disableJournalEntriesEditLoading,
			clearJournalEntriesEditError := createErrAndLoadingHandlers(
			&journalEntriesEditErrorStatusPage,
			&journalEntriesEditErrorRefreshButton,
			&journalEntriesEditErrorCopyDetailsButton,

			func() {
				homeNavigation.ReplaceWithTags([]string{resources.PageJournalEntries, resources.PageJournalEntriesView, resources.PageJournalEntriesEdit}, 3)
			},

			func() {
				journalEntriesEditStack.SetVisibleChildName(resources.PageJournalEntriesEditLoading)
			},
			func(err string) {
				if err == "" {
					journalEntriesEditStack.SetVisibleChildName(resources.PageJournalEntriesEditData)
					journalEntriesEditPageTitleInput.GrabFocus()
				} else {
					journalEntriesEditStack.SetVisibleChildName(resources.PageJournalEntriesEditError)
				}
			},
		)

		connectButtonClicked(&contactsCreateDialogAddButton, func() {
			log.Info("Handling contact creation")

			contactsCreateDialogAddButton.SetSensitive(false)
			contactsCreateDialogAddSpinner.SetVisible(true)

			go func() {
				defer contactsCreateDialogAddSpinner.SetVisible(false)
				defer contactsCreateDialogAddButton.SetSensitive(true)

				redirected, c, _, err := authorize(
					ctx,

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for create contact action", "err", err)

					onPanic(err)

					return
				} else if redirected {
					return
				}

				var nickname *string
				if v := contactsCreateDialogNicknameInput.GetText(); v != "" {
					nickname = &v
				}

				req := api.CreateContactJSONRequestBody{
					Email:     (types.Email)(contactsCreateDialogEmailInput.GetText()),
					FirstName: contactsCreateDialogFirstNameInput.GetText(),
					LastName:  contactsCreateDialogLastNameInput.GetText(),
					Nickname:  nickname,
					Pronouns:  contactsCreateDialogPronounsInput.GetText(),
				}

				log.Debug("Creating contact", "request", req)

				res, err := c.CreateContactWithResponse(ctx, req)
				if err != nil {
					onPanic(err)

					return
				}

				log.Debug("Created contact", "status", res.StatusCode())

				if res.StatusCode() != http.StatusOK {
					onPanic(errors.New(res.Status()))

					return
				}

				mto.AddToast(adw.NewToast(L("Created contact")))

				contactsCreateDialog.Close()

				homeNavigation.ReplaceWithTags([]string{resources.PageContacts}, 1)
			}()
		})

		connectButtonClicked(&debtsCreateDialogAddButton, func() {
			id := debtsCreateDialogAddButton.GetActionTargetValue().GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling debt creation")

			debtsCreateDialogAddButton.SetSensitive(false)
			debtsCreateDialogAddSpinner.SetVisible(true)

			go func() {
				defer debtsCreateDialogAddSpinner.SetVisible(false)
				defer debtsCreateDialogAddButton.SetSensitive(true)

				redirected, c, _, err := authorize(
					ctx,

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for create debt action", "err", err)

					onPanic(err)

					return
				} else if redirected {
					return
				}

				var description *string
				if v := getTextBufferText(debtsCreateDialogDescriptionInput.GetBuffer()); v != "" {
					description = &v
				}

				req := api.CreateDebtJSONRequestBody{
					Amount:      float32(debtsCreateDialogAmountInput.GetValue()),
					ContactId:   id,
					Currency:    debtsCreateDialogCurrencyInput.GetText(),
					Description: description,
					YouOwe:      debtsCreateDialogYouOweRadio.GetActive(),
				}

				log.Debug("Creating debt", "request", req)

				res, err := c.CreateDebtWithResponse(ctx, req)
				if err != nil {
					onPanic(err)

					return
				}

				log.Debug("Created debt", "status", res.StatusCode())

				if res.StatusCode() != http.StatusOK {
					onPanic(errors.New(res.Status()))

					return
				}

				mto.AddToast(adw.NewToast(L("Created debt")))

				debtsCreateDialog.Close()

				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView}, 2)
			}()
		})

		connectButtonClicked(&activitiesCreateDialogAddButton, func() {
			id := activitiesCreateDialogAddButton.GetActionTargetValue().GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling activity creation")

			activitiesCreateDialogAddButton.SetSensitive(false)
			activitiesCreateDialogAddSpinner.SetVisible(true)

			go func() {
				defer activitiesCreateDialogAddSpinner.SetVisible(false)
				defer activitiesCreateDialogAddButton.SetSensitive(true)

				redirected, c, _, err := authorize(
					ctx,

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for create activity action", "err", err)

					onPanic(err)

					return
				} else if redirected {
					return
				}

				var description *string
				if v := getTextBufferText(activitiesCreateDialogDescriptionInput.GetBuffer()); v != "" {
					description = &v
				}

				localeDate, err := parseLocaleDate(activitiesCreateDialogDateInput.GetText())
				if err != nil {
					onPanic(err)

					return
				}

				req := api.CreateActivityJSONRequestBody{
					ContactId: id,
					Date: types.Date{
						Time: localeDate,
					},
					Description: description,
					Name:        activitiesCreateDialogNameInput.GetText(),
				}

				log.Debug("Creating activity", "request", req)

				res, err := c.CreateActivityWithResponse(ctx, req)
				if err != nil {
					onPanic(err)

					return
				}

				log.Debug("Created activity", "status", res.StatusCode())

				if res.StatusCode() != http.StatusOK {
					onPanic(errors.New(res.Status()))

					return
				}

				mto.AddToast(adw.NewToast(L("Created activity")))

				activitiesCreateDialog.Close()

				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView}, 2)
			}()
		})

		connectButtonClicked(&activitiesEditPageSaveButton, func() {
			id := activitiesEditPageSaveButton.GetActionTargetValue().GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling activity update")

			activitiesEditPageSaveButton.SetSensitive(false)
			activitiesEditPageSaveSpinner.SetVisible(true)

			go func() {
				defer activitiesEditPageSaveSpinner.SetVisible(false)
				defer activitiesEditPageSaveButton.SetSensitive(true)

				redirected, c, _, err := authorize(
					ctx,

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for update activity action", "err", err)

					onPanic(err)

					return
				} else if redirected {
					return
				}

				var description *string
				if v := getTextBufferText(activitiesEditPageDescriptionInput.GetBuffer()); v != "" {
					description = &v
				}

				localeDate, err := parseLocaleDate(activitiesEditPageDateInput.GetText())
				if err != nil {
					onPanic(err)

					return
				}

				req := api.UpdateActivityJSONRequestBody{
					Date: types.Date{
						Time: localeDate,
					},
					Description: description,
					Name:        activitiesEditPageNameInput.GetText(),
				}

				log.Debug("Updating activity", "request", req)

				res, err := c.UpdateActivityWithResponse(ctx, id, req)
				if err != nil {
					onPanic(err)

					return
				}

				log.Debug("Updated activity", "status", res.StatusCode())

				if res.StatusCode() != http.StatusOK {
					onPanic(errors.New(res.Status()))

					return
				}

				mto.AddToast(adw.NewToast(L("Updated activity")))

				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView, resources.PageActivitiesView}, 3)
			}()
		})

		connectButtonClicked(&debtsEditPageSaveButton, func() {
			id := debtsEditPageSaveButton.GetActionTargetValue().GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling debt update")

			debtsEditPageSaveButton.SetSensitive(false)
			debtsEditPageSaveSpinner.SetVisible(true)

			go func() {
				defer debtsEditPageSaveSpinner.SetVisible(false)
				defer debtsEditPageSaveButton.SetSensitive(true)

				redirected, c, _, err := authorize(
					ctx,

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for update debt action", "err", err)

					onPanic(err)

					return
				} else if redirected {
					return
				}

				var description *string
				if v := getTextBufferText(debtsEditPageDescriptionInput.GetBuffer()); v != "" {
					description = &v
				}

				req := api.UpdateDebtJSONRequestBody{
					Amount:      float32(debtsEditPageAmountInput.GetValue()),
					Currency:    debtsEditPageCurrencyInput.GetText(),
					Description: description,
					YouOwe:      debtsEditPageYouOweRadio.GetActive(),
				}

				log.Debug("Updating debt", "request", req)

				res, err := c.UpdateDebtWithResponse(ctx, id, req)
				if err != nil {
					onPanic(err)

					return
				}

				log.Debug("Updated debt", "status", res.StatusCode())

				if res.StatusCode() != http.StatusOK {
					onPanic(errors.New(res.Status()))

					return
				}

				mto.AddToast(adw.NewToast(L("Updated debt")))

				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView}, 2)
			}()
		})

		connectButtonClicked(&contactsEditPageSaveButton, func() {
			id := contactsEditPageSaveButton.GetActionTargetValue().GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling contact update")

			contactsEditPageSaveButton.SetSensitive(false)
			contactsEditPageSaveSpinner.SetVisible(true)

			go func() {
				defer contactsEditPageSaveSpinner.SetVisible(false)
				defer contactsEditPageSaveButton.SetSensitive(true)

				redirected, c, _, err := authorize(
					ctx,

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for update contact action", "err", err)

					onPanic(err)

					return
				} else if redirected {
					return
				}

				var nickname *string
				if v := contactsEditPageNicknameInput.GetText(); v != "" {
					nickname = &v
				}

				var birthday *types.Date
				if v := contactsEditPageBirthdayInput.GetText(); v != "" {
					localeBirthday, err := parseLocaleDate(contactsEditPageBirthdayInput.GetText())
					if err != nil {
						onPanic(err)

						return
					}

					birthday = &types.Date{
						Time: localeBirthday,
					}
				}

				var address *string
				if v := getTextBufferText(contactsEditPageAddressInput.GetBuffer()); v != "" {
					address = &v
				}

				var notes *string
				if v := getTextBufferText(contactsEditPageNotesInput.GetBuffer()); v != "" {
					notes = &v
				}

				req := api.UpdateContactJSONRequestBody{
					Email:     (types.Email)(contactsEditPageEmailInput.GetText()),
					FirstName: contactsEditPageFirstNameInput.GetText(),
					LastName:  contactsEditPageLastNameInput.GetText(),
					Nickname:  nickname,
					Pronouns:  contactsEditPagePronounsInput.GetText(),

					Birthday: birthday,
					Address:  address,
					Notes:    notes,
				}

				log.Debug("Creating contact", "request", req)

				res, err := c.UpdateContactWithResponse(ctx, id, req)
				if err != nil {
					onPanic(err)

					return
				}

				log.Debug("Updated contact", "status", res.StatusCode())

				if res.StatusCode() != http.StatusOK {
					onPanic(errors.New(res.Status()))

					return
				}

				mto.AddToast(adw.NewToast(L("Updated contact")))

				homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView}, 2)
			}()
		})

		connectButtonClicked(&journalEntriesCreateDialogAddButton, func() {
			log.Info("Handling journal entry creation")

			journalEntriesCreateDialogAddButton.SetSensitive(false)
			journalEntriesCreateDialogAddSpinner.SetVisible(true)

			go func() {
				defer journalEntriesCreateDialogAddSpinner.SetVisible(false)
				defer journalEntriesCreateDialogAddButton.SetSensitive(true)

				redirected, c, _, err := authorize(
					ctx,

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for create journal entry action", "err", err)

					onPanic(err)

					return
				} else if redirected {
					return
				}

				req := api.CreateJournalEntryJSONRequestBody{
					Body:   getTextBufferText(journalEntriesCreateDialogBodyInput.GetBuffer()),
					Rating: int32(3 - journalEntriesCreateDialogRatingToggleGroup.GetActive()), // The toggle group is zero-indexed, but the rating is one-indexed
					Title:  journalEntriesCreateDialogTitleInput.GetText(),
				}

				log.Debug("Creating journal entry", "request", req)

				res, err := c.CreateJournalEntryWithResponse(ctx, req)
				if err != nil {
					onPanic(err)

					return
				}

				log.Debug("Created journal entry", "status", res.StatusCode())

				if res.StatusCode() != http.StatusOK {
					onPanic(errors.New(res.Status()))

					return
				}

				mto.AddToast(adw.NewToast(L("Created journal entry")))

				journalEntriesCreateDialog.Close()

				homeNavigation.ReplaceWithTags([]string{resources.PageJournalEntries}, 1)
			}()
		})

		connectButtonClicked(&journalEntriesEditPageSaveButton, func() {
			id := journalEntriesEditPageSaveButton.GetActionTargetValue().GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling journal entry update")

			journalEntriesEditPageSaveButton.SetSensitive(false)
			journalEntriesEditPageSaveSpinner.SetVisible(true)

			go func() {
				defer journalEntriesEditPageSaveSpinner.SetVisible(false)
				defer journalEntriesEditPageSaveButton.SetSensitive(true)

				redirected, c, _, err := authorize(
					ctx,

					false,
				)
				if err != nil {
					log.Warn("Could not authorize user for update journal entry action", "err", err)

					onPanic(err)

					return
				} else if redirected {
					return
				}

				req := api.UpdateJournalEntryJSONRequestBody{
					Body:   getTextBufferText(journalEntriesEditPageBodyInput.GetBuffer()),
					Rating: int32((3 - journalEntriesEditPageRatingToggleGroup.GetActive())), // The toggle group is zero-indexed, but the rating is one-indexed
					Title:  journalEntriesEditPageTitleInput.GetText(),
				}

				log.Debug("Creating journal entry", "request", req)

				res, err := c.UpdateJournalEntryWithResponse(ctx, id, req)
				if err != nil {
					onPanic(err)

					return
				}

				log.Debug("Updated journal entry", "status", res.StatusCode())

				if res.StatusCode() != http.StatusOK {
					onPanic(errors.New(res.Status()))

					return
				}

				mto.AddToast(adw.NewToast(L("Updated journal entry")))

				homeNavigation.ReplaceWithTags([]string{resources.PageJournalEntries, resources.PageJournalEntriesView}, 2)
			}()
		})

		logoutAction := gio.NewSimpleAction("logout", nil)
		connectSimpleActionActivate(logoutAction, func() {
			nv.ReplaceWithTags([]string{resources.PageExchangeLogout}, 1)

			go func() {
				if _, err := gio.AppInfoLaunchDefaultForUri(u.LogoutURL, nil); err != nil {
					onPanic(err)

					return
				}
			}()
		})
		a.AddAction(logoutAction)

		licenseAction := gio.NewSimpleAction("license", nil)
		connectSimpleActionActivate(licenseAction, func() {
			log.Info("Handling getting license action", "url", spec.Info.License.URL)

			go func() {
				if _, err := gio.AppInfoLaunchDefaultForUri(spec.Info.License.URL, nil); err != nil {
					onPanic(err)

					return
				}
			}()
		})
		a.AddAction(licenseAction)

		privacyAction := gio.NewSimpleAction("privacy", nil)
		connectSimpleActionActivate(privacyAction, func() {
			var privacyURL string
			if v := spec.Info.Extensions[api.PrivacyPolicyExtensionKey]; v != nil {
				vv, ok := v.(string)
				if ok {
					privacyURL = vv
				} else {
					onPanic(errMissingPrivacyURL)

					return
				}
			}

			log.Info("Handling getting privacy action", "url", privacyURL)

			go func() {
				if _, err := gio.AppInfoLaunchDefaultForUri(privacyURL, nil); err != nil {
					onPanic(err)

					return
				}
			}()
		})
		a.AddAction(privacyAction)

		tosAction := gio.NewSimpleAction("tos", nil)
		connectSimpleActionActivate(tosAction, func() {
			log.Info("Handling getting terms of service action", "url", spec.Info.TermsOfService)

			go func() {
				if _, err := gio.AppInfoLaunchDefaultForUri(spec.Info.TermsOfService, nil); err != nil {
					onPanic(err)

					return
				}
			}()
		})
		a.AddAction(tosAction)

		imprintAction := gio.NewSimpleAction("imprint", nil)
		connectSimpleActionActivate(imprintAction, func() {
			log.Info("Handling getting imprint action", "url", spec.Info.Contact.URL)

			go func() {
				if _, err := gio.AppInfoLaunchDefaultForUri(spec.Info.Contact.URL, nil); err != nil {
					onPanic(err)

					return
				}
			}()
		})
		a.AddAction(imprintAction)

		codeAction := gio.NewSimpleAction("code", nil)
		connectSimpleActionActivate(codeAction, func() {
			log.Info("Handling getting code action")

			enableHomeHamburgerMenuLoading()

			redirected, c, _, err := authorize(
				ctx,

				false,
			)
			if err != nil {
				disableHomeHamburgerMenuLoading()

				log.Warn("Could not authorize user for getting code action", "err", err)

				onPanic(err)

				return
			} else if redirected {
				disableHomeHamburgerMenuLoading()

				return
			}

			log.Debug("Getting code")

			res, err := c.GetSourceCode(ctx)
			if err != nil {
				disableHomeHamburgerMenuLoading()

				onPanic(err)

				return
			}

			log.Debug("Received code", "status", res.StatusCode)

			if res.StatusCode != http.StatusOK {
				_ = res.Body.Close()

				disableHomeHamburgerMenuLoading()

				onPanic(errors.New(res.Status))

				return
			}

			log.Debug("Writing code to file")

			fd := gtk.NewFileDialog()
			fd.SetTitle(L("Senbara REST source code"))
			fd.SetInitialName("code.tar.gz")
			fileDialogSave(fd, &w.Window, func(file *gio.FileBase, err error) {
				go func() {
					defer disableHomeHamburgerMenuLoading()
					defer res.Body.Close()

					if err != nil {
						onPanic(err)

						return
					}

					log.Debug("Writing code to file", "path", file.GetPath())

					f, err := os.OpenFile(file.GetPath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
					if err != nil {
						onPanic(err)

						return
					}
					defer f.Close()

					if _, err := io.Copy(f, res.Body); err != nil {
						onPanic(err)

						return
					}

					log.Debug("Downloaded code", "status", res.StatusCode)

					mto.AddToast(adw.NewToast(L("Downloaded code")))
				}()
			})
		})
		a.AddAction(codeAction)

		exportUserDataAction := gio.NewSimpleAction("exportUserData", nil)
		connectSimpleActionActivate(exportUserDataAction, func() {
			log.Info("Handling export user data action")

			enableHomeUserMenuLoading()

			redirected, c, _, err := authorize(
				ctx,

				false,
			)
			if err != nil {
				disableHomeUserMenuLoading()

				log.Warn("Could not authorize user for export user data action", "err", err)

				onPanic(err)

				return
			} else if redirected {
				disableHomeUserMenuLoading()

				return
			}

			log.Debug("Exporting user data")

			res, err := c.ExportUserData(ctx)
			if err != nil {
				disableHomeUserMenuLoading()

				onPanic(err)

				return
			}

			log.Debug("Exported user data", "status", res.StatusCode)

			if res.StatusCode != http.StatusOK {
				_ = res.Body.Close()

				disableHomeUserMenuLoading()

				onPanic(errors.New(res.Status))

				return
			}

			log.Debug("Writing user data to file")

			fd := gtk.NewFileDialog()
			fd.SetTitle(L("Senbara Forms userdata"))
			fd.SetInitialName("userdata.jsonl")
			fileDialogSave(fd, &w.Window, func(file *gio.FileBase, err error) {
				go func() {
					defer disableHomeUserMenuLoading()
					defer res.Body.Close()

					if err != nil {
						onPanic(err)

						return
					}

					log.Debug("Writing user data to file", "path", file.GetPath())

					f, err := os.OpenFile(file.GetPath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
					if err != nil {
						onPanic(err)

						return
					}
					defer f.Close()

					if _, err := io.Copy(f, res.Body); err != nil {
						onPanic(err)

						return
					}

					log.Debug("Exported user data", "status", res.StatusCode)

					mto.AddToast(adw.NewToast(L("Exported user data")))
				}()
			})
		})
		a.AddAction(exportUserDataAction)

		refreshSidebarWithLatestSummary := func() bool {
			enableHomeSidebarLoading()
			defer disableHomeSidebarLoading()

			redirected, c, _, err := authorize(
				ctx,

				true,
			)
			if err != nil {
				log.Warn("Could not authorize user for home page", "err", err)

				onPanic(err)

				return false
			} else if redirected {
				return false
			}

			settings.SetBoolean(resources.SettingAnonymousMode, false)

			log.Debug("Getting summary")

			res, err := c.GetSummaryWithResponse(ctx)
			if err != nil {
				onPanic(err)

				return false
			}

			log.Debug("Got summary", "status", res.StatusCode())

			if res.StatusCode() != http.StatusOK {
				onPanic(errors.New(res.Status()))

				return false
			}

			homeSidebarContactsCountLabel.SetText(fmt.Sprintf("%v", *res.JSON200.ContactsCount))
			homeSidebarJournalEntriesCountLabel.SetText(fmt.Sprintf("%v", *res.JSON200.JournalEntriesCount))

			return true
		}

		importUserDataAction := gio.NewSimpleAction("importUserData", nil)
		connectSimpleActionActivate(importUserDataAction, func() {
			log.Info("Handling import user data action")

			fd := gtk.NewFileDialog()
			fd.SetTitle(L("Senbara Forms userdata"))

			ls := gio.NewListStore(gobject.ObjectGLibType())

			{
				fi := gtk.NewFileFilter()
				fi.SetName(L("Senbara Forms userdata files"))
				fi.AddPattern("*.jsonl")
				ls.Append(&fi.Filter.Object)
			}

			{
				fi := gtk.NewFileFilter()
				fi.SetName(L("All files"))
				fi.AddPattern("*")
				ls.Append(&fi.Filter.Object)
			}

			fd.SetFilters(ls)

			fileDialogOpen(fd, &w.Window, func(file *gio.FileBase, err error) {
				if err != nil {
					onPanic(err)

					return
				}

				confirm := adw.NewAlertDialog(
					L("Importing user data"),
					L("Are you sure you want to import this user data into your account?"),
				)
				confirm.AddResponse("cancel", L("Cancel"))
				confirm.AddResponse("import", L("Import"))
				confirm.SetResponseAppearance("import", adw.ResponseSuggestedValue)
				connectAlertDialogResponse(confirm, func(response string) {
					if response == "import" {
						go func() {
							enableHomeUserMenuLoading()
							defer disableHomeUserMenuLoading()

							redirected, c, _, err := authorize(
								ctx,

								false,
							)
							if err != nil {
								disableHomeUserMenuLoading()

								log.Warn("Could not authorize user for import user data action", "err", err)

								onPanic(err)

								return
							} else if redirected {
								disableHomeUserMenuLoading()

								return
							}

							log.Debug("Reading user data from file", "path", file.GetPath())

							f, err := os.OpenFile(file.GetPath(), os.O_RDONLY, os.ModePerm)
							if err != nil {
								onPanic(err)

								return
							}
							defer f.Close()

							log.Debug("Importing user data, reading from file and streaming to API")

							reader, writer := io.Pipe()
							enc := multipart.NewWriter(writer)
							go func() {
								defer writer.Close()

								if err := func() error {
									formFile, err := enc.CreateFormFile("userData", "")
									if err != nil {
										return err
									}

									if _, err := io.Copy(formFile, f); err != nil {
										return err
									}

									if err := enc.Close(); err != nil {
										return err
									}

									return nil
								}(); err != nil {
									log.Warn("Could not stream user data to API", "err", err)

									writer.CloseWithError(err)

									return
								}
							}()

							res, err := c.ImportUserDataWithBodyWithResponse(ctx, enc.FormDataContentType(), reader)
							if err != nil {
								onPanic(err)

								return
							}

							log.Debug("Imported user data", "status", res.StatusCode())

							if res.StatusCode() != http.StatusOK {
								onPanic(errors.New(res.Status()))

								return
							}

							mto.AddToast(adw.NewToast(L("Imported user data")))

							go func() {
								_ = refreshSidebarWithLatestSummary()
							}()
						}()
					}
				})

				confirm.Present(&w.ApplicationWindow.Window.Widget)
			})
		})
		a.AddAction(importUserDataAction)

		deleteUserDataAction := gio.NewSimpleAction("deleteUserData", nil)
		connectSimpleActionActivate(deleteUserDataAction, func() {
			log.Info("Handling delete user data action")

			confirm := adw.NewAlertDialog(
				L("Deleting your data"),
				L("Are you sure you want to delete your data and your account?"),
			)
			confirm.AddResponse("cancel", L("Cancel"))
			confirm.AddResponse("delete", L("Delete"))
			confirm.SetResponseAppearance("delete", adw.ResponseDestructiveValue)
			connectAlertDialogResponse(confirm, func(response string) {
				if response == "delete" {
					redirected, c, _, err := authorize(
						ctx,

						false,
					)
					if err != nil {
						log.Warn("Could not authorize user for delete user data action", "err", err)

						onPanic(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Deleting user data")

					res, err := c.DeleteUserDataWithResponse(ctx)
					if err != nil {
						onPanic(err)

						return
					}

					log.Debug("Deleted user data", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						onPanic(errors.New(res.Status()))

						return
					}

					mto.AddToast(adw.NewToast(L("Deleted user data")))

					logoutAction.Activate(nil)
				}
			})

			confirm.Present(&w.ApplicationWindow.Window.Widget)
		})
		a.AddAction(deleteUserDataAction)

		aboutAction := gio.NewSimpleAction("about", nil)
		connectSimpleActionActivate(aboutAction, func() {
			aboutDialog.Present(&w.ApplicationWindow.Window.Widget)
		})
		a.AddAction(aboutAction)

		quitAction := gio.NewSimpleAction("quit", nil)
		connectSimpleActionActivate(quitAction, func() {
			a.Quit()
		})
		a.SetAccelsForAction("app.quit", []string{`<Primary>q`})
		a.AddAction(quitAction)

		copyErrorToClipboardAction := gio.NewSimpleAction("copyErrorToClipboard", nil)
		connectSimpleActionActivate(copyErrorToClipboardAction, func() {
			w.GetClipboard().SetText(rawError)
		})
		a.AddAction(copyErrorToClipboardAction)

		deleteContactAction := gio.NewSimpleAction("deleteContact", glib.NewVariantType("x"))
		connectSimpleActionActivateWithParam(deleteContactAction, func(parameter *glib.Variant) {
			id := parameter.GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling delete contact action")

			confirm := adw.NewAlertDialog(
				L("Deleting a contact"),
				L("Are you sure you want to delete this contact?"),
			)
			confirm.AddResponse("cancel", L("Cancel"))
			confirm.AddResponse("delete", L("Delete"))
			confirm.SetResponseAppearance("delete", adw.ResponseDestructiveValue)
			connectAlertDialogResponse(confirm, func(response string) {
				if response == "delete" {
					redirected, c, _, err := authorize(
						ctx,

						false,
					)
					if err != nil {
						log.Warn("Could not authorize user for delete contact action", "err", err)

						onPanic(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Deleting contact")

					res, err := c.DeleteContactWithResponse(ctx, int64(id))
					if err != nil {
						onPanic(err)

						return
					}

					log.Debug("Deleted contact", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						onPanic(errors.New(res.Status()))

						return
					}

					mto.AddToast(adw.NewToast(L("Contact deleted")))

					homeNavigation.ReplaceWithTags([]string{resources.PageContacts}, 1)
				}
			})

			confirm.Present(&w.ApplicationWindow.Window.Widget)
		})
		a.AddAction(deleteContactAction)

		settleDebtAction := gio.NewSimpleAction("settleDebt", glib.NewVariantType("x"))
		connectSimpleActionActivateWithParam(settleDebtAction, func(parameter *glib.Variant) {
			id := parameter.GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling settle debt action")

			confirm := adw.NewAlertDialog(
				L("Settling a debt"),
				L("Are you sure you want to settle this debt?"),
			)
			confirm.AddResponse("cancel", L("Cancel"))
			confirm.AddResponse("delete", L("Delete"))
			confirm.SetResponseAppearance("delete", adw.ResponseDestructiveValue)
			connectAlertDialogResponse(confirm, func(response string) {
				if response == "delete" {
					redirected, c, _, err := authorize(
						ctx,

						false,
					)
					if err != nil {
						log.Warn("Could not authorize user for settle debt action", "err", err)

						onPanic(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Settling debt")

					res, err := c.SettleDebtWithResponse(ctx, int64(id))
					if err != nil {
						onPanic(err)

						return
					}

					log.Debug("Settled debt", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						onPanic(errors.New(res.Status()))

						return
					}

					mto.AddToast(adw.NewToast(L("Settled debt")))

					homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView}, 2)
				}
			})

			confirm.Present(&w.ApplicationWindow.Window.Widget)
		})
		a.AddAction(settleDebtAction)

		deleteActivityAction := gio.NewSimpleAction("deleteActivity", glib.NewVariantType("x"))
		connectSimpleActionActivateWithParam(deleteActivityAction, func(parameter *glib.Variant) {
			id := parameter.GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling delete activity action")

			confirm := adw.NewAlertDialog(
				L("Deleting an activity"),
				L("Are you sure you want to delete this activity?"),
			)
			confirm.AddResponse("cancel", L("Cancel"))
			confirm.AddResponse("delete", L("Delete"))
			confirm.SetResponseAppearance("delete", adw.ResponseDestructiveValue)
			connectAlertDialogResponse(confirm, func(response string) {
				if response == "delete" {
					redirected, c, _, err := authorize(
						ctx,

						false,
					)
					if err != nil {
						log.Warn("Could not authorize user for delete activity action", "err", err)

						onPanic(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Deleting activity")

					res, err := c.DeleteActivityWithResponse(ctx, int64(id))
					if err != nil {
						onPanic(err)

						return
					}

					log.Debug("Deleted activity", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						onPanic(errors.New(res.Status()))

						return
					}

					mto.AddToast(adw.NewToast(L("Activity deleted")))

					homeNavigation.ReplaceWithTags([]string{resources.PageContacts, resources.PageContactsView}, 2)
				}
			})

			confirm.Present(&w.ApplicationWindow.Window.Widget)
		})
		a.AddAction(deleteActivityAction)

		deleteJournalEntryAction := gio.NewSimpleAction("deleteJournalEntry", glib.NewVariantType("x"))
		connectSimpleActionActivateWithParam(deleteJournalEntryAction, func(parameter *glib.Variant) {
			id := parameter.GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling delete journal entry action")

			confirm := adw.NewAlertDialog(
				L("Deleting a journal entry"),
				L("Are you sure you want to delete this journal entry?"),
			)
			confirm.AddResponse("cancel", L("Cancel"))
			confirm.AddResponse("delete", L("Delete"))
			confirm.SetResponseAppearance("delete", adw.ResponseDestructiveValue)
			connectAlertDialogResponse(confirm, func(response string) {
				if response == "delete" {
					redirected, c, _, err := authorize(
						ctx,

						false,
					)
					if err != nil {
						log.Warn("Could not authorize user for delete journal entry action", "err", err)

						onPanic(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Deleting journal entry")

					res, err := c.DeleteJournalEntryWithResponse(ctx, int64(id))
					if err != nil {
						onPanic(err)

						return
					}

					log.Debug("Deleted journal entry", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						onPanic(errors.New(res.Status()))

						return
					}

					mto.AddToast(adw.NewToast(L("Journal entry deleted")))

					homeNavigation.ReplaceWithTags([]string{resources.PageJournalEntries}, 1)
				}
			})

			confirm.Present(&w.ApplicationWindow.Window.Widget)
		})
		a.AddAction(deleteJournalEntryAction)

		var selectedActivityID = -1

		editActivityAction := gio.NewSimpleAction("editActivity", glib.NewVariantType("x"))
		connectSimpleActionActivateWithParam(editActivityAction, func(parameter *glib.Variant) {
			id := parameter.GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling edit activity action")

			selectedActivityID = int(id)

			homeNavigation.PushByTag(resources.PageActivitiesEdit)
		})
		a.AddAction(editActivityAction)

		var selectedDebtID = -1

		editDebtAction := gio.NewSimpleAction("editDebt", glib.NewVariantType("x"))
		connectSimpleActionActivateWithParam(editDebtAction, func(parameter *glib.Variant) {
			id := parameter.GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling edit debt action")

			selectedDebtID = int(id)

			homeNavigation.PushByTag(resources.PageDebtsEdit)
		})
		a.AddAction(editDebtAction)

		var selectedContactID = -1

		editContactAction := gio.NewSimpleAction("editContact", glib.NewVariantType("x"))
		connectSimpleActionActivateWithParam(editContactAction, func(parameter *glib.Variant) {
			id := parameter.GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling edit contact action")

			selectedContactID = int(id)

			homeNavigation.PushByTag(resources.PageContactsEdit)
		})
		a.AddAction(editContactAction)

		var selectedJournalEntryID = -1

		editJournalEntryAction := gio.NewSimpleAction("editJournalEntry", glib.NewVariantType("x"))
		connectSimpleActionActivateWithParam(editJournalEntryAction, func(parameter *glib.Variant) {
			id := parameter.GetInt64()

			log := log.With(
				"id", id,
			)

			log.Info("Handling edit journal entry action")

			selectedJournalEntryID = int(id)

			homeNavigation.PushByTag(resources.PageJournalEntriesEdit)
		})
		a.AddAction(editJournalEntryAction)

		createItemAction := gio.NewSimpleAction("createItem", nil)
		connectSimpleActionActivate(createItemAction, func() {
			var (
				tag = homeNavigation.GetVisiblePage().GetTag()
				log = log.With("tag", tag)
			)

			log.Info("Handling create item action")

			switch tag {
			case resources.PageContacts:
				contactsCreateDialog.Present(&w.ApplicationWindow.Window.Widget)
				contactsCreateDialogFirstNameInput.GrabFocus()

			case resources.PageJournalEntries:
				journalEntriesCreateDialog.Present(&w.ApplicationWindow.Window.Widget)
				journalEntriesCreateDialogTitleInput.GrabFocus()
			}
		})
		a.AddAction(createItemAction)
		a.SetAccelsForAction("app.createItem", []string{`<Primary>n`})

		searchListAction := gio.NewSimpleAction("searchList", nil)
		connectSimpleActionActivate(searchListAction, func() {
			var (
				tag = homeNavigation.GetVisiblePage().GetTag()
				log = log.With("tag", tag)
			)

			log.Info("Handling search list action")

			switch tag {
			case resources.PageContacts:
				contactsSearchButton.SetActive(!contactsSearchButton.GetActive())

			case resources.PageJournalEntries:
				journalEntriesSearchButton.SetActive(!journalEntriesSearchButton.GetActive())
			}
		})
		a.AddAction(searchListAction)
		a.SetAccelsForAction("app.searchList", []string{`<Primary>f`})

		editItemAction := gio.NewSimpleAction("editItem", nil)
		connectSimpleActionActivate(editItemAction, func() {
			var (
				tag = homeNavigation.GetVisiblePage().GetTag()
				log = log.With("tag", tag)
			)

			log.Info("Handling edit item action")

			switch tag {
			case resources.PageContactsView:
				if selectedContactID != -1 {
					editContactAction.Activate(glib.NewVariantInt64(int64(selectedContactID)))
				}

			case resources.PageActivitiesView:
				if selectedActivityID != -1 {
					editActivityAction.Activate(glib.NewVariantInt64(int64(selectedActivityID)))
				}

			case resources.PageJournalEntriesView:
				if selectedJournalEntryID != -1 {
					editJournalEntryAction.Activate(glib.NewVariantInt64(int64(selectedJournalEntryID)))
				}
			}
		})
		a.AddAction(editItemAction)
		a.SetAccelsForAction("app.editItem", []string{`<Primary>e`})

		deleteItemAction := gio.NewSimpleAction("deleteItem", nil)
		connectSimpleActionActivate(deleteItemAction, func() {
			var (
				tag = homeNavigation.GetVisiblePage().GetTag()
				log = log.With("tag", tag)
			)

			log.Info("Handling delete item action")

			switch tag {
			case resources.PageContactsView:
				if selectedContactID != -1 {
					deleteContactAction.Activate(glib.NewVariantInt64(int64(selectedContactID)))
				}

			case resources.PageActivitiesView:
				if selectedActivityID != -1 {
					deleteActivityAction.Activate(glib.NewVariantInt64(int64(selectedActivityID)))
				}

			case resources.PageJournalEntriesView:
				if selectedJournalEntryID != -1 {
					deleteJournalEntryAction.Activate(glib.NewVariantInt64(int64(selectedJournalEntryID)))
				}
			}
		})
		a.AddAction(deleteItemAction)
		a.SetAccelsForAction("app.deleteItem", []string{`<Primary>Delete`})

		navigateToContactsAction := gio.NewSimpleAction("navigateToContacts", nil)
		connectSimpleActionActivate(navigateToContactsAction, func() {
			log.Info("Handling navigate to contacts action")

			contactsRow := homeSidebarListbox.GetRowAtIndex(0)
			contactsRow.GrabFocus()
			homeSidebarListbox.SelectRow(contactsRow)

			homeNavigation.ReplaceWithTags([]string{resources.PageContacts}, 1)
		})
		a.AddAction(navigateToContactsAction)
		a.SetAccelsForAction("app.navigateToContacts", []string{`<Alt>1`})

		navigateToJournalAction := gio.NewSimpleAction("navigateToJournal", nil)
		connectSimpleActionActivate(navigateToJournalAction, func() {
			log.Info("Handling navigate to journal action")

			journalEntriesRow := homeSidebarListbox.GetRowAtIndex(1)
			journalEntriesRow.GrabFocus()
			homeSidebarListbox.SelectRow(journalEntriesRow)

			homeNavigation.ReplaceWithTags([]string{resources.PageJournalEntries}, 1)
		})
		a.AddAction(navigateToJournalAction)
		a.SetAccelsForAction("app.navigateToJournal", []string{`<Alt>2`})

		md := goldmark.New(
			goldmark.WithExtensions(extension.GFM),
		)

		onHomeNavigation := func() {
			var (
				tag = homeNavigation.GetVisiblePage().GetTag()
				log = log.With("tag", tag)
			)

			log.Info("Handling page")

			homeContentPage.SetTitle(homeNavigation.GetVisiblePage().GetTitle())

			homeSplitView.SetShowContent(true)

			switch tag {
			case resources.PageContacts:
				go func() {
					enableContactsLoading()
					defer disableContactsLoading()

					redirected, c, _, err := authorize(
						ctx,

						true,
					)
					if err != nil {
						log.Warn("Could not authorize user for contacts page", "err", err)

						handleContactsError(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Listing contacts")

					res, err := c.GetContactsWithResponse(ctx)
					if err != nil {
						handleContactsError(err)

						return
					}

					log.Debug("Got contacts", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handleContactsError(errors.New(res.Status()))

						return
					}

					defer clearContactsError()

					onValidateContactsCreateDialogForm()

					contactsListBox.RemoveAll()

					contactsCount = len(*res.JSON200)
					if contactsCount > 0 {
						contactsAddButton.SetVisible(true)
						contactsSearchButton.SetVisible(true)

						for _, contact := range *res.JSON200 {
							r := adw.NewActionRow()

							title := *contact.FirstName + " " + *contact.LastName
							if *contact.Nickname != "" {
								title += " (" + *contact.Nickname + ")"
							}
							r.SetTitle(title)

							subtitle := ""
							if *contact.Email != "" {
								subtitle = string(*contact.Email)
							}
							if string(*contact.Email) != "" && string(*contact.Pronouns) != "" {
								subtitle += " | " + *contact.Pronouns
							} else if string(*contact.Pronouns) != "" {
								subtitle = *contact.Pronouns
							}
							if subtitle != "" {
								r.SetSubtitle(subtitle)
							}

							r.SetName("/contacts/view?id=" + strconv.Itoa(int(*contact.Id)))

							menuButton := gtk.NewMenuButton()
							menuButton.SetValign(gtk.AlignCenterValue)
							menuButton.SetIconName("view-more-symbolic")
							menuButton.AddCssClass("flat")

							menu := gio.NewMenu()

							deleteContactMenuItem := gio.NewMenuItem(L("Delete contact"), "app.deleteContact")
							deleteContactMenuItem.SetActionAndTargetValue("app.deleteContact", glib.NewVariantInt64(*contact.Id))
							menu.AppendItem(deleteContactMenuItem)

							editContactMenuItem := gio.NewMenuItem(L("Edit contact"), "app.editContact")
							editContactMenuItem.SetActionAndTargetValue("app.editContact", glib.NewVariantInt64(*contact.Id))
							menu.AppendItem(editContactMenuItem)

							menuButton.SetMenuModel(&menu.MenuModel)

							r.AddSuffix(&menuButton.Widget)

							r.AddSuffix(&gtk.NewImageFromIconName("go-next-symbolic").Widget)

							r.SetActivatable(true)

							contactsListBox.Append(&r.PreferencesRow.ListBoxRow.Widget)
						}
					} else {
						contactsAddButton.SetVisible(false)
						contactsSearchButton.SetVisible(false)
					}
				}()

			case resources.PageContactsView:
				go func() {
					enableContactsViewLoading()
					defer disableContactsViewLoading()

					redirected, c, _, err := authorize(
						ctx,

						true,
					)
					if err != nil {
						log.Warn("Could not authorize user for contacts view page", "err", err)

						handleContactsViewError(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Getting contact", "id", selectedContactID)

					res, err := c.GetContactWithResponse(ctx, int64(selectedContactID))
					if err != nil {
						handleContactsViewError(err)

						return
					}

					log.Debug("Got contact", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handleContactsViewError(errors.New(res.Status()))

						return
					}

					defer clearContactsViewError()

					contactsViewEditButton.SetActionTargetValue(glib.NewVariantInt64(*res.JSON200.Entry.Id))
					contactsViewDeleteButton.SetActionTargetValue(glib.NewVariantInt64(*res.JSON200.Entry.Id))

					title := *res.JSON200.Entry.FirstName + " " + *res.JSON200.Entry.LastName
					if *res.JSON200.Entry.Nickname != "" {
						title += " (" + *res.JSON200.Entry.Nickname + ")"
					}
					contactsViewPageTitle.SetTitle(title)

					subtitle := ""
					if *res.JSON200.Entry.Email != "" {
						subtitle = string(*res.JSON200.Entry.Email)
					}
					if string(*res.JSON200.Entry.Email) != "" && string(*res.JSON200.Entry.Pronouns) != "" {
						subtitle += " | " + *res.JSON200.Entry.Pronouns
					} else if string(*res.JSON200.Entry.Pronouns) != "" {
						subtitle = *res.JSON200.Entry.Pronouns
					}
					if subtitle != "" {
						contactsViewPageTitle.SetSubtitle(subtitle)
					}

					var (
						birthday = res.JSON200.Entry.Birthday
						address  = res.JSON200.Entry.Address
						notes    = res.JSON200.Entry.Notes
					)
					if (birthday != nil) || (*address != "") || (*notes != "") {
						if birthday != nil {
							contactsViewBirthdayRow.SetVisible(true)
							contactsViewBirthdayRow.SetSubtitle(glibDateTimeFromGo(birthday.Time).Format("%x"))
						} else {
							contactsViewBirthdayRow.SetVisible(false)
						}

						if *address != "" {
							contactsViewAddressRow.SetVisible(true)
							contactsViewAddressRow.SetSubtitle(*address)
						} else {
							contactsViewAddressRow.SetVisible(false)
						}

						if *notes != "" {
							contactsViewNotesRow.SetVisible(true)
							contactsViewNotesRow.SetSubtitle(*notes)
						} else {
							contactsViewNotesRow.SetVisible(false)
						}

						contactsViewOptionalFieldsPreferencesGroup.SetVisible(true)
					} else {
						contactsViewOptionalFieldsPreferencesGroup.SetVisible(false)
					}

					onValidateDebtsCreateDialogForm()

					debtsCreateDialogAddButton.SetActionTargetValue(glib.NewVariantInt64(*res.JSON200.Entry.Id))

					debtsCreateDialogTitle.SetSubtitle(*res.JSON200.Entry.FirstName + " " + *res.JSON200.Entry.LastName)

					debtsCreateDialogYouOweActionRow.SetTitle(L(fmt.Sprintf("_You owe %v", *res.JSON200.Entry.FirstName)))
					debtsCreateDialogYouOweActionRow.SetUseUnderline(true)
					debtsCreateDialogTheyOweActionRow.SetTitle(L(fmt.Sprintf("%v ow_es you", *res.JSON200.Entry.FirstName)))
					debtsCreateDialogTheyOweActionRow.SetUseUnderline(true)

					contactsViewDebtsListBox.RemoveAll()

					for _, debt := range *res.JSON200.Debts {
						r := adw.NewActionRow()

						subtitle := ""
						if *debt.Amount <= 0.0 {
							subtitle = L(fmt.Sprintf("You owe %v %v %v", *res.JSON200.Entry.FirstName, math.Abs(float64(*debt.Amount)), *debt.Currency))
						} else {
							subtitle = L(fmt.Sprintf("%v owes you %v %v", *res.JSON200.Entry.FirstName, math.Abs(float64(*debt.Amount)), *debt.Currency))
						}

						r.SetTitle(subtitle)

						if *debt.Description != "" {
							r.SetSubtitle(*debt.Description)
						}

						menuButton := gtk.NewMenuButton()
						menuButton.SetValign(gtk.AlignCenterValue)
						menuButton.SetIconName("view-more-symbolic")
						menuButton.AddCssClass("flat")

						menu := gio.NewMenu()

						settleDebtMenuItem := gio.NewMenuItem(L("Settle debt"), "app.settleDebt")
						settleDebtMenuItem.SetActionAndTargetValue("app.settleDebt", glib.NewVariantInt64(*debt.Id))

						menu.AppendItem(settleDebtMenuItem)

						editDebtMenuItem := gio.NewMenuItem(L("Edit debt"), "app.editDebt")
						editDebtMenuItem.SetActionAndTargetValue("app.editDebt", glib.NewVariantInt64(*debt.Id))

						menu.AppendItem(editDebtMenuItem)

						menuButton.SetMenuModel(&menu.MenuModel)

						r.AddSuffix(&menuButton.Widget)

						contactsViewDebtsListBox.Append(&r.PreferencesRow.ListBoxRow.Widget)
					}

					addDebtButton := adw.NewButtonRow()
					addDebtButton.SetStartIconName("list-add-symbolic")
					addDebtButton.SetTitle(L("Add a _debt"))
					addDebtButton.SetUseUnderline(true)

					onActivated := func(_ adw.ButtonRow) {
						debtsCreateDialog.Present(&w.ApplicationWindow.Window.Widget)
						debtsCreateDialogAmountInput.GrabFocus()
					}
					addDebtButton.ConnectActivated(&onActivated)

					contactsViewDebtsListBox.Append(&addDebtButton.PreferencesRow.ListBoxRow.Widget)

					onValidateActivitiesCreateDialogForm()

					activitiesCreateDialogAddButton.SetActionTargetValue(glib.NewVariantInt64(*res.JSON200.Entry.Id))

					activitiesCreateDialogTitle.SetSubtitle(*res.JSON200.Entry.FirstName + " " + *res.JSON200.Entry.LastName)

					contactsViewActivitiesListBox.RemoveAll()

					for _, activity := range *res.JSON200.Activities {
						r := adw.NewActionRow()

						r.SetTitle(*activity.Name)
						r.SetSubtitle(glibDateTimeFromGo(activity.Date.Time).Format("%x"))

						r.SetName("/activities/view?id=" + strconv.Itoa(int(*activity.Id)))

						menuButton := gtk.NewMenuButton()
						menuButton.SetValign(gtk.AlignCenterValue)
						menuButton.SetIconName("view-more-symbolic")
						menuButton.AddCssClass("flat")

						menu := gio.NewMenu()

						deleteActivityMenuItem := gio.NewMenuItem(L("Delete activity"), "app.deleteActivity")
						deleteActivityMenuItem.SetActionAndTargetValue("app.deleteActivity", glib.NewVariantInt64(*activity.Id))
						menu.AppendItem(deleteActivityMenuItem)

						editActivityMenuItem := gio.NewMenuItem(L("Edit activity"), "app.editActivity")
						editActivityMenuItem.SetActionAndTargetValue("app.editActivity", glib.NewVariantInt64(*activity.Id))
						menu.AppendItem(editActivityMenuItem)

						menuButton.SetMenuModel(&menu.MenuModel)

						r.AddSuffix(&menuButton.Widget)

						r.AddSuffix(&gtk.NewImageFromIconName("go-next-symbolic").Widget)

						r.SetActivatable(true)

						contactsViewActivitiesListBox.Append(&r.PreferencesRow.ListBoxRow.Widget)
					}

					addActivityButton := adw.NewButtonRow()
					addActivityButton.SetStartIconName("list-add-symbolic")
					addActivityButton.SetTitle(L("Add an ac_tivity"))
					addActivityButton.SetUseUnderline(true)

					onAddActivityActivated := func(_ adw.ButtonRow) {
						activitiesCreateDialog.Present(&w.ApplicationWindow.Window.Widget)
						activitiesCreateDialogNameInput.GrabFocus()
					}
					addActivityButton.ConnectActivated(&onAddActivityActivated)

					addActivityButton.SetActivatable(true)

					contactsViewActivitiesListBox.Append(&addActivityButton.PreferencesRow.ListBoxRow.Widget)
				}()

			case resources.PageActivitiesView:
				go func() {
					enableActivitiesViewLoading()
					defer disableActivitiesViewLoading()

					redirected, c, _, err := authorize(
						ctx,

						true,
					)
					if err != nil {
						log.Warn("Could not authorize user for activities view page", "err", err)

						handleActivitiesViewError(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Getting activity", "id", selectedActivityID)

					res, err := c.GetActivityWithResponse(ctx, int64(selectedActivityID))
					if err != nil {
						handleActivitiesViewError(err)

						return
					}

					log.Debug("Got activity", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handleActivitiesViewError(errors.New(res.Status()))

						return
					}

					activitiesViewEditButton.SetActionTargetValue(glib.NewVariantInt64(*res.JSON200.ActivityId))
					activitiesViewDeleteButton.SetActionTargetValue(glib.NewVariantInt64(*res.JSON200.ActivityId))

					activitiesViewPageTitle.SetTitle(*res.JSON200.Name)
					activitiesViewPageTitle.SetSubtitle(glibDateTimeFromGo(res.JSON200.Date.Time).Format("%x"))

					var buf bytes.Buffer
					if err := md.Convert([]byte(*res.JSON200.Description), &buf); err != nil {
						log.Warn("Could not render Markdown for activities view page", "err", err)

						handleActivitiesViewError(err)

						return
					}

					if description := *res.JSON200.Description; description != "" {
						idleAdd(func() {
							activitiesViewPageBodyWebView.LoadHtml(renderedMarkdownHTMLPrefix+buf.String(), "about:blank")
						})
					} else {
						idleAdd(func() {
							activitiesViewPageBodyWebView.LoadHtml(renderedMarkdownHTMLPrefix+L("No description provided."), "about:blank")
						})
					}

					defer clearActivitiesViewError()
				}()

			case resources.PageActivitiesEdit:
				go func() {
					enableActivitiesEditLoading()
					defer disableActivitiesEditLoading()

					redirected, c, _, err := authorize(
						ctx,

						true,
					)
					if err != nil {
						log.Warn("Could not authorize user for activities edit page", "err", err)

						handleActivitiesEditError(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Getting activity", "id", selectedActivityID)

					res, err := c.GetActivityWithResponse(ctx, int64(selectedActivityID))
					if err != nil {
						handleActivitiesEditError(err)

						return
					}

					log.Debug("Got activity", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handleActivitiesEditError(errors.New(res.Status()))

						return
					}

					defer clearActivitiesEditError()

					activitiesEditPageSaveButton.SetActionTargetValue(glib.NewVariantInt64(*res.JSON200.ActivityId))

					activitiesEditPageTitle.SetSubtitle(*res.JSON200.FirstName + " " + *res.JSON200.LastName)

					activitiesEditPageNameInput.SetText(*res.JSON200.Name)
					activitiesEditPageDateInput.SetText(glibDateTimeFromGo(res.JSON200.Date.Time).Format("%x"))

					setValidationSuffixVisible(&activitiesEditPageDateInput, &activitiesEditPageDateWarningButton, false)

					activitiesEditPageDescriptionExpander.SetExpanded(*res.JSON200.Description != "")
					activitiesEditPageDescriptionInput.GetBuffer().SetText(*res.JSON200.Description, -1)
				}()

			case resources.PageDebtsEdit:
				go func() {
					enableDebtsEditLoading()
					defer disableDebtsEditLoading()

					redirected, c, _, err := authorize(
						ctx,

						true,
					)
					if err != nil {
						log.Warn("Could not authorize user for debts edit page", "err", err)

						handleDebtsEditError(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Getting contact", "id", selectedContactID)

					res, err := c.GetContactWithResponse(ctx, int64(selectedContactID))
					if err != nil {
						handleDebtsEditError(err)

						return
					}

					log.Debug("Got contact", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handleDebtsEditError(errors.New(res.Status()))

						return
					}

					var debt *api.Debt
					for _, d := range *res.JSON200.Debts {
						if *d.Id == int64(selectedDebtID) {
							debt = &d

							break
						}
					}

					if debt == nil {
						handleDebtsEditError(errDebtDoesNotExist)

						return
					}

					defer clearDebtsEditError()

					debtsEditPageSaveButton.SetActionTargetValue(glib.NewVariantInt64(*debt.Id))

					debtsEditPageTitle.SetSubtitle(*res.JSON200.Entry.FirstName + " " + *res.JSON200.Entry.LastName)

					debtsEditPageYouOweActionRow.SetTitle(L(fmt.Sprintf("_You owe %v", *res.JSON200.Entry.FirstName)))
					debtsEditPageYouOweActionRow.SetUseUnderline(true)
					debtsEditPageTheyOweActionRow.SetTitle(L(fmt.Sprintf("%v ow_es you", *res.JSON200.Entry.FirstName)))
					debtsEditPageTheyOweActionRow.SetUseUnderline(true)

					debtsEditPageYouOweRadio.SetActive(*debt.Amount < 0)
					debtsEditPageAmountInput.SetValue(math.Abs(float64(*debt.Amount)))
					debtsEditPageCurrencyInput.SetText(*debt.Currency)

					debtsEditPageDescriptionExpander.SetExpanded(*debt.Description != "")
					debtsEditPageDescriptionInput.GetBuffer().SetText(*debt.Description, -1)
				}()

			case resources.PageContactsEdit:
				go func() {
					enableContactsEditLoading()
					defer disableContactsEditLoading()

					redirected, c, _, err := authorize(
						ctx,

						true,
					)
					if err != nil {
						log.Warn("Could not authorize user for contacts edit page", "err", err)

						handleContactsEditError(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Getting contact", "id", selectedContactID)

					res, err := c.GetContactWithResponse(ctx, int64(selectedContactID))
					if err != nil {
						handleContactsEditError(err)

						return
					}

					log.Debug("Got contact", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handleContactsEditError(errors.New(res.Status()))

						return
					}

					defer clearContactsEditError()

					contactsEditPageSaveButton.SetActionTargetValue(glib.NewVariantInt64(*res.JSON200.Entry.Id))

					contactsEditPageTitle.SetSubtitle(*res.JSON200.Entry.FirstName + " " + *res.JSON200.Entry.LastName)

					contactsEditPageFirstNameInput.SetText(*res.JSON200.Entry.FirstName)
					contactsEditPageLastNameInput.SetText(*res.JSON200.Entry.LastName)
					contactsEditPageNicknameInput.SetText(*res.JSON200.Entry.Nickname)
					contactsEditPageEmailInput.SetText(string(*res.JSON200.Entry.Email))
					contactsEditPagePronounsInput.SetText(string(*res.JSON200.Entry.Pronouns))

					var (
						birthday = res.JSON200.Entry.Birthday
						address  = res.JSON200.Entry.Address
						notes    = res.JSON200.Entry.Notes
					)
					if birthday != nil {
						contactsEditPageBirthdayInput.SetText(glibDateTimeFromGo(birthday.Time).Format("%x"))
					}

					if *address != "" {
						contactsEditPageAddressExpander.SetExpanded(true)
						contactsEditPageAddressInput.GetBuffer().SetText(*address, -1)
					} else {
						contactsEditPageAddressExpander.SetExpanded(false)
					}

					if *notes != "" {
						contactsEditPageNotesExpander.SetExpanded(true)
						contactsEditPageNotesInput.GetBuffer().SetText(*notes, -1)
					} else {
						contactsEditPageNotesExpander.SetExpanded(false)
					}
				}()

			case resources.PageJournalEntries:
				go func() {
					enableJournalEntriesLoading()
					defer disableJournalEntriesLoading()

					redirected, c, _, err := authorize(
						ctx,

						true,
					)
					if err != nil {
						log.Warn("Could not authorize user for journal entries page", "err", err)

						handleJournalEntriesError(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Listing journal entries")

					res, err := c.GetJournalEntriesWithResponse(ctx)
					if err != nil {
						handleJournalEntriesError(err)

						return
					}

					log.Debug("Got journal entries", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handleJournalEntriesError(errors.New(res.Status()))

						return
					}

					defer clearJournalEntriesError()

					onValidateJournalEntriesCreateDialogForm()

					journalEntriesListBox.RemoveAll()

					journalEntriesCount = len(*res.JSON200)
					if journalEntriesCount > 0 {
						journalEntriesAddButton.SetVisible(true)
						journalEntriesSearchButton.SetVisible(true)

						for _, journalEntry := range *res.JSON200 {
							r := adw.NewActionRow()

							r.SetTitle(*journalEntry.Title)

							subtitle := glibDateTimeFromGo(*journalEntry.Date).Format("%x") + " | "
							switch *journalEntry.Rating {
							case 3:
								subtitle += L("Great")

							case 2:
								subtitle += L("OK")

							case 1:
								subtitle += L("Bad")
							}
							r.SetSubtitle(subtitle)

							r.SetName("/journal/view?id=" + strconv.Itoa(int(*journalEntry.Id)))

							menuButton := gtk.NewMenuButton()
							menuButton.SetValign(gtk.AlignCenterValue)
							menuButton.SetIconName("view-more-symbolic")
							menuButton.AddCssClass("flat")

							menu := gio.NewMenu()

							deleteContactMenuItem := gio.NewMenuItem(L("Delete journal entry"), "app.deleteJournalEntry")
							deleteContactMenuItem.SetActionAndTargetValue("app.deleteJournalEntry", glib.NewVariantInt64(*journalEntry.Id))
							menu.AppendItem(deleteContactMenuItem)

							editContactMenuItem := gio.NewMenuItem(L("Edit journal entry"), "app.editJournalEntry")
							editContactMenuItem.SetActionAndTargetValue("app.editJournalEntry", glib.NewVariantInt64(*journalEntry.Id))
							menu.AppendItem(editContactMenuItem)

							menuButton.SetMenuModel(&menu.MenuModel)

							r.AddSuffix(&menuButton.Widget)

							r.AddSuffix(&gtk.NewImageFromIconName("go-next-symbolic").Widget)

							r.SetActivatable(true)

							journalEntriesListBox.Append(&r.PreferencesRow.ListBoxRow.Widget)
						}
					} else {
						journalEntriesAddButton.SetVisible(false)
						journalEntriesSearchButton.SetVisible(false)
					}
				}()

			case resources.PageJournalEntriesView:
				go func() {
					enableJournalEntriesViewLoading()
					defer disableJournalEntriesViewLoading()

					redirected, c, _, err := authorize(
						ctx,

						true,
					)
					if err != nil {
						log.Warn("Could not authorize user for journal entries view page", "err", err)

						handleJournalEntriesViewError(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Getting journal entry", "id", selectedJournalEntryID)

					res, err := c.GetJournalEntryWithResponse(ctx, int64(selectedJournalEntryID))
					if err != nil {
						handleJournalEntriesViewError(err)

						return
					}

					log.Debug("Got journal entry", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handleJournalEntriesViewError(errors.New(res.Status()))

						return
					}

					journalEntriesViewEditButton.SetActionTargetValue(glib.NewVariantInt64(*res.JSON200.Id))
					journalEntriesViewDeleteButton.SetActionTargetValue(glib.NewVariantInt64(*res.JSON200.Id))

					journalEntriesViewPageTitle.SetTitle(*res.JSON200.Title)
					subtitle := glibDateTimeFromGo(*res.JSON200.Date).Format("%x") + " | "
					switch *res.JSON200.Rating {
					case 3:
						subtitle += L("Great")

					case 2:
						subtitle += L("OK")

					case 1:
						subtitle += L("Bad")
					}
					journalEntriesViewPageTitle.SetSubtitle(subtitle)

					var buf bytes.Buffer
					if err := md.Convert([]byte(*res.JSON200.Body), &buf); err != nil {
						log.Warn("Could not render Markdown for journal entries view page", "err", err)

						handleJournalEntriesViewError(err)

						return
					}

					idleAdd(func() {
						journalEntriesViewPageBodyWebView.LoadHtml(renderedMarkdownHTMLPrefix+buf.String(), "about:blank")
					})

					defer clearJournalEntriesViewError()
				}()

			case resources.PageJournalEntriesEdit:
				go func() {
					enableJournalEntriesEditLoading()
					defer disableJournalEntriesEditLoading()

					redirected, c, _, err := authorize(
						ctx,

						true,
					)
					if err != nil {
						log.Warn("Could not authorize user for journal entry edit page", "err", err)

						handleJournalEntriesEditError(err)

						return
					} else if redirected {
						return
					}

					log.Debug("Getting journal entry", "id", selectedJournalEntryID)

					res, err := c.GetJournalEntryWithResponse(ctx, int64(selectedJournalEntryID))
					if err != nil {
						handleJournalEntriesEditError(err)

						return
					}

					log.Debug("Got journal entry", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						handleJournalEntriesEditError(errors.New(res.Status()))

						return
					}

					defer clearJournalEntriesEditError()

					journalEntriesEditPageSaveButton.SetActionTargetValue(glib.NewVariantInt64(*res.JSON200.Id))

					journalEntriesEditPageTitle.SetSubtitle(*res.JSON200.Title)

					journalEntriesEditPageRatingToggleGroup.SetActive(3 - uint(*res.JSON200.Rating)) // The toggle group is zero-indexed, but the rating is one-indexed

					journalEntriesEditPageTitleInput.SetText(*res.JSON200.Title)

					journalEntriesEditPageBodyExpander.SetExpanded(true)
					journalEntriesEditPageBodyInput.GetBuffer().SetText(*res.JSON200.Body, -1)
				}()
			}
		}

		connectListBoxRowActivated(&contactsListBox, func(row *gtk.ListBoxRow) {
			if row != nil {
				var actionRow adw.ActionRow
				row.Cast(&actionRow)
				u, err := url.Parse(actionRow.GetName())
				if err != nil {
					log.Warn("Could not parse contact row URL", "err", err)

					onPanic(err)

					return
				}

				rid := u.Query().Get("id")
				if strings.TrimSpace(rid) == "" {
					log.Warn("Could not get ID from contact row URL", "err", errMissingContactID)

					onPanic(errMissingContactID)

					return
				}

				id, err := strconv.Atoi(rid)
				if err != nil {
					log.Warn("Could not parse ID from contact row URL", "err", errInvalidContactID)

					onPanic(errInvalidContactID)

					return
				}

				selectedContactID = id

				homeNavigation.PushByTag(resources.PageContactsView)
			}
		})

		connectListBoxRowActivated(&contactsViewActivitiesListBox, func(row *gtk.ListBoxRow) {
			if row != nil {
				var actionRow adw.ActionRow
				row.Cast(&actionRow)

				u, err := url.Parse(actionRow.GetName())
				if err != nil {
					log.Warn("Could not parse activity row URL", "err", err)

					onPanic(err)

					return
				}

				rid := u.Query().Get("id")
				if strings.TrimSpace(rid) == "" {
					log.Warn("Could not get ID from activity row URL", "err", errMissingActivityID)

					onPanic(errMissingActivityID)

					return
				}

				id, err := strconv.Atoi(rid)
				if err != nil {
					log.Warn("Could not parse ID from activity row URL", "err", errInvalidActivityID)

					onPanic(errInvalidActivityID)

					return
				}

				selectedActivityID = id

				homeNavigation.PushByTag(resources.PageActivitiesView)
			}
		})

		connectListBoxRowActivated(&journalEntriesListBox, func(row *gtk.ListBoxRow) {
			if row != nil {
				var actionRow adw.ActionRow
				row.Cast(&actionRow)
				u, err := url.Parse(actionRow.GetName())
				if err != nil {
					log.Warn("Could not parse journal entry row URL", "err", err)

					onPanic(err)

					return
				}

				rid := u.Query().Get("id")
				if strings.TrimSpace(rid) == "" {
					log.Warn("Could not get ID from journal entry row URL", "err", errMissingJournalEntryID)

					onPanic(errMissingJournalEntryID)

					return
				}

				id, err := strconv.Atoi(rid)
				if err != nil {
					log.Warn("Could not parse ID from journal entry row URL", "err", errInvalidJournaEntrylID)

					onPanic(errInvalidJournaEntrylID)

					return
				}

				selectedJournalEntryID = id

				homeNavigation.PushByTag(resources.PageJournalEntriesView)
			}
		})

		connectNavigationViewPopped(&homeNavigation, func(page *adw.NavigationPage) {
			onHomeNavigation()

			var (
				tag = page.GetTag()
				log = log.With("tag", tag)
			)

			log.Info("Handling popped page")

			switch tag {
			case resources.PageContactsView:
				contactsViewPageTitle.SetTitle("")
				contactsViewPageTitle.SetSubtitle("")

			case resources.PageActivitiesView:
				activitiesViewPageTitle.SetTitle("")
				activitiesViewPageTitle.SetSubtitle("")

			case resources.PageActivitiesEdit:
				activitiesEditPageTitle.SetSubtitle("")

				activitiesEditPageNameInput.SetText("")
				activitiesEditPageDateInput.SetText("")

				setValidationSuffixVisible(&activitiesEditPageDateInput, &activitiesEditPageDateWarningButton, false)

				activitiesEditPageDescriptionExpander.SetExpanded(false)
				activitiesEditPageDescriptionInput.GetBuffer().SetText("", 0)

			case resources.PageDebtsEdit:
				debtsEditPageTitle.SetSubtitle("")

				debtsEditPageYouOweActionRow.SetTitle("")
				debtsEditPageTheyOweActionRow.SetTitle("")

				debtsEditPageYouOweRadio.SetActive(true)

				debtsEditPageAmountInput.SetValue(0)
				debtsEditPageCurrencyInput.SetText("")

				debtsEditPageDescriptionExpander.SetExpanded(false)
				debtsEditPageDescriptionInput.GetBuffer().SetText("", 0)

			case resources.PageContactsEdit:
				contactsEditPageTitle.SetSubtitle("")

				contactsEditPageFirstNameInput.SetText("")
				contactsEditPageLastNameInput.SetText("")
				contactsEditPageNicknameInput.SetText("")
				contactsEditPageEmailInput.SetText("")
				contactsEditPagePronounsInput.SetText("")

				contactsEditPageBirthdayInput.SetText("")

				contactsEditPageAddressExpander.SetExpanded(false)
				contactsEditPageAddressInput.GetBuffer().SetText("", 0)

				contactsEditPageNotesExpander.SetExpanded(false)
				contactsEditPageNotesInput.GetBuffer().SetText("", 0)

				setValidationSuffixVisible(&contactsEditPageEmailInput, &contactsEditPageEmailWarningButton, false)
			}
		})
		connectNavigationViewPushed(&homeNavigation, onHomeNavigation)
		connectNavigationViewReplaced(&homeNavigation, onHomeNavigation)

		connectListBoxRowActivated(&homeSidebarListbox, func(row *gtk.ListBoxRow) {
			var actionRow adw.ActionRow
			row.Cast(&actionRow)
			homeNavigation.ReplaceWithTags([]string{actionRow.GetName()}, 1)
		})

		onNavigation := func() {
			var (
				tag = nv.GetVisiblePage().GetTag()
				log = log.With("tag", tag)
			)

			log.Info("Handling page")

			switch tag {
			case resources.PageIndex:
				go func() {
					if err := checkSenbaraServerConfiguration(); err != nil {
						log.Info("Could not check Senbara server configuration, redirecting to login", "err", err)

						updateDeregisterClientActionEnabled()

						nv.PushByTag(resources.PageWelcome)

						return
					}

					if err := setupAuthn(false); err != nil {
						log.Info("Could not setup authn, redirecting to login", "err", err)

						nv.PushByTag(resources.PageWelcome)

						return
					}

					redirected, _, _, err := authorize(
						ctx,

						false,
					)
					if err != nil {
						log.Warn("Could not authorize user for index page", "err", err)

						onPanic(err)

						return
					} else if redirected {
						return
					}

					if settings.GetBoolean(resources.SettingAnonymousMode) {
						nv.PushByTag(resources.PagePreview)

						return
					}

					if strings.TrimSpace(u.Email) != "" {
						nv.PushByTag(resources.PageHome)

						return
					}

					nv.PushByTag(resources.PageWelcome)
				}()

			case resources.PageConfigServerURL:
				configServerURLContinueButton.SetSensitive(false)
				configServerURLContinueSpinner.SetVisible(true)
				configServerURLInput.GrabFocus()

				go func() {
					defer configServerURLContinueSpinner.SetVisible(false)

					if err := deregisterOIDCClient(); err != nil {
						onPanic(err)

						return
					}

					updateDeregisterClientActionEnabled()
					updateConfigServerURLContinueButtonSensitive()
				}()

			case resources.PagePreview:
				go func() {
					enablePreviewLoading()
					defer disablePreviewLoading()

					redirected, c, _, err := authorize(
						ctx,

						false,
					)
					if err != nil {
						log.Warn("Could not authorize user for preview page", "err", err)

						onPanic(err)

						return
					} else if redirected {
						return
					}

					settings.SetBoolean(resources.SettingAnonymousMode, true)

					log.Debug("Getting statistics")

					res, err := c.GetStatisticsWithResponse(ctx)
					if err != nil {
						onPanic(err)

						return
					}

					log.Debug("Got statistics", "status", res.StatusCode())

					if res.StatusCode() != http.StatusOK {
						onPanic(errors.New(res.Status()))

						return
					}

					previewContactsCountLabel.SetLabel(fmt.Sprintf("%v", *res.JSON200.ContactsCount))
					previewJournalEntriesCountLabel.SetLabel(fmt.Sprintf("%v", *res.JSON200.JournalEntriesCount))
				}()

			case resources.PageRegister:
				configInitialAccessTokenInput.SetText("")

			case resources.PageHome:
				go func() {
					if ok := refreshSidebarWithLatestSummary(); !ok {
						return
					}

					contactsRow := homeSidebarListbox.GetRowAtIndex(0)
					contactsRow.GrabFocus()
					homeSidebarListbox.SelectRow(contactsRow)

					homeNavigation.ReplaceWithTags([]string{resources.PageContacts}, 1)
				}()
			}
		}

		connectNavigationViewPopped(&nv, func(page *adw.NavigationPage) {
			onNavigation()

			var (
				tag = page.GetTag()
				log = log.With("tag", tag)
			)

			log.Info("Handling popped page")

			switch tag {
			case resources.PagePreview:
				enablePreviewLoading()
			}
		})
		connectNavigationViewPushed(&nv, onNavigation)
		connectNavigationViewReplaced(&nv, onNavigation)

		onNavigation()

		a.AddWindow(&w.Window)
	}
	a.ConnectActivate(&onActivate)

	onOpen := func(_ gio.Application, filesPtr uintptr, nFiles int, hint string) {
		if w.GoPointer() == 0 {
			a.Activate()
		} else {
			w.Present()
		}

		if nFiles == 0 {
			return
		}

		file := &gio.FileBase{
			Ptr: *(*uintptr)(unsafe.Pointer(filesPtr)), // Get first file from the array
		}

		u, err := url.Parse(file.GetUri())
		if err != nil {
			onPanic(err)

			return
		}

		authCode := u.Query().Get("code")
		state := u.Query().Get("state")

		log := log.With(
			"authCode", authCode != "",
			"state", state,
		)

		log.Debug("Handling user auth exchange")

		var (
			stateNonce,
			pkceCodeVerifier,
			oidcNonce string
		)
		sn, err := keyring.Get(resources.AppID, resources.SecretStateNonceKey)
		if err != nil {
			if !errors.Is(err, keyring.ErrNotFound) {
				log.Debug("Failed to read state nonce cookie", "error", err)

				onPanic(errors.Join(errCouldNotLogin, err))

				return
			}
		} else {
			stateNonce = sn
		}

		pcv, err := keyring.Get(resources.AppID, resources.SecretPKCECodeVerifierKey)
		if err != nil {
			if !errors.Is(err, keyring.ErrNotFound) {
				log.Debug("Failed to read PKCE code verifier cookie", "error", err)

				onPanic(errors.Join(errCouldNotLogin, err))

				return
			}
		} else {
			pkceCodeVerifier = pcv
		}

		on, err := keyring.Get(resources.AppID, resources.SecretOIDCNonceKey)
		if err != nil {
			if !errors.Is(err, keyring.ErrNotFound) {
				log.Debug("Failed to read OIDC nonce cookie", "error", err)

				onPanic(errors.Join(errCouldNotLogin, err))

				return
			}
		} else {
			oidcNonce = on
		}

		nextURL, signedOut, err := authner.Exchange(
			ctx,

			authCode,
			state,

			stateNonce,
			pkceCodeVerifier,
			oidcNonce,

			func(s string, t time.Time) error {
				// TODO: Handle expiry time
				return keyring.Set(resources.AppID, resources.SecretRefreshTokenKey, s)
			},
			func(s string, t time.Time) error {
				// TODO: Handle expiry time
				return keyring.Set(resources.AppID, resources.SecretIDTokenKey, s)
			},

			func() error {
				if err := keyring.Delete(resources.AppID, resources.SecretRefreshTokenKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
					return err
				}

				return nil
			},
			func() error {
				if err := keyring.Delete(resources.AppID, resources.SecretIDTokenKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
					return err
				}

				return nil
			},

			func() error {
				if err := keyring.Delete(resources.AppID, resources.SecretStateNonceKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
					return err
				}

				return nil
			},
			func() error {
				if err := keyring.Delete(resources.AppID, resources.SecretPKCECodeVerifierKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
					return err
				}

				return nil
			},
			func() error {
				if err := keyring.Delete(resources.AppID, resources.SecretOIDCNonceKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
					return err
				}

				return nil
			},
		)
		if err != nil {
			onPanic(err)

			return
		}

		// In the web version, redirecting to the home page after signing out is possible without
		// authn. In the GNOME version, that is not the case since the unauthenticated
		// page is a separate page from home, so we need to rewrite the path to distinguish
		// between the two manually
		if signedOut && nextURL == resources.PageHome {
			nextURL = resources.PageIndex
		}

		nv.ReplaceWithTags([]string{nextURL}, 1)
	}
	a.ConnectOpen(&onOpen)

	if code := a.Run(len(os.Args), os.Args); code > 0 {
		os.Exit(code)
	}
}
