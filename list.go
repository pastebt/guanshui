package main

import (
    "os"
    "fmt"
    "path"
    "strconv"
    "net/http"
)


type list struct {
    LoginId  int
    TotalNum int
    VisitCnt int
    ActiveCnt int
    Left  int64
    Right int64
    Last  bool
    Posts []*post
    num   int
    first int64
    lastcle int64
}


func hList(w http.ResponseWriter, r *http.Request) {
    i, _ := auth(w, r)
    l := list{LoginId:i}
    /*
    db, err := connect()
    if err != nil {
        logging.Error(err)
        return
    }
    defer db.Close()
    */
    l.Last = r.FormValue("new") == "1"
    f, err := strconv.ParseInt(r.FormValue("first"), 10, 64)
    if err == nil {
        l.first = f
    }
    l.num = 20
    ln, ok := cfg["last_num"]
    if ok {
        i, e := strconv.ParseInt(ln, 10, 32)
        if e == nil { l.num = int(i) }
    }

    l.lastcle = getLastCle(w, r)

    if err = getList(&l); err != nil {
        logging.Error(err)
    }

    tpl.ExecuteTemplate(w, "head.tpl", "")
    if err = tpl.ExecuteTemplate(w, "list.tpl", l); err != nil {
        logging.Error(err)
    }
    tpl.ExecuteTemplate(w, "foot.tpl", "")
}


func hAttach(w http.ResponseWriter, r *http.Request) {
    i, _ := auth(w, r)
    if i < 1 {
        http.NotFound(w, r)
        return
    }
    var cle int64
    s := r.FormValue("cle")
    if len(s) > 1 {
        i, err := strconv.ParseInt(s, 10, 64)
        if err != nil {
            logging.Error(err)
        } else {
            cle = i
        }
    }
    if cle == 0 {
        http.NotFound(w, r)
        return
    }
    fn := path.Join(cfg["htdoc"], cfg["path_attach"], fmt.Sprintf("/atta_%09d", cle))
    f, err := os.Open(fn)
    if err != nil {
        logging.Error(err)
        http.NotFound(w, r)
        return
    }
    defer f.Close()
    d, err := f.Stat()
    if err != nil || d.IsDir() {
        logging.Error(err)
        http.NotFound(w, r)
        return
    }
    http.ServeContent(w, r, d.Name(), d.ModTime(), f)
}
