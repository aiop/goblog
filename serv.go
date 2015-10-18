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

const Port = "8080"
const RedisAddr = "localhost:6379"
const RedisPassword = ""
const RedisDb = 0
const PreNum = 3

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
    Tags        string `json:"tags"`
    Time        int64 `json:"time"`
    Author      Author `json"Author"`
}

func main() {
    http.Handle("/css/", http.FileServer(http.Dir("template")))
    http.Handle("/js/", http.FileServer(http.Dir("template")))
    http.Handle("/", http.FileServer(http.Dir("template")))
    http.HandleFunc("/set", sethello)
    http.HandleFunc("/get", gethello)

    http.ListenAndServe(":"+Port, nil)
}

func gethello(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    if r.Method == "GET" {
        id := r.FormValue("id")
        start, err := strconv.Atoi(id)
        if err != nil {
            jsonString := ""
        } else {
            getArr := getAlist("index:a:list", start, PreNum)
            if getArr != nil {
                jsonString := strings.Join(getArr, ",")
            } else {
                jsonString := ""
            }
        }
        jsonString = "{\"articles\":[" + jsonString + "]}"
        io.WriteString(w, jsonString)
    }
}

func sethello(w http.ResponseWriter, r *http.Request) {
    var art Article

    r.ParseForm()

    if r.Method == "POST" {

        art.Title        = r.PostFormValue("title")
        art.Link         = r.PostFormValue("link")
        art.Description  = r.PostFormValue("description")
        art.Tags         = r.PostFormValue("tags")
        art.Author.Name  = r.PostFormValue("author[name]")
        art.Author.Url   = r.PostFormValue("author[url]")
        art.Author.Img   = r.PostFormValue("author[img]")
        art.Time         = time.Now().Unix()

        client := redis.NewClient(&redis.Options{
            Addr:     RedisAddr,
            Password: RedisPassword,
            DB:       RedisDb,
        })

        pong, err := client.Ping().Result()

        if err != nil {
            fmt.Println("Ping Redis err:", pong, err)
        } else {
            id, err := client.Incr("max:a:id").Result()
            if err != nil {
                fmt.Println("Incr err:", err)
            } else {
                art.Id = id
                str, err := json.Marshal(art)
                if err != nil {
                    fmt.Println("json.Marshal err:", err)
                } else {
                    client.Sadd("index:a:list", id)
                    client.Set("index:a:sort:" + id, id)
                    client.Set("a:" + id, str)
                    io.WriteString(w, str)
                }
            }
        }
    }
}

func getAlist(ListKey string, Offset int, Count int) []string {
    var sort redis.Sort
    getkey := []string{"a:*"}

    client := redis.NewClient(&redis.Options{
        Addr:     RedisAddr,
        Password: RedisPassword,
        DB:       RedisDb,
    })
    pong, err := client.Ping().Result()

    if err != nil {
        fmt.Println(pong, err)
    } else {
        sort.By = "a:sort:*"
        sort.Offset = 0
        sort.Count = 1
        sort.Get = getkey
        sort.Order = "DESC"

        getArr, err := client.Sort(ListKey, sort).Result()

        if err != nil {
            fmt.Println(getArr, err)
        } else {
            return getArr
        }
    }
    return nil
}