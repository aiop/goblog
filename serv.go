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
    "sort"
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
const PreNum = 6
const BaseNum = 1000000000
const BegainNm = 1000000001

func main() {
    http.Handle("/css/", http.FileServer(http.Dir("template")))
    http.Handle("/js/", http.FileServer(http.Dir("template")))
    http.Handle("/img/", http.FileServer(http.Dir("template")))
    http.Handle("/", http.FileServer(http.Dir("template")))
    http.HandleFunc("/set", sethello)
    http.HandleFunc("/get", gethello)

    http.ListenAndServe(":"+Port, nil)
}

func gethello(w http.ResponseWriter, r *http.Request) {
    var jsonString, _jsonString string
    r.ParseForm()
    if r.Method == "GET" {
        id := r.FormValue("id")
        order := r.FormValue("order")

        start, err := strconv.Atoi(id)
        prenum := PreNum
        _jsonString = ""
        fmt.Println(start)
        if err != nil {
            _jsonString = ""
        } else {

            if start == 0 && order == "up" {
                maxid := getMaxid()
                start = maxid - PreNum
            }

            if order == "down" {
                start = start - PreNum - 1
            }

            if start < BegainNm {
                fmt.Println(start)
                prenum = PreNum + (start - BegainNm) + 1
                start = BegainNm - 1
                fmt.Println("prenum:",prenum)
            }




            if prenum < 1 {
                _jsonString = ""
            } else {
                fmt.Println(start)

                getArr := getAlist("index:a:list", float64(start-BaseNum), float64(prenum))
                fmt.Println("start:", start)
                if getArr != nil {
                    sort.Sort(sort.Reverse(sort.StringSlice(getArr)))
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

func sethello(w http.ResponseWriter, r *http.Request) {
    var art Article
    var strid string
    r.ParseForm()
    if r.Method == "POST" {
        art.Title        = r.PostFormValue("title")
        art.Link         = r.PostFormValue("link")
        art.Description  = r.PostFormValue("description")
        art.Author.Name  = r.PostFormValue("author[name]")
        art.Author.Url   = r.PostFormValue("author[url]")
        art.Author.Img   = r.PostFormValue("author[img]")
        art.Time         = time.Now().Unix()

        _tags           := r.PostFormValue("tags")
        tags            := strings.Split(_tags,",")
        art.Tags         = tags

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
            strid = strconv.Itoa(int(id))
            if err != nil {
                fmt.Println("Incr err:", err)
            } else {
                art.Id = strid
                str, err := json.Marshal(art)
                if err != nil {
                    fmt.Println("json.Marshal err:", err)
                } else {
                    client.SAdd("index:a:list", strid)
                    client.Set("index:a:sort:" + strid, strid, 0)
                    client.Set("a:" + strid, str, 0)
                    io.WriteString(w, strid)
                }
            }
        }
    }
}

func getAlist(ListKey string, Offset float64, Count float64) []string {
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
        sort.Offset = Offset
        sort.Count = Count
        sort.Get = getkey
        sort.Order = "ASC"

        getArr, err := client.Sort(ListKey, sort).Result()

        if err != nil {
            fmt.Println(getArr, err)
        } else {
            return getArr
        }
    }
    return nil
}

func getMaxid() int {
    client := redis.NewClient(&redis.Options{
        Addr:     RedisAddr,
        Password: RedisPassword,
        DB:       RedisDb,
    })
    pong, err := client.Ping().Result()

    if err != nil {
        fmt.Println(pong, err)
    } else {

        Maxid, err := client.Get("max:a:id").Result()
        if err != nil {
            fmt.Println(Maxid, err)
        } else {
            id, err := strconv.Atoi(Maxid)
            if err != nil {
                fmt.Println(id, err)
            } else {
                return id
            }
        }
    }
    return 0
}
