package main

import (
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"os"
	"github.com/gin-gonic/gin/render"
	"path/filepath"

	// plugin package
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/justinas/nosurf"
	"gopkg.in/authboss.v0"
	// register authboss register module
	_ "gopkg.in/authboss.v0/register"
	// register authboss login module
	_ "gopkg.in/authboss.v0/auth"
	// to confirm authboss
	_ "gopkg.in/authboss.v0/confirm"
	// to lock user after N authentication failures
	_ "gopkg.in/authboss.v0/lock"
	_ "gopkg.in/authboss.v0/recover"
	_ "gopkg.in/authboss.v0/remember"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	aboauth "gopkg.in/authboss.v0/oauth2"
)

func initAuthBossPolicy(ab *authboss.Authboss)  {
	ab.Policies = []authboss.Validator{
		authboss.Rules{
			FieldName:       "email",
			Required:        true,
			AllowWhitespace: false,
		},
		authboss.Rules{
			FieldName:       "password",
			Required:        true,
			MinLength:       4,
			MaxLength:       8,
			AllowWhitespace: false,
		},
	}
}

func initAuthBossLayout(ab *authboss.Authboss, r *gin.Engine) {
	if os.Getenv(gin.ENV_GIN_MODE) == gin.ReleaseMode {
		ab.Layout = r.HTMLRender.(render.HTMLProduction).Template.Funcs(funcs).Lookup("authboss.tmpl")
	} else {
		html := r.HTMLRender.(render.HTMLDebug).Instance("authboss.tmpl", nil).(render.HTML)
		ab.Layout = html.Template.Funcs(template.FuncMap(funcs)).Lookup("authboss.tmpl")
		// ab.Layout.ExecuteTemplate(os.Stdout, "layout.html.tpl", nil)
	}
}

func initAuthBossParam(r *gin.Engine) *authboss.Authboss {
	ab := authboss.New()
	ab.Storer = database
	ab.OAuth2Storer = database
	ab.CookieStoreMaker = NewCookieStorer
	ab.SessionStoreMaker = NewSessionStorer
	ab.ViewsPath = filepath.Join("ab_views")
	//ab.RootURL = `http://localhost:5567`

	ab.LayoutDataMaker = layoutData
	ab.OAuth2Providers = map[string]authboss.OAuth2Provider{
		"google": authboss.OAuth2Provider{
			OAuth2Config: &oauth2.Config{
				ClientID:     ``,
				ClientSecret: ``,
				Scopes:       []string{`profile`, `email`},
				Endpoint:     google.Endpoint,
			},
			Callback: aboauth.Google,
		},
	}

	ab.MountPath = "/auth"
	ab.LogWriter = os.Stdout

	ab.XSRFName = "csrf_token"
	ab.XSRFMaker = func(_ http.ResponseWriter, r *http.Request) string {
		return nosurf.Token(r)
	}

	initAuthBossLayout(ab, r)
	ab.Mailer = authboss.LogMailer(os.Stdout)
	initAuthBossPolicy(ab)

	if err := ab.Init(); err != nil {
		// Handle error, don't let program continue to run
		log.Fatalln(err)
	}
	return ab
}

func initAuthBossRoute(r *gin.Engine) {
	cookieStoreKey, _ := base64.StdEncoding.DecodeString(`NpEPi8pEjKVjLGJ6kYCS+VTCzi6BUuDzU0wrwXyf5uDPArtlofn2AG6aTMiPmN3C909rsEWMNqJqhIVPGP3Exg==`)
	sessionStoreKey, _ := base64.StdEncoding.DecodeString(`AbfYwmmt8UCwUuhd9qvfNA9UCuN1cVcKJN1ofbiky6xCyyBj20whe40rJa3Su0WOWLWcPpO1taqJdsEI/65+JA==`)
	cookieStore = securecookie.New(cookieStoreKey, nil)
	sessionStore = sessions.NewCookieStore(sessionStoreKey)
	ab = initAuthBossParam(r)
	r.Any("/auth/*w", gin.WrapH(ab.NewRouter()))
}

func mainGin()  {

    r := gin.Default()

    r.Static("resources", "./resources")

    r.LoadHTMLGlob("resources/views/gin-gonic/*")

    r.GET("/", func(c *gin.Context) {
        data := layoutData(c.Writer, c.Request)
		data["content"] = "Authboss + Gin = Awesome!"
        c.HTML(http.StatusOK, "index.tmpl", data)
    })

    initAuthBossRoute(r)

    r.Run(":3000")
}
