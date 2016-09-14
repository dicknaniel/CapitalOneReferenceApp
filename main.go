package main

import (
  "io/ioutil"
  "fmt"
  "net/http"
  "golang.org/x/oauth2"
)

const htmlIndex = `<html><body>
<a href="/login">login</a>
</body></html>
`

func main() {
    http.HandleFunc("/", handleMain)
    http.HandleFunc("/login", handleLogin)
    http.HandleFunc("/redirect", handleRedirect)
    fmt.Println(http.ListenAndServe(":3000", nil))
}

func handleMain(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, htmlIndex)
}

var (
    oAuthConfig = &oauth2.Config{
        RedirectURL:    "http://localhost:3000/redirect",
        ClientID:     "",
        ClientSecret: "",
        Scopes:       []string{"openid", "read_rewards_account_info"},
        Endpoint: oauth2.Endpoint{
            AuthURL:  "https://api-sandbox.capitalone.com/oauth/auz/authorize",
            TokenURL: "https://api-sandbox.capitalone.com/oauth/oauth20/token",
        },
    }
    // Some random string, random for each request
    oauthStateString = "random"
)

func handleLogin(w http.ResponseWriter, r *http.Request) {
    url := oAuthConfig.AuthCodeURL(oauthStateString)
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
    state := r.FormValue("state")
    if state != oauthStateString {
        fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
        http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
        return
    }

    code := r.FormValue("code")
    token, err := oAuthConfig.Exchange(oauth2.NoContext, code)
    if err != nil {
        fmt.Println("Code exchange failed with '%s'\n", err)
        http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
        return
    }

    client := &http.Client{}
    req, _ := http.NewRequest("GET", "https://api-sandbox.capitalone.com/rewards/accounts", nil)
    token.SetAuthHeader(req)
    response, _ := client.Do(req)

    defer response.Body.Close()
    contents, err := ioutil.ReadAll(response.Body)
    fmt.Fprintf(w, "Content: %s\n", contents)
}
