package main

import (
    "io"
    "fmt"
    "time"
    "bytes"
    "net/url"
    "strings"
    "crypto/md5"
    "crypto/rand"
    "html/template"
    "database/sql"
  _ "github.com/go-sql-driver/mysql"
)


var (
    db  *sql.DB
)


const (
    DATATB string = "gsdata"
    LASTTB string = "gslast"
    USERTB string = "gsuser"
)


func initDB() {
    var err error
    //n := cfg.GetIntDefault("db_conn", 3)
    n := 3
    if n < 3 { n = 3 }
    host := "127.0.0.1:3306"
    name := cfg["db_name"]
    user := cfg["db_user"]
    pass := cfg["db_pass"]
    db, err = sql.Open("mysql",
                       user + ":" + pass + "@tcp(" + host + ")/" + name +
                       "?allowOldPasswords=1&parseTime=1&autocommit=true")
    if err != nil { panic(err) }
    db.SetMaxIdleConns(n)
    db.SetMaxOpenConns(n)
}


func getSummary(add bool) (t, v, a int, err error) {
    if add {
        db.Exec("update " + DATATB + " set visit = visit + 1 where postid=1")
    }
    err = db.QueryRow("select count(*) from " + DATATB + " union " +
                      "select visit from " + DATATB + " where postid=1 union " +
                      "select count(*) from " + DATATB +
                      " where (last + 2 * 60) > now()").Scan(&t, &v, &a)
    return
}


func getTopThreadId(l *list) (ts []int64, err error) {
    ts = make([]int64, 0, l.num + 1)
    var who string = "thread"
    if l.Last { who = "newest" }
    w := ""
    if l.first > 0 { w = fmt.Sprintf("where %s <= %d", who, l.first) }
    rows, err := db.Query(fmt.Sprintf("select thread, %s from %s %s " +
                                      "order by %s desc limit %d",
                                      who, LASTTB, w, who, l.num + 1))
    if err != nil { return }
    defer rows.Close()
    for rows.Next() {
        var i int64
        if err = rows.Scan(&i); err != nil { return }
        ts = append(ts, i)
    }

    if len(ts) > l.num {
        ts = ts[:l.num]
        l.Right = ts[l.num - 1]
    }
    if l.first <= 0 { return }
    // may need left
    rs, err := db.Query(fmt.Sprintf("select %s from %s where %s >= %d " +
                                    "order by %s limit %d",
                                    who, LASTTB, who, l.first, who, l.num))
    if err != nil { return }
    defer rs.Close()
    for rs.Next() {
        if err = rs.Scan(&l.Left); err != nil { return }
    }
    return
}


func fixPost(p *post, flag int, date *time.Time, tow *sql.NullString) {
    p.Del = flag > 0x7f
    p.Cool = make([]byte, flag & 0x1f)
    if date != nil {
        p.When = date.Format("2006/01/02-15:04:05")
    }
    if tow != nil {
        p.Tow = tow.String
    }
    ln := strings.ToLower(p.Afile)
    p.IsPic = strings.HasSuffix(ln, ".jpg") || strings.HasSuffix(ln, ".png") || strings.HasSuffix(ln, ".gif")
}


func getList(l *list) (err error) {
    m := make(map[int64]*post)
    m[0] = &post{Children:make([]*post, 0, l.num)}

    l.TotalNum, l.VisitCnt, l.ActiveCnt, err = getSummary(l.first == 0)
    if err != nil { return }

    ts, err := getTopThreadId(l)
    if err != nil { return }
    for _, t := range ts {
        var rows *sql.Rows
        var date time.Time
        var flag int
        var tow sql.NullString
        rows, err = db.Query("select postid, subject, parent, size, " +
                             "  visit, flag, date, user, u.name as name," +
                             "  only, t.name as tow, afile " +
                             "from " + DATATB + " d left join " + USERTB +
                             " u on d.user=u.id left join " + USERTB +
                             " t on d.only=t.id where thread=? " +
                             "order by postid", t)
        if err != nil { return }
        for rows.Next() {
            p := &post{Children:make([]*post, 0)}
            err = rows.Scan(&p.Cle, &p.Sub, &p.Parent, &p.Size,
                            &p.Visit, &flag, &date, &p.Uid, &p.Who,
                            &p.Tid, &tow, &p.Afile)
            if err != nil {
                // TODO log error
                continue
            }
            fixPost(p, flag, &date, &tow)
            if p.Tid == l.LoginId || p.Uid == l.LoginId { p.Tid = 0 }
            p.IsNew = p.Cle > l.lastcle
            m[p.Cle] = p
            pa := m[p.Parent]
            pa.Children = append(pa.Children, p)
        }
        rows.Close()
    }
    l.Posts = m[0].Children
    return nil
}


func procBody(src string) template.HTML {
    out := bytes.Buffer{}
    si := 0
    sp := false
    for i, r := range src {
        if r == ' ' {
            if sp {
                out.WriteString("&nbsp;")
                sp = false
            } else {
                out.WriteString(" ")
                sp = true
            }
            continue
        }
        sp = false
        switch {
        case i < si:
            continue
        case r == '\r':
            continue
        case r == '\n':
            out.WriteString("<br>\n")
        case r == '\t':
            out.WriteString("&nbsp; &nbsp; &nbsp; &nbsp;&nbsp;")
        //case unicode.IsSpace:
        case r == '<':
            out.WriteString("&lt;")
        case r == '>':
            out.WriteString("&gt;")
        case r == '&':
            out.WriteString("&amp;")
        case r == '\'':
            out.WriteString("&#39;")
        case r == '"':
            out.WriteString("&#34;")
        case r == 'h' && (strings.HasPrefix(src[i:], "http://") ||
             strings.HasPrefix(src[i:], "https://")):
            si = strings.IndexAny(src[i:], " ,，")
            if si < 0 || si > 500{
                out.WriteByte('h')
                si = 0
                continue
            }
            si += i
            out.WriteString("<a href=\"")
            out.WriteString(template.HTMLEscapeString(src[i:si]))
            out.WriteString("\">")
            out.WriteString(template.HTMLEscapeString(src[i:si]))
            out.WriteString("</a>")
        default:
            out.WriteRune(r)
        }
    }
    return template.HTML(out.String())
}


func dbGetPost(cle int64, uid int) (p *post, err error) {
    p = &post{Cle: cle, Uid: uid}
    flag := 0
    var tow sql.NullString
    err = db.QueryRow("select only, flag, user, name from " + DATATB +
                      " left join " + USERTB + " u on d.only=u.id " +
                      "where postid=?", cle).Scan(
                      &p.Tid, &flag, &p.Uid, &tow)
    if err != nil { return }
    if p.Tid != 0 {
        p.Tow = tow.String
        if uid != p.Tid && uid != p.Uid { return }
    }
    p.Del = flag > 0x7f
    if p.Del { return }

    var body string
    var date time.Time
    err = db.QueryRow("select subject, body, size, " +
                      "visit, flag, date, user, name, afile " +
                      "from " + DATATB + " d, " + USERTB + " u " +
                      "where d.user=u.id and postid=?", cle).Scan(
                       &p.Sub, &body, &p.Size, &p.Visit,
                       &flag, &date, &p.Uid, &tow, &p.Afile)
    if err != nil { return }
    p.Tid = 0   // because p.Tid == 0 or p.Tid == uid
    fixPost(p, flag, &date, &tow)
    p.Body = procBody(body)
    _, err = db.Exec("update " + DATATB +
                     " set visit = visit + 1 where postid=?", cle)
    return
}


func dbSavePost(po *post) (err error) {
    if po.Parent > 1 {
        err = db.QueryRow("select thread from " + DATATB +
                          " where postid=?", po.Parent).Scan(&po.Thread)
        if err != nil { return }
    }
    if po.Tow != "" {
        err = db.QueryRow("select id from " + USERTB +
                          " where name=?", po.Tow).Scan(&po.Tid)
        if err != nil { return }
    }
    res, err := db.Exec("insert into " + DATATB +
                        " (parent, thread, user, only, subject, size, " +
                        "  body, afile, visit)" +
                        " values (?, ?, ?, ?, ?, ?, ?, ?, 1)",
                        po.Parent, po.Thread, po.Uid, po.Tid,
                        po.Sub, po.Size, po.Body, po.Afile)
    if err != nil { return }

    po.Cle, err = res.LastInsertId()
    if err != nil { return }

    if po.Parent == 0 {
        po.Thread = po.Cle
        _, err = db.Exec("update " + DATATB +
                         " set thread = postid where postid=?", po.Cle)
        if err != nil { return }
    }
    _, err = db.Exec("insert into " + LASTTB +
                     " (thread, newest) values (?, ?) " +
                     "on duplicate key update newest="+
                     "if(values(newest) > newest, values(newest), newest)",
                     po.Thread, po.Cle)
    return
}


func dbSwapDel(cle int64) (err error) {
    _, err = db.Exec("update " + DATATB +
                     " set flag=if(flag>127, flag&0x7f, flag|0x80)" +
                     " where postid=?", cle)
    return
}


type User struct {
    uid int
    name string
    salt string
    pass string
    mail string
}


func (uo *User)genSalt() {
    h := md5.New()
    b := make([]byte, 10)
    _, _ = rand.Read(b)
    io.WriteString(h, string(b))
    io.WriteString(h, uo.name)
    io.WriteString(h, time.Now().String())
    uo.salt = fmt.Sprintf("%x", h.Sum(nil))
}


func (uo *User)calPwd(pwd string) (p string){
    h := md5.New()
    io.WriteString(h, uo.name)
    io.WriteString(h, pwd)
    io.WriteString(h, uo.salt)
    io.WriteString(h, uo.name)
    io.WriteString(h, pwd)
    io.WriteString(h, uo.salt)
    p = fmt.Sprintf("%x", h.Sum(nil))
    return
}


func queryUser(name string) (uobj *User, err error) {
    uobj = &User{}
    err = db.QueryRow("select id, name, salt, pwd, mail from " + USERTB +
                      " where name=?", name).Scan(&uobj.uid, &uobj.name,
                      &uobj.salt, &uobj.pass, &uobj.mail)
    if err != nil { return }
    if uobj.salt == "" {
        // conver old data
        uobj.genSalt()
        p, _ := url.QueryUnescape(uobj.pass)
        uobj.pass = uobj.calPwd(p)
        err = uobj.updatePSE()
        if err != nil { return }
    }
    return
}


func (uobj *User)updatePSE() (err error) {
    _, err = db.Exec("update " + USERTB +
                     " set pwd=?, salt=?, mail=? where id=?",
                     uobj.pass, uobj.salt, uobj.mail, uobj.uid)
    return
}


func (uobj *User)SaveNew() (err error) {
    _, err = db.Exec("insert into " + USERTB +
                     " (name, pwd, salt, mail) values (?, ?, ?, ?)",
                     uobj.name, uobj.pass, uobj.salt, uobj.mail)
    return
}


func UpdateLast(uid int) {
    db.Exec("update " + USERTB + " set last=now() where id=?", uid)
}
