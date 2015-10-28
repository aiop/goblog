package main

import (
    "net/http"
    "gopkg.in/redis.v3"
    "encoding/json"
    "io"
    "time"
    "fmt"
    "strings"
    "strconv"
)

type Author struct {
    Name string `json:"name"`
    Url  string `json:"url"`
    Img  string `json:"img"`
}

type Article struct {
    Id          string `json:"id"`
    Title       string `json:"title"`
    Link        string `json:"link"`
    Description string `json:"description"`
    Tags        []string `json:"tags"`
    Time        int64 `json:"time"`
    Author      Author `json"Author"`
}

const Port = "8080"
const RedisAddr = "localhost:6379"
const RedisPassword = ""
const RedisDb = 0
const PreNum = 10
const BaseNum = 1000000000
const BegainNm = 1000000001

var client *redis.Client

func main() {
    client = redis.NewClient(&redis.Options{
            Addr:     RedisAddr,
            Password: RedisPassword,
            DB:       RedisDb,
    })
    pong, err := client.Ping().Result()

    http.Handle("/css/", http.FileServer(http.Dir("template")))
    http.Handle("/js/", http.FileServer(http.Dir("template")))
    http.Handle("/img/", http.FileServer(http.Dir("template")))
    http.Handle("/", http.FileServer(http.Dir("template")))

    if err != nil {
        fmt.Println(pong, err)
        http.HandleFunc("/set", showerror)
        http.HandleFunc("/get", showerror)
        http.HandleFunc("/tag", showerror)
    } else {
        http.HandleFunc("/set", sethello)
        http.HandleFunc("/get", gethello)
        http.HandleFunc("/tag", taghello)
    }
    http.ListenAndServe(":"+Port, nil)
}

func showerror(w http.ResponseWriter, r *http.Request) {
    jsonString:="something is wrong"
    io.WriteString(w, jsonString)
}

func sethello(w http.ResponseWriter, r *http.Request) {
    var art Article
    var strid string
    r.ParseForm()
    if r.Method == "POST" {
        art.Id           = r.PostFormValue("id")
        art.Title        = r.PostFormValue("title")
        art.Link         = r.PostFormValue("link")
        art.Description  = r.PostFormValue("description")
        art.Author.Name  = r.PostFormValue("author[name]")
        art.Author.Url   = r.PostFormValue("author[url]")
        art.Author.Img   = r.PostFormValue("author[img]")
        art.Time         = time.Now().Unix()

        _tags           := r.PostFormValue("tags")
        tags            := strings.Split(_tags,",")
        for _, t := range tags {
            t = strings.Trim(t," ")
            if t != "" {
                art.Tags = append(art.Tags, t)
            }
        }

        if client != nil {
            if art.Id == "" {
                id, err := client.Incr("max:a:id").Result()
                strid = strconv.Itoa(int(id))
                if err != nil {
                    fmt.Println("Incr err:", err)
                } else {
                    art.Id = strid
                    str, err := json.Marshal(art)
                    if err != nil {
                        fmt.Println("json.Marshal err:", err)
                    } else {
                        akey := "a:" + strid
                        client.LPush("list:index", akey)
                        for _, t := range tags {
                            t = strings.Trim(t," ")
                            if t != "" {
                                kt := str2utf(t)
                                client.LPush("list:tag:"+kt, akey)
                                art.Tags = append(art.Tags, t)
                            }
                        }
                        client.Set(akey, str, 0)
                        io.WriteString(w, strid)
                    }
                }
            } else {
                var p Article
                oart, err := client.Get("a:" + art.Id).Result()
                if err != nil {
                    fmt.Println(oart, err)
                } else {
                    json.Unmarshal([]byte(oart), &p)
                    art.Time = p.Time
                    str, err := json.Marshal(art)
                    if err != nil {
                        fmt.Println("json.Marshal err:", err)
                    } else {
                        client.Set("a:" + art.Id, str, 0)
                        io.WriteString(w, art.Id)
                    }
                }
            }
        }
    }
}

func  taghello(w http.ResponseWriter, r *http.Request) {
    var jsonString, _jsonString string
    r.ParseForm()
    if r.Method == "GET" {
        client.Incr("visit")
        client.Incr("visit:tag")
        name := r.FormValue("name")
        p := r.FormValue("p")

        page, err := strconv.Atoi(p)
        prenum := PreNum
        _jsonString = ""
        if err != nil {
            _jsonString = ""
        } else {

            if page <= 1 {
                page = 1
            }
            page = page - 1
            start := page * prenum
            stop := start + prenum

            if prenum < 1 {
                _jsonString = ""
            } else {
                name = str2utf(name)
                getArr := getAlist("list:tag:"+name, int64(start), int64(stop))
                if getArr != nil {
                    _jsonString = strings.Join(getArr, ",")
                } else {
                    _jsonString = ""
                }
            }

        }

        jsonString = "{\"articles\":[" + _jsonString + "]}"
        io.WriteString(w, jsonString)
    }
}

func gethello(w http.ResponseWriter, r *http.Request) {
    var jsonString, _jsonString string
    r.ParseForm()
    if r.Method == "GET" {
        client.Incr("visit")
        client.Incr("visit:get")
        p := r.FormValue("p")

        page, err := strconv.Atoi(p)
        prenum := PreNum
        _jsonString = ""
        if err != nil {
            _jsonString = ""
        } else {
            if page <= 1 {
                page = 1
            }
            page = page - 1
            start := page * prenum
            stop := start + prenum

            getArr := getAlist("list:index", int64(start), int64(stop))

            if getArr != nil {
                _jsonString = strings.Join(getArr, ",")
            } else {
                _jsonString = ""
            }

        }

        jsonString = "{\"articles\":[" + _jsonString + "]}"
        io.WriteString(w, jsonString)
    }
}

func getAlist(ListKey string, start int64, stop int64) []string {

    var Arr []string

    getArr, err := client.LRange(ListKey, start, stop).Result()
    client.Incr("visit:list")
    if err != nil {
        fmt.Println(getArr, err)
    } else {
        for _, k := range getArr {
            s := getA(k)
            Arr = append(Arr, s)
        }
    }

    return Arr
}

func getA(akey string) string {

    Art, err := client.Get(akey).Result()
    client.Incr("visit:a")
    if err != nil {
        fmt.Println(Art, err)
    }
    return Art
}

func getMaxid() int {
    var id int
    Maxid, err := client.Get("max:a:id").Result()
    if err != nil {
        fmt.Println(Maxid, err)
    }

    id, err = strconv.Atoi(Maxid)
    if err != nil {
        fmt.Println(id, err)
    }

    return id
}

func str2utf(str string) string {
    rs := []rune(str)
    json := ""
    for _, r := range rs {
        rint := int(r)
        if rint < 128 {
            json += string(r)
        } else {
            json += "u"+strconv.FormatInt(int64(rint), 16) // json
        }
    }
    return json
}
