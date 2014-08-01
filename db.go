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
    "github.com/ziutek/mymysql/mysql"
    _ "github.com/ziutek/mymysql/thrsafe"
)


func connect() (db mysql.Conn, err error) {
    db = mysql.New("tcp", "", "127.0.0.1:3306",
                   cfg["db_user"], cfg["db_pass"], cfg["db_name"])
    db.Register("set names 'utf8'")
    err = db.Connect()
    if err != nil {
        panic(err)
    }
}


const (
    DATATB string = "gsdata"
    LASTTB string = "gslast"
    USERTB string = "gsuser"
)


func getSummary(db mysql.Conn, add bool) (t, v, a int, err error) {
    if add {
        db.Query("update %s set visit = visit + 1 where postid=1", DATATB)
    }
    rows, _, err := db.Query("select count(*) from %s union " +
                             "select visit from %s where postid=1 union " +
                             "select count(*) from %s where (last + 2 * 60) > now()",
                             DATATB, DATATB, USERTB)
    if err != nil { return }
    return rows[0].Int(0), rows[1].Int(0), rows[2].Int(0), nil
}


func getTopThreadId(db mysql.Conn, l *list) (ts []int64, err error) {
    ts = make([]int64, 0, l.num + 1)
    var who string = "thread"
    if l.Last { who = "newest" }
    var rows []mysql.Row
    w := ""
    if l.first > 0 { w = fmt.Sprintf("where %s <= %d", who, l.first) }
    rows, _, err = db.Query("select thread, %s from %s %s " +
                            "order by %s desc limit %d",
                            who, LASTTB, w, who, l.num + 1)
    if err != nil { return }

    if len(rows) > l.num {
        rows = rows[:l.num]
        l.Right = rows[l.num - 1].Int64(1)
    }
    for _, row := range rows {
        ts = append(ts, row.Int64(0))
    }
    if l.first > 0 {
        // may need left
        rows, _, err = db.Query("select %s from %s where %s >= %d " +
                                "order by %s limit %d",
                                who, LASTTB, who, l.first, who, l.num)
        if err == nil && len(rows) > 1 {
            l.Left = rows[len(rows) - 1].Int64(0)
        }
    }
    return
}


func fillPost(p *post, row *mysql.Row, res mysql.Result) (err error) {
    flag := row.Int(res.Map("flag"))
    p.Cle = row.Int64(res.Map("postid"))
    p.Sub = row.Str(res.Map("subject"))
    p.Who = row.Str(res.Map("name"))
    p.Uid = row.Int(res.Map("user"))
    p.Size = row.Int64(res.Map("size"))
    p.Del = flag > 0x7f
    p.Cool = make([]byte, flag & 0x1f)
    p.Visit = row.Int(res.Map("visit")) - 1
    p.When = row.Localtime(res.Map("date")).Format("2006/01/02-15:04:05")
    p.Afile = row.Str(res.Map("afile"))
    ln := strings.ToLower(p.Afile)
    p.IsPic = strings.HasSuffix(ln, ".jpg") || strings.HasSuffix(ln, ".png") || strings.HasSuffix(ln, ".gif")
    return
}


func getList(db mysql.Conn, l *list) (err error) {
    m := make(map[int64]*post)
    m[0] = &post{Children:make([]*post, 0, l.num)}

    l.TotalNum, l.VisitCnt, l.ActiveCnt, err = getSummary(db, l.first == 0)
    if err != nil { return }

    ts, err := getTopThreadId(db, l)
    if err != nil { return }
    for _, t := range ts {
        rows, res, e := db.Query("select postid, subject, parent, size, " +
                                 "  visit, flag, date, user, u.name as name," +
                                 "  only, t.name as tow, afile " +
                                 "from %s d left join %s u on d.user=u.id " +
                                 "  left join  %s t on d.only=t.id " +
                                 "where thread=%d order by postid",
                                 DATATB, USERTB, USERTB, t)
        if e != nil { return e }
        for _, row := range rows {
            p := &post{Children:make([]*post, 0)}
            fillPost(p, &row, res)
            p.Uid = row.Int(res.Map("user"))
            p.Tid = row.Int(res.Map("only"))
            if p.Tid > 0 { p.Tow = row.Str(res.Map("tow")) }
            if p.Tid == l.LoginId || p.Uid == l.LoginId { p.Tid = 0 }
            p.IsNew = p.Cle > l.lastcle
            m[p.Cle] = p
            p.Parent = row.Int64(res.Map("parent"))
            pa := m[p.Parent]
            pa.Children = append(pa.Children, p)
        }
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


func dbGetPost(db mysql.Conn, cle int64, uid int) (p *post, err error) {
    row, _, err := db.QueryFirst("select only, name, flag, user " +
                                 "from %s d left join %s u on d.only=u.id " +
                                 "where postid=%d", DATATB, USERTB, cle)
    if err != nil { return }
    p = &post{Cle: cle, Uid: uid}
    p.Tid = row.Int(0)
    if p.Tid != 0 {
        logging.Debug("Tid=", p.Tid)
        p.Tow = row.Str(1)
        p.Uid = row.Int(3)
        if uid != p.Tid && uid != p.Uid { return }
    }
    p.Del = row.Int(2) > 0x7f
    if p.Del { return }
    row, res, err := db.QueryFirst("select postid, subject, body, size, " +
                                   "visit, flag, date, user, name, afile " +
                                   "from %s d, %s u where d.user=u.id and postid=%d",
                                   DATATB, USERTB, cle)
    if err != nil { return }
    p.Tid = 0   // because p.Tid == 0 or p.Tid == uid
    fillPost(p, &row, res)
    p.Body = procBody(row.Str(res.Map("body")))
    _, _, err = db.Query("update %s set visit = visit + 1 where postid=%d", DATATB, cle)
    return
}


func dbSavePost(db mysql.Conn, po *post) (err error) {
    if po.Parent > 1 {
        row, _, e := db.QueryFirst("select thread from %s where postid=%d",
                                    DATATB, po.Parent)
        if e != nil { return e }
        po.Thread = row.Int64(0)
    }
    if po.Tow != "" {
        st, e := db.Prepare("select id from " + USERTB + " where name=?")
        if e != nil { return e }
        row, _, e := st.ExecFirst(po.Tow)
        if e != nil { return e }
        po.Tid = row.Int(0)
    }
    stmt, err := db.Prepare("insert into " + DATATB +
                            "(parent, thread, user, only, subject, size, " +
                            " body, afile, visit)" +
                            " values (?, ?, ?, ?, ?, ?, ?, ?, 1)")
    if err != nil { return }
    _, res, err := stmt.Exec(po.Parent, po.Thread, po.Uid, po.Tid,
                             po.Sub, po.Size, po.Body, po.Afile)
    if err != nil { return }
    po.Cle = int64(res.InsertId())
    if po.Parent == 0 {
        po.Thread = po.Cle
        _, _, err = db.Query("update %s set thread = postid where postid=%d",
                              DATATB, po.Cle)
    }
    _, _, err = db.Query("insert into %s (thread, newest) values (%d, %d) " +
                         "on duplicate key update newest="+
                         "if(values(newest) > newest, values(newest), newest)",
                         LASTTB, po.Thread, po.Cle)
    return
}


func dbSwapDel(db mysql.Conn, cle int64) (err error) {
    _, _, err = db.Query("update %s set flag=if(flag>127, flag&0x7f, flag|0x80)" +
                         " where postid=%d", DATATB, cle)
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


func queryUser(db mysql.Conn, name string) (uobj *User, err error) {
    stmt, err := db.Prepare(fmt.Sprintf("select * from %s where name=?", USERTB))
    if err != nil { return }
    row, res, err := stmt.ExecFirst(name)
    if err != nil { return }
    uobj = &User{name:name}
    if len(row) == 0 { return }
    uobj.uid = row.Int(res.Map("id"))
    uobj.name = row.Str(res.Map("name"))
    uobj.salt = row.Str(res.Map("salt"))
    uobj.pass = row.Str(res.Map("pwd"))
    uobj.mail = row.Str(res.Map("mail"))
    if uobj.salt == "" {
        // conver old data
        uobj.genSalt()
        p, _ := url.QueryUnescape(uobj.pass)
        uobj.pass = uobj.calPwd(p)
        err = uobj.updatePSE(db)
        if err != nil { return }
    }
    return
}


func (uobj *User)updatePSE(db mysql.Conn) (err error) {
    stmt, err := db.Prepare(fmt.Sprintf("update %s set pwd=?, salt=?, mail=? where id=?", USERTB))
    if err != nil { return }
    _, _, err = stmt.Exec(uobj.pass, uobj.salt, uobj.mail, uobj.uid)
    return
}


func (uobj *User)SaveNew(db mysql.Conn) (err error) {
    stmt, err := db.Prepare(fmt.Sprintf("insert into %s (name, pwd, salt, mail) values (?, ?, ?, ?)", USERTB))
    if err != nil { return }
    _, _, err = stmt.Exec(uobj.name, uobj.pass, uobj.salt, uobj.mail)
    return
}


func UpdateLast(uid int) {
    db, err := connect()
    if err != nil { return }
    _, _, _ = db.Query("update %s set last=now() where id=%d", USERTB, uid)
}
