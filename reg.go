package main

import (
    "fmt"
    "net/http"
    "github.com/pastebt/sess"
)


func update(name, pwd, np1, np2, mail string) (msg string) {
    if np1 != np2 { return "err: password not match" }

    uo, err := queryUser(name)
    if err != nil {
        logging.Error(err)
        return
    }
    if uo.calPwd(pwd) != uo.pass { return "err: wrong password" }
    // use new salt
    uo.genSalt()
    uo.pass = uo.calPwd(np1)
    if mail != "" { uo.mail = mail }
    err = uo.updatePSE()
    if err != nil {
        logging.Error(err)
        return "err: failed update"
    }
    return "Updated"
}


func reg(name, np1, np2, mail string) (msg string) {
    if np1 != np2 { return "err: password not match" }

    uo := &User{name:name, mail:mail}
    uo.genSalt()
    uo.pass = uo.calPwd(np1)
    err := uo.SaveNew()
    if err != nil {
        logging.Error(err)
        return "err: failed"
    }
    return "New user " + name
}


func hReg(w http.ResponseWriter, r *http.Request) {
    s := sess.Start(w, r)
    l := r.FormValue("logout")
    if l == "1" {
        s.Set("loginuserid", "")
        s.Set("loginusername", "")
        http.Redirect(w, r, "/phh/post.phh", http.StatusFound) //302) never use 301
        return
    }
    un := r.FormValue("username")
    pw := r.FormValue("password")
    p1 := r.FormValue("newpwd1")
    p2 := r.FormValue("newpwd2")
    el := r.FormValue("email")
    msg := ""
    if un != "" && p1 != "" && p2 != "" {
        if pw == "" {
            msg = reg(un, p1, p2, el)
        } else {
            msg = update(un, pw, p1, p2, el)
        }
    }
    tpl.ExecuteTemplate(w, "head.tpl", "a")
    if msg == "" {
        tpl.ExecuteTemplate(w, "reg.tpl", "")
    } else {
        fmt.Fprintf(w, msg)
    }
    tpl.ExecuteTemplate(w, "foot.tpl", "a")
}
