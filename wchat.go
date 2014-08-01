package main

import (
    "log"
    "path"
    "strconv"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "html/template"
    "github.com/pastebt/sess"
    "github.com/pastebt/gslog"
)


const (
    basedir string = "/opt/guanshui"
    sessid string  = "GuanShuiSessionID"
)


var tpl *template.Template
var cfg = make(map[string]string)
var limit_body, limit_file, limit_all int64


func cfgInt64(name string, dft int64) (ret int64) {
    ret = dft
    v, ok := cfg[name]
    if ! ok { return }
    ret, err := strconv.ParseInt(v, 10, 64)
    if err != nil { return dft }
    return
}


func initCfg(cfgfn string){
    // load configure
    data, err := ioutil.ReadFile(cfgfn)
    if err != nil {
        log.Fatal(err)
    }
    if err = json.Unmarshal(data, &cfg); err != nil {
        log.Fatal(err)
    }
}


func initWchat(){
    var err error
    // init logging
    lp, ok := cfg["log_file"]
    if ok && len(lp) > 0 {
        logging.SetWriter(gslog.WriterNew(lp))
    }
    // init template
    logging.Debug("tpl_path = ", path.Join(cfg["htdoc"], cfg["path_tpl"], "*.tpl"))
    tpl, err = template.ParseGlob(path.Join(cfg["htdoc"], cfg["path_tpl"], "*.tpl"))
    if err != nil {
        log.Fatal(err)
    }
    tpl, err = tpl.New("foot.tpl").Parse("\n</body>\n</html>")
    // init session
    err = sess.Init(path.Join(cfg["htdoc"], cfg["path_sess"]), 0)
    if err != nil {
        log.Fatal(err)
    }
    limit_body = cfgInt64("body_size", 1 << 20)
    limit_file = cfgInt64("file_size", 3 << 20)
    limit_all = limit_body + limit_file + (2 << 10)
}


// return logged in user id and username
func auth(w http.ResponseWriter, r *http.Request) (uid int, un string) {
    s := sess.Start(w, r)
    if us := s.Get("loginuserid"); len(us) > 0 {
        i, err := strconv.ParseInt(us, 10, 32)
        if err != nil {
            logging.Error("loginuserid = ", us, " ", err)
        } else {
            uid = int(i)
        }
        un = s.Get("loginusername")
    }
    if uid > 0 { UpdateLast(uid) }
    return
}


func getLastCle(w http.ResponseWriter, r *http.Request) int64 {
    l := sess.Start(w, r).Get("lastcle")
    i, err := strconv.ParseInt(l, 10, 64)
    if err == nil { return i }
    return 0
}


func hWchat(w http.ResponseWriter, r *http.Request) {
    tpl.ExecuteTemplate(w, "wchat.tpl", "")
}
