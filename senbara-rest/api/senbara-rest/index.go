package senbaraForms

//go:generate tar czf code.tar.gz --exclude .git -C ../../.. .

import (
	_ "embed"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/pojntfx/senbara/senbara-common/pkg/persisters"
	"github.com/pojntfx/senbara/senbara-rest/pkg/controllers"
)

//go:embed code.tar.gz
var code []byte

var (
	p *persisters.Persister
	c *controllers.Controller
)

func SenbaraRESTHandler(
	w http.ResponseWriter,
	r *http.Request,
	c *controllers.Controller,
) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /journal", c.HandleJournal)
	mux.HandleFunc("GET /journal/{id}", c.HandleViewJournal)
	mux.HandleFunc("POST /journal", c.HandleCreateJournal)
	mux.HandleFunc("DELETE /journal/{id}", c.HandleDeleteJournal)
	mux.HandleFunc("PUT /journal/{id}", c.HandleUpdateJournal)

	mux.HandleFunc("GET /contacts", c.HandleContacts)
	mux.HandleFunc("GET /contacts/{id}", c.HandleViewContact)
	mux.HandleFunc("POST /contacts", c.HandleCreateContact)
	mux.HandleFunc("DELETE /contacts/{id}", c.HandleDeleteContact)
	mux.HandleFunc("PUT /contacts/{id}", c.HandleUpdateContact)

	mux.HandleFunc("POST /debts", c.HandleCreateDebt)
	mux.HandleFunc("DELETE /debts/{id}", c.HandleSettleDebt)
	mux.HandleFunc("PUT /debts/{id}", c.HandleUpdateDebt)

	mux.HandleFunc("GET /activities/{id}", c.HandleViewActivity)
	mux.HandleFunc("POST /activities", c.HandleCreateActivity)
	mux.HandleFunc("DELETE /activities/{id}", c.HandleDeleteActivity)
	mux.HandleFunc("PUT /activities/{id}", c.HandleUpdateActivity)

	mux.HandleFunc("GET /userdata", c.HandleUserData)
	mux.HandleFunc("POST /userdata", c.HandleCreateUserData)
	mux.HandleFunc("DELETE /userdata", c.HandleDeleteUserData)

	mux.HandleFunc("GET /code/", func(w http.ResponseWriter, r *http.Request) {
		c.HandleCode(w, r, code)
	})

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
		)

		if err := c.Init(r.Context()); err != nil {
			panic(err)
		}
	}

	SenbaraRESTHandler(w, r, c)
}
