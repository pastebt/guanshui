#GuanShui (灌水)

Guan Shui BBS

## Build

After install and setup golang
```bash
go get github.com/pastebt/sess
go get github.com/pastebt/gslog
go get github.com/go-sql-driver/mysql

go build -o guanshui *.go
```

## Install

Setup mysql db, create tables in scheme.sql

```bash
mkdir /opt/guanshui
mkdir /opt/guanshui/static
mkdir /opt/guanshui/sess
```

Update conf.json

## Run

```bash
./guanshui conf.json start
```

or Stop

```bash
./guanshui conf.json stop
```
