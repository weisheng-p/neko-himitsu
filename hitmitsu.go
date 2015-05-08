package hitmitsu

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "time"
    "strings"
    "strconv"

    "appengine"
    "appengine/datastore"
    "appengine/urlfetch"
)

type Hitmitsu struct {
    Content string
    Date    time.Time
    Gold    int
    Silver  int
    TheirDate string
}

func init() {
    http.HandleFunc("/", handler)
    http.HandleFunc("/passupdate", update)

}
func getPassword(c appengine.Context) (*Hitmitsu,error){
    client := urlfetch.Client(c)
    resp, err := client.Get("http://hpmobile.jp/app/nekoatsume/neko_daily.php")
    if err != nil {
        return nil, err;
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err;
    }
    s := strings.Split(string(body), ",")
    password := s[1]
    silver,err := strconv.Atoi(s[2])
    gold,err := strconv.Atoi(s[3])
    theirDate := s[4]
    h := Hitmitsu{
        Date: time.Now(),
        Content: password,
        Silver: silver,
        Gold: gold,
        TheirDate: theirDate,
    }
    return &h, nil
}

func getOurPassword(c appengine.Context) (*Hitmitsu,error){
    q := datastore.NewQuery("Hitmitsu").
            Order("-Date").
            Limit(1)
    var hitmitsu []Hitmitsu
    _, err := q.GetAll(c, &hitmitsu)
    if err != nil {
        return nil,err
    }
    if len(hitmitsu) == 0 {
        return nil, nil
    }
    return &hitmitsu[0],nil
}

func update(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    var ourPassword *Hitmitsu

    ourPassword, err := getOurPassword(c)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return;
    }
    if ourPassword != nil {
        lastUpdate := int64(time.Now().Sub(ourPassword.Date) / time.Hour)
        if lastUpdate < 20 {
            fmt.Fprint(w, "too early")
            return
        }
    }

    password, err := getPassword(c)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return;
    }

    if ourPassword != nil &&  ourPassword.Content == password.Content {
        fmt.Fprint(w, "same")
        return;
    }

    key := datastore.NewKey(
        c,
        "Hitmitsu",
        "",
        0,
        nil,
    )
    _, err = datastore.Put(c, key, password)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return;
    }
    fmt.Fprint(w, "ok, updated")
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello, world!")
}
