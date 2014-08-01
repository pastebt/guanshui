package main

import (
    "os"
    "io"
    "fmt"
    "path"
    "strconv"
    "net/http"
    "unicode/utf8"
    "html/template"
    "github.com/pastebt/sess"
)


type post struct {
    Cle int64
    Parent int64
    Thread int64
    Sub string
    Who string
    Uid int
    Tow string  // only to who
    Tid int     // only to user id, if it equal loginid, will reset to 0
    When string
    Visit int
    Cool []byte
    Size int64
    Del bool
    Afile string
    IsPic bool
    IsNew bool
    Body template.HTML
    Children []*post
    LoginId int
    LoginName string
}


func logIn(w http.ResponseWriter, r *http.Request) bool {
    if r.PostForm == nil {
        e := r.ParseMultipartForm(32 << 20)
        if e != nil {
            logging.Error(e)
        }
    }

    u := r.FormValue("username")
    p := r.FormValue("password")
    if u == "" || p == "" { return false }
    logging.Debug("u=", u, ", p=", p, ", form=", r.PostForm)

    db, err := connect()
    if err != nil {
        logging.Error(err)
        return false
    }
    defer db.Close()

    uobj, err := queryUser(db, u)
    if err != nil {
        logging.Error(err)
        return false
    }
    np := uobj.calPwd(p)
    if np != uobj.pass {
        postMsg(w, "Who are you?")
        return true
    }
    s := sess.Start(w, r)
    s.Set("loginuserid", fmt.Sprint(uobj.uid))
    s.Set("loginusername", fmt.Sprint(uobj.name))
    postMsg(w, fmt.Sprintf("Welcome back %s", uobj.name))
    return true
}


func savePost(w http.ResponseWriter, r *http.Request) {
    if logIn(w, r) { return }
    uid, name := auth(w, r)
    if uid == 0 { return }
    db, err := connect()
    if err != nil {
        logging.Error("user ", name, " post: ", err)
        return
    }
    defer db.Close()
    delcle := r.FormValue("delcle")
    if delcle != "" {
        cle, err := strconv.ParseInt(delcle, 10, 64)
        if err != nil {
            logging.Error(err)
            return
        }
        dbSwapDel(db, cle)
        postMsg(w, "Done")
        return
    }
    sub := r.FormValue("subject")
    if sub == "" {
        postMsg(w, "Missed subject?")
        return
    }
    if len(sub) > (1 << 10) {
        postMsg(w, "Subject too long (> 1024)")
        return
    }
    p, err := strconv.ParseInt(r.FormValue("parent"), 10, 64)
    if err != nil {
        logging.Error("Can not process parent:", err)
        return
    }
    b := r.FormValue("body")
    if int64(len(b)) > limit_body {
        postMsg(w, fmt.Sprintf("Body too long (> %d)", limit_body))
        return
    }
    fn := ""
    fobj, fh, err := r.FormFile("attachment")
    if err == nil {
        fn = fh.Filename
        size, err := fobj.Seek(0, os.SEEK_END)
        if err != nil {
            logging.Error(err)
            return
        }
        if size > limit_file {
            postMsg(w, fmt.Sprintf("File too long (> %d)", limit_file))
            return
        }
        fobj.Seek(0, os.SEEK_SET)
    }
    po := &post{Uid: uid,
                Parent: p,
                Sub: sub,
                Tow: r.FormValue("onlyto"),
                Size: int64(utf8.RuneCountInString(b)),
                Body: template.HTML(b),
                Afile: fn}
    err = dbSavePost(db, po)
    if err != nil {
        logging.Error("user ", name, " save post: ", err)
        postMsg(w, "Sorry, something wrong")
        return
    }
    if len(fn) > 0 {
        ln := path.Join(cfg["htdoc"], cfg["path_attach"],
                        fmt.Sprintf("/atta_%09d", po.Cle))
        fout, err := os.OpenFile(ln, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
        if err != nil {
            logging.Error(err)
            return
        }
        defer fout.Close()
        io.Copy(fout, fobj)
    }
    if po.Cle > getLastCle(w, r) {
        sess.Start(w, r).Set("lastcle", fmt.Sprint(po.Cle))
    }
    postMsg(w, "Thank You")
}


func postMsg(w http.ResponseWriter, msg string) {
    tpl.ExecuteTemplate(w, "head.tpl", "")
    fmt.Fprint(w, msg)
    tpl.ExecuteTemplate(w, "foot.tpl", "")
}


func hPost(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        savePost(w, r)
        return
    }
    uid, name := auth(w, r)
    logging.Debug("hPost, ", uid, ", ", name)
    tpl.ExecuteTemplate(w, "head.tpl", "")

    db, err := connect()
    if err != nil {
        logging.Error(err)
    } else {
        defer db.Close()
        var cle int64 = 0
        s := r.FormValue("cle")
        if len(s) > 1 {
            cle, err = strconv.ParseInt(s, 10, 64)
            if err != nil {
                logging.Error(err)
                cle = 0
            }
        }
        var p *post
        if cle > 0 {
            p, err = dbGetPost(db, cle, uid)
            if err != nil {
                logging.Error(err)
            }
        }
        if p == nil { p = &post{} }
        if p.Cle > getLastCle(w, r) {
            sess.Start(w, r).Set("lastcle", fmt.Sprint(p.Cle))
        }
        p.LoginId = uid
        p.LoginName = name
        tpl.ExecuteTemplate(w, "post.tpl", p)
    }
    tpl.ExecuteTemplate(w, "foot.tpl", "")
}
