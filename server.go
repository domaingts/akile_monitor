package main

import (
	"akile_monitor/client/model"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

func newServer() error {
	mux := &http.ServeMux{}
	mux.HandleFunc(cfg.WebUri, ws)
	mux.HandleFunc(cfg.UpdateUri, monitor)
	mux.HandleFunc("/delete", delete)
	mux.HandleFunc("/info", info)
	engine := nbhttp.NewEngine(nbhttp.Config{
		Network:                 "tcp",
		Addrs:                   []string{cfg.Listen},
		MaxLoad:                 100000,
		ReleaseWebsocketPayload: true,
		Handler:                 mux,
	})
	return engine.Start()
}

func info(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getInfo(w, r)
	case "POST":
	}
}

func getInfo(w http.ResponseWriter, _ *http.Request) {
	var ret []*Host
	err := filedb.Model(&Host{}).Find(&ret).Error
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
		return
	}
	by, _ := json.Marshal(&ret)
	w.WriteHeader(http.StatusOK)
	w.Write(by)
}

func updateInfo(w http.ResponseWriter, r *http.Request) {
	var ret UpdateRequest
	err := json.NewDecoder(r.Body).Decode(&ret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "bad request")
		return
	}

	if ret.AuthSecret != cfg.AuthSecret {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "auth failed")
		return
	}

	var h Host

	filedb.Model(&Host{}).Where("name = ?", ret.Name).First(&h)
	if h.Name == "" {
		h = ret.Host
		filedb.Model(&Host{}).Create(&h)
	} else {
		h = ret.Host
		filedb.Model(&Host{}).Where("name = ?", ret.Name).Save(&h)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func delete(w http.ResponseWriter, r *http.Request) {
	var req DeleteHostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "bad request")
		return
	}

	if req.AuthSecret != cfg.AuthSecret {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "auth failed")
		return
	}

	var data Data
	db.Model(&Data{}).Where("name = ?", req.Name).First(&data)
	if data.Name == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "not found")
		return
	}

	db.Delete(&Data{}, "name = ?", req.Name)
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func ws(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.NewUpgrader()
	upgrader.KeepaliveTime = time.Duration(time.Second * 10)
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	upgrader.EnableCompression(true)
	upgrader.OnClose(close)
	upgrader.OnOpen(func(c *websocket.Conn) {
		log.Println("connected", c.RemoteAddr().String())
	})
	upgrader.OnMessage(func(c *websocket.Conn, mt websocket.MessageType, b []byte) {
		data := fetchData()
		if err := c.WriteMessage(mt, append([]byte("data"), data...)); err != nil {
			log.Printf("client: %s, write :%s\n", c.RemoteAddr().String(), err.Error())
			c.Close()
		}
	})
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade: %s\n", conn.RemoteAddr().String())
	}
}

func monitor(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.NewUpgrader()
	upgrader.KeepaliveTime = time.Duration(time.Second * 10)
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	upgrader.EnableCompression(true)
	upgrader.OnClose(close)
	upgrader.OnMessage(func(c *websocket.Conn, mt websocket.MessageType, b []byte) {
		if authed, ok := c.Session().(bool); ok && authed {
			var buf bytes.Buffer
			buf.Write(b)
			r, _ := gzip.NewReader(&buf)
			message, _ := io.ReadAll(r)
			r.Close()

			var d model.Data
			err := json.Unmarshal(message, &d)
			if err != nil {
				log.Printf("client: %s,unmarshal: %s\n", c.RemoteAddr().String(), err.Error())
				c.Close()
				return
			}

			var data Data
			db.Model(&Data{}).Where("name = ?", d.Host.Name).First(&data)
			if data.Name == "" {
				db.Create(&Data{Name: d.Host.Name, Data: string(message)})
			} else {
				db.Model(&Data{}).Where("name = ?", d.Host.Name).Update("data", string(message))
			}
		} else {
			if string(b) != cfg.AuthSecret {
				log.Printf("client: %s, auth failed\n", c.Conn.RemoteAddr().String())
				c.Close()
				return
			}
			c.SetSession(true)
			if err := c.WriteMessage(mt, []byte("auth success")); err != nil {
				log.Printf("client: %s, write: %s\n", c.Conn.RemoteAddr().String(), err.Error())
				c.Close()
				return
			}
		}
	})
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade: %s\n", conn.RemoteAddr().String())
	}
}

func close(c *websocket.Conn, err error) {
	log.Printf("client: %s, closed: %s\n", c.RemoteAddr().String(), err.Error())
}
