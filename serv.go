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
    Tags        string `json:"tags"`
    Time        int64 `json:"time"`
    Author      Author `json"Author"`
}

type User struct {
    UserName string
}

func main() {
    http.Handle("/css/", http.FileServer(http.Dir("template")))
    http.Handle("/js/", http.FileServer(http.Dir("template")))
    http.Handle("/", http.FileServer(http.Dir("template")))
    http.HandleFunc("/set", sethello)
    http.HandleFunc("/get", gethello)

    http.ListenAndServe(":8080", nil)
}

func gethello(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    if r.Method == "GET" {
        id := r.FormValue("id")
        start,err := strconv.Atoi(id)
        if err != nil {
            jstr := ""
        } else {
            jstr := getjson(start,1,2)
        }
        io.WriteString(w, jstr)
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
            Addr:     "localhost:6379",
            Password: "", // no password set
            DB:       0,  // use default DB
        })
        pong, err := client.Ping().Result()
        print(pong)
        if err != nil {
            fmt.Println("redis err:", err)
        } else {
            b, err := json.Marshal(art)
            if err != nil {
                fmt.Println("dbsize err:", err)
            } else {
                id, err := client.Incr("max:a:id").Result()
                client.Sadd("index:a:list",id)
                client.Set("index:a:sort:"+id, id)
                client.Set("a:"+id,b)
                io.WriteString(w, b)
            }
        }
    }
}

func getjson(start int, sign int, limit int ) string {
    var sarr []string
    var js string
    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })
    pong, err := client.Ping().Result()
    print(pong,err)

    for i := 0; i < limit; i++ {

        u := limit*sign - i + start
        d := strconv.Itoa(u)
        key := "a:"+string(d)

        exist, err := client.Exists(key).Result()

        if err != nil {
            fmt.Println("exist key:", key, "err" , err)
        } else if exist {
            val, err := client.Get(key).Result()
            if err != nil {
                fmt.Println("get key:", key, "err" , err)
            } else {
                sarr = append(sarr, val)
            }
        } else {
            fmt.Println("not exist key:", key, "err" , err)
        }
    }
    s := strings.Join(sarr,",")
    if s != ""  {
        js = "{\"articles\":[" + s + "]}"
    } else {
        js = "{\"error:1\"}"
    }
    print(js)
    return js
}

func redispagelist() {
    var sort redis.Sort
    getkey := []string{"a:*"}

    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       1,  // use default DB
    })
    pong, err := client.Ping().Result()
    print(pong,err)

    sort.By = "a:sort:*"
    sort.Offset = 0
    sort.Count = 1
    sort.Get = getkey
    sort.Order = "asc"

    rs,err := client.Sort("index:a:list",sort).Result()
    print(err)
    for i,v:=range rs {
        print(i,v)
    }
    fmt.Println(rs)  
}