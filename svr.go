package main

import (
    "os"
    "fmt"
    "time"
    "strings"
    "net/http"
    "github.com/pastebt/gslog"
)


var logging = gslog.GetLogger("").SetFmt(func(n string, l string, msg string) string{
    return time.Now().Format("20060102-150405") + "-" + l[:1] + "-" + msg + "\n"
})


type MyRW struct {
    http.ResponseWriter
}


func (m MyRW)WriteHeader(s int) {
    logging.Debug("return code: ", s)
    m.ResponseWriter.WriteHeader(s)
}


type mux struct {
    http.ServeMux
}


func (m *mux)ServeHTTP(w http.ResponseWriter, r *http.Request) {
    logging.Debugf("%-15s %s %s", strings.Split(r.RemoteAddr, ":")[0],
                   r.Method, r.RequestURI)
    m.ServeMux.ServeHTTP(MyRW{w}, r)
}


// limit request size
func limitx(h func(http.ResponseWriter, *http.Request), limit int64) func(
        http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.ContentLength > limit {
            w.WriteHeader(http.StatusLengthRequired)
            return
        }
        r.Body = http.MaxBytesReader(w, r.Body, limit)
        h(w, r)
    }
}


// limit request size to 1k
func limit1k(h func(http.ResponseWriter, *http.Request)) func(
        http.ResponseWriter, *http.Request) {
    return limitx(h, 1 << 10)
}


func addHandle(m *mux) {
    m.HandleFunc("/", limit1k(hIndex))
    m.HandleFunc("/guanshui/",  limit1k(hWchat))
    m.HandleFunc("/static/", limit1k(hStatic))
    m.HandleFunc("/phh/reg.phh",  limit1k(hReg))
    m.HandleFunc("/phh/list.phh", limit1k(hList))
    m.HandleFunc("/phh/post.phh", limitx(hPost, limit_all))
    m.HandleFunc("/phh/atta.phh", limit1k(hAttach))
}


func startSvr() {
    h := &mux{*http.NewServeMux()}
    addHandle(h)
    go func() {
        logging.Infof("serve http at %s", cfg["port_http"])
        err := http.ListenAndServe(":" + cfg["port_http"], h)
        if err != nil { panic (err) }
    }()
    s := &mux{*http.NewServeMux()}
    addHandle(s)
    logging.Infof("serve https at %s", cfg["port_https"])
    err := http.ListenAndServeTLS(":" + cfg["port_https"],
                                  cfg["cert.pem"], cfg["key.pem"], s)
    if err != nil { panic(err) }
}


func stopSvr() {
    http.Get("http://127.0.0.1:" + cfg["port_http"] + "/stop_this_server")
}


func usage() {
    fmt.Printf("Usage: %s config_json_file start|stop\n", os.Args[0])
    os.Exit(1)
}


func main() {
    if len(os.Args) != 3 { usage() }

    initCfg(os.Args[1])

    switch os.Args[2] {
    case "start":
        initWchat()
        initDB()
        startSvr()
    case "stop":
        stopSvr()
    default:
        usage()
    }
}
