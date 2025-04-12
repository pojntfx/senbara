package main

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	gcore "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/pojntfx/senbara/senbara-gnome/assets/resources"
	"github.com/pojntfx/senbara/senbara-gnome/config/locales"
	"github.com/rymdport/portal/openuri"
	"github.com/zalando/go-keyring"
)

const (
	refreshTokenKey = "refresh_token"
)

func main() {
	i18t, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(i18t)

	if err := fs.WalkDir(locales.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if err := os.MkdirAll(filepath.Join(i18t, path), os.ModePerm); err != nil {
				return err
			}

			return nil
		}

		src, err := locales.FS.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(filepath.Join(i18t, path))
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}

	gcore.InitI18n("default", i18t)

	r, err := gio.NewResourceFromData(glib.NewBytesWithGo(resources.ResourceContents))
	if err != nil {
		panic(err)
	}
	gio.ResourcesRegister(r)

	st, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(st)

	sc, err := r.LookupData(resources.ResourceGSchemasCompiledPath, gio.ResourceLookupFlagsNone)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(filepath.Join(st, path.Base(resources.ResourceGSchemasCompiledPath)), sc.Data(), os.ModePerm); err != nil {
		panic(err)
	}

	if err := os.Setenv("GSETTINGS_SCHEMA_DIR", st); err != nil {
		panic(err)
	}

	_ = gio.NewSettings(resources.AppID)

	c := gtk.NewCSSProvider()
	c.LoadFromResource(resources.ResourceIndexCSSPath)

	a := adw.NewApplication(resources.AppID, gio.ApplicationDefaultFlags)
	a.ConnectActivate(func() {
		b := gtk.NewBuilderFromResource(resources.ResourceWindowUIPath)

		w := b.GetObject("main-window").Cast().(*adw.Window)

		nv := b.GetObject("main-navigation").Cast().(*adw.NavigationView)

		lb := b.GetObject("login-button").Cast().(*gtk.Button)
		lb.ConnectClicked(func() {
			nv.PushByTag("privacy-policy")
		})

		ppcb := b.GetObject("privacy-policy-checkbutton").Cast().(*gtk.CheckButton)
		cb := b.GetObject("continue-button").Cast().(*gtk.Button)

		ppcb.ConnectToggled(func() {
			cb.SetSensitive(ppcb.Active())
		})

		nv.ConnectPopped(func(page *adw.NavigationPage) {
			if page.Tag() == "privacy-policy" {
				ppcb.SetActive(false)
			}
		})

		cb.ConnectClicked(func() {
			nv.PushByTag("exchange")
		})

		nv.ConnectPushed(func() {
			if nv.VisiblePage().Tag() == "exchange" {
				if err := openuri.OpenURI("", "https://example.com/", nil); err != nil {
					panic(err)
				}

				time.AfterFunc(time.Second*2, func() {
					if err := keyring.Set(resources.AppID, refreshTokenKey, "testvalue"); err != nil {
						panic(err)
					}

					if nv.VisiblePage().Tag() == "exchange" {
						nv.PushByTag("home")
					}
				})
			}
		})

		logoutAction := gio.NewSimpleAction("logout", nil)
		logoutAction.ConnectActivate(func(parameter *glib.Variant) {
			if err := keyring.Delete(resources.AppID, refreshTokenKey); err != nil {
				panic(err)
			}

			nv.PopToTag("loading-config")
		})
		a.AddAction(logoutAction)

		aboutAction := gio.NewSimpleAction("about", nil)
		aboutAction.ConnectActivate(func(parameter *glib.Variant) {
			log.Println("Showing about screen")
		})
		a.AddAction(aboutAction)

		gtk.StyleContextAddProviderForDisplay(
			gdk.DisplayGetDefault(),
			c,
			gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
		)

		hydrateFromConfig := func() {
			// TODO: Check if existing auth configuration works here, if not try to renew
			// ID token or sign in again. If no configuration is set or it is not recoverable,
			// clear configuration (although the library should do that automatically) and
			// continue to login page for setup.
			time.AfterFunc(time.Millisecond*500, func() {
				if _, err := keyring.Get(resources.AppID, refreshTokenKey); err != nil {
					if errors.Is(err, keyring.ErrNotFound) {
						nv.PushByTag("login")
					} else {
						panic(err)
					}
				} else {
					nv.PushByTag("home")
				}
			})
		}

		nv.ConnectPushed(func() {
			if nv.VisiblePage().Tag() == "loading-config" {
				hydrateFromConfig()
			}
		})

		nv.ConnectPopped(func(page *adw.NavigationPage) {
			if nv.VisiblePage().Tag() == "loading-config" {
				hydrateFromConfig()
			}
		})

		hydrateFromConfig()

		a.AddWindow(&w.Window)
	})

	if code := a.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
