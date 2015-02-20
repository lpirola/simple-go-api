// main.go
package main

import (
    "github.com/gorilla/mux"
    "net/http"
    "github.com/codegangsta/negroni"
    "github.com/unrolled/render"  // or "gopkg.in/unrolled/render.v1"
    oauth2 "github.com/goincremental/negroni-oauth2"
    sessions "github.com/goincremental/negroni-sessions"
    "github.com/goincremental/negroni-sessions/cookiestore"
    "github.com/joho/godotenv"
    "os"
)

func main() {
    log := negroni.NewLogger()
    err := godotenv.Load()

    if err != nil {
        log.Fatal("Error loading .env file")
    }

    client_id := os.Getenv("GOOGLE_CLIENT_ID")
    client_secret := os.Getenv("GOOGLE_CLIENT_SECRET")
    refresh_url := os.Getenv("GOOGLE_REFRESH_URL")

    n := negroni.Classic()
    r := render.New()
    router := mux.NewRouter()

    router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

        token := oauth2.GetToken(req)
        if token == nil || !token.Valid() {
            r.JSON(w, http.StatusOK, map[string]string{"message": "not logged in, or the access token is expired"})
            return
        }
        r.JSON(w, http.StatusOK, map[string]string{"message": "logged in"})
        return
    })

    secureMux := http.NewServeMux()

    // Routes that require a logged in user
    // can be protected by using a separate route handler
    // If the user is not authenticated, they will be
    // redirected to the login path.
    secureMux.HandleFunc("/auth", func(w http.ResponseWriter, req *http.Request) {
        token := oauth2.GetToken(req)
        r.JSON(w, http.StatusOK, map[string]string{"token": token.Access()})
    })

    secure := negroni.New()
    secure.Use(oauth2.LoginRequired())
    secure.UseHandler(secureMux)

    router.Handle("/auth", secure)

    n.Use(sessions.Sessions("my_session", cookiestore.New([]byte("secret123"))))
    n.Use(oauth2.Google(&oauth2.Config{
        ClientID:     client_id,
        ClientSecret: client_secret,
        RedirectURL:  refresh_url,
        Scopes:       []string{"https://www.googleapis.com/auth/drive"},
    }))

    n.UseHandler(router)
    n.Run(":3001")
}