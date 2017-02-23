package main

import (
    "os"
//    "fmt"
    "path"
//    "path/filepath"
    "strings"
    "net/http"
)


func hIndex(w http.ResponseWriter, r *http.Request) {
    if r.RequestURI == "/stop_this_server" {
        rip := strings.Split(r.RemoteAddr, ":")[0]
        if rip == "127.0.0.1" {
            //logging.Info("/stop_this_server")
            os.Exit(0)
        }
    }
    if r.RequestURI != "/guanshui/" {
        http.Redirect(w, r, "/guanshui/", http.StatusFound)
        return
    }
}


func hStatic(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, path.Join(cfg["htdoc"], cfg["path_static"], path.Base(r.URL.Path)))
}
