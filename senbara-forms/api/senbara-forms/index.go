package senbaraForms

import (
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/pojntfx/senbara/senbara-forms/pkg/controllers"
	"github.com/pojntfx/senbara/senbara-forms/pkg/persisters"
	"github.com/pojntfx/senbara/senbara-forms/pkg/static"
)

var (
	p *persisters.Persister
	c *controllers.Controller
)

func SenbaraFormsHandler(
	w http.ResponseWriter,
	r *http.Request,
	c *controllers.Controller,
) {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.FS))))

	mux.HandleFunc("GET /journal", c.HandleJournal)
	mux.HandleFunc("GET /journal/add", c.HandleAddJournal)
	mux.HandleFunc("GET /journal/edit", c.HandleEditJournal)
	mux.HandleFunc("GET /journal/view", c.HandleViewJournal)

	mux.HandleFunc("POST /journal", c.HandleCreateJournal)
	mux.HandleFunc("POST /journal/delete", c.HandleDeleteJournal)
	mux.HandleFunc("POST /journal/update", c.HandleUpdateJournal)

	mux.HandleFunc("GET /contacts", c.HandleContacts)
	mux.HandleFunc("GET /contacts/add", c.HandleAddContact)
	mux.HandleFunc("GET /contacts/edit", c.HandleEditContact)
	mux.HandleFunc("GET /contacts/view", c.HandleViewContact)

	mux.HandleFunc("POST /contacts", c.HandleCreateContact)
	mux.HandleFunc("POST /contacts/delete", c.HandleDeleteContact)
	mux.HandleFunc("POST /contacts/update", c.HandleUpdateContact)

	mux.HandleFunc("GET /debts/add", c.HandleAddDebt)
	mux.HandleFunc("GET /debts/edit", c.HandleEditDebt)

	mux.HandleFunc("POST /debts", c.HandleCreateDebt)
	mux.HandleFunc("POST /debts/settle", c.HandleSettleDebt)
	mux.HandleFunc("POST /debts/update", c.HandleUpdateDebt)

	mux.HandleFunc("GET /activities/add", c.HandleAddActivity)
	mux.HandleFunc("GET /activities/view", c.HandleViewActivity)
	mux.HandleFunc("GET /activities/edit", c.HandleEditActivity)

	mux.HandleFunc("POST /activities", c.HandleCreateActivity)
	mux.HandleFunc("POST /activities/delete", c.HandleDeleteActivity)
	mux.HandleFunc("POST /activities/update", c.HandleUpdateActivity)

	mux.HandleFunc("GET /userdata", c.HandleUserData)

	mux.HandleFunc("POST /userdata", c.HandleCreateUserData)
	mux.HandleFunc("POST /userdata/delete", c.HandleDeleteUserData)

	mux.HandleFunc("GET /login", c.HandleLogin)
	mux.HandleFunc("GET /authorize", c.HandleAuthorize)

	// mux.Handle("GET /code/", http.StripPrefix("/code/", http.FileServer(http.FS(senbaraForms.FS))))

	mux.HandleFunc("/", c.HandleIndex)

	mux.ServeHTTP(w, r)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = r.URL.Query().Get("path")

	if p == nil {
		p = persisters.NewPersister(os.Getenv("POSTGRES_URL"))

		if err := p.Init(); err != nil {
			panic(err)
		}
	}

	if c == nil {
		c = controllers.NewController(
			p,

			os.Getenv("OIDC_ISSUER"),
			os.Getenv("OIDC_CLIENT_ID"),
			os.Getenv("OIDC_REDIRECT_URL"),

			os.Getenv("PRIVACY_URL"),
			os.Getenv("IMPRINT_URL"),
		)

		if err := c.Init(r.Context()); err != nil {
			panic(err)
		}
	}

	SenbaraFormsHandler(w, r, c)
}
