package main

import (
"flag"
"fmt"
"runtime"
_"os"  
tokyocabinet "tc" 
"strconv"
"log"
"math"
_"net"
"time" 
"net/http"
"github.com/codegangsta/negroni"
"io/ioutil"
"bytes"
)
const VERSION = "0.8" 
var default_maxqueue, keepalive, cpu  *int
var ip, port, dbpath *string
var db tokyocabinet.HDB 
var verbose *bool

func mq_read_metadata(name string) []string {
	maxqueue := name + ".maxqueue"
	data1, _ := db.Get([]byte(maxqueue))
	if len(data1) == 0 {
		data1 = []byte(strconv.Itoa(*default_maxqueue))
	}
	putpos := name + ".putpos"
	data2, _ := db.Get([]byte(putpos))
	getpos := name + ".getpos"
	data3, _ := db.Get([]byte(getpos))
	return []string{string(data1), string(data2), string(data3)}
}

func mq_now_getpos(name string) string {
	metadata := mq_read_metadata(name)
	maxqueue, _ := strconv.Atoi(metadata[0])
	putpos, _ := strconv.Atoi(metadata[1])
	getpos, _ := strconv.Atoi(metadata[2])

	if getpos == 0 && putpos > 0 {
		getpos = 1 // first get operation, set getpos 1
	} else if getpos < putpos {
		getpos++ // 1nd lap, increase getpos
	} else if getpos > putpos && getpos < maxqueue {
		getpos++ // 2nd lap
	} else if getpos > putpos && getpos == maxqueue {
		getpos = 1 // 2nd first operation, set getpos 1
	} else {
		return "0" // all data in queue has been get
	}

	data := strconv.Itoa(getpos)
	db.Put([]byte(name+".getpos"), []byte(data))
	return data
}
func mq_now_putpos(name string) string {
	metadata := mq_read_metadata(name)
	maxqueue, _ := strconv.Atoi(metadata[0])
	putpos, _ := strconv.Atoi(metadata[1])
	getpos, _ := strconv.Atoi(metadata[2])

	putpos++              // increase put queue pos
	if putpos == getpos { // queue is full
		return "0" // return 0 to reject put operation
	} else if getpos <= 1 && putpos > maxqueue { // get operation less than 1
		return "0" // and queue is full, just reject it
	} else if putpos > maxqueue { //  2nd lap
		metadata[1] = "1"  
	} else {  
		metadata[1] = strconv.Itoa(putpos)
	}

	err:=db.Put([]byte(name+".putpos"), []byte(metadata[1]))
	if err != nil {
	log.Fatalln("db.Get(), err:", err)
	}
	return metadata[1]
}
 

func main() {
	default_maxqueue = flag.Int("maxqueue", 1000000, "the max queue length")
	ip = flag.String("ip", "0.0.0.0", "ip address to listen on")
	port = flag.String("port", "9091", "port to listen on") 
	dbpath = flag.String("db", "MQ.db", "database path") 
	cpu = flag.Int("cpu", runtime.NumCPU(), "cpu number for MQ")
	keepalive = flag.Int("k", 60, "keepalive timeout for MQ")
	flag.Parse()
	fmt.Printf("%d,%s,%s,%s,%d,%d",*default_maxqueue,*ip,*port,*dbpath,*cpu,*keepalive)
	
	log.Printf("start(), sucess:")
	
	
	var err error
	db =*tokyocabinet.NewHDB()
	err = db.Open(*dbpath, tokyocabinet.BDBOWRITER|tokyocabinet.BDBOCREAT|tokyocabinet.BDBOTRUNC)
// 	os.Remove(*dbpath)
	if err != nil {
	log.Fatalln("db.Get(), err:", err)
	}
	defer   db.Close()
	runtime.GOMAXPROCS(*cpu)
	putnamechan := make(chan string, 1000)
	putposchan := make(chan string, 1000)
	getnamechan := make(chan string, 1000)
	getposchan := make(chan string, 1000) 
	go func(chan string, chan string) {
		for {
			name := <-putnamechan
			putpos := mq_now_putpos(name)
			putposchan <- putpos
		}
	}(putnamechan, putposchan)

	go func(chan string, chan string) {
		for {
			name := <-getnamechan
			getpos := mq_now_getpos(name)
			getposchan <- getpos
		}
	}(getnamechan, getposchan)
	
	go func() {
		for {
		   time.Sleep(5*time.Second);
			err := db.Sync()
			if err != nil {
			log.Fatalln("db.Sync(), err:", err)
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var data string
		var buf []byte
		name := r.FormValue("name")
		opt := r.FormValue("opt")
		pos := r.FormValue("pos")
		num := r.FormValue("num")
		charset := r.FormValue("charset")
			if r.Method == "GET" {
			data = r.FormValue("data")
		} else if r.Method == "POST" {
			if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
				data = r.PostFormValue("data")
			} else {
				buf, _ = ioutil.ReadAll(r.Body)
				r.Body.Close()
			}
		}

		if len(name) == 0 || len(opt) == 0 {
			w.Write([]byte("MQ_ERROR"))
			return
		}

		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-type", "text/plain")
		if len(charset) > 0 {
			w.Header().Set("Content-type", "text/plain; charset="+charset)
		}
		if opt == "put" {
			if len(data) == 0 && len(buf) == 0 {
				w.Write([]byte("MQ_PUT_ERROR"))
				return
			}

			putnamechan <- name
			putpos := <-putposchan

			if putpos != "0" {
				queue_name := name + putpos
				if data != "" {
					
	
		 
				err:=db.Put([]byte(queue_name), []byte(data))
					if err != nil {
						log.Fatalln("db.Get(), err:", err)
				}
			 

				} else if len(buf) > 0 {
					err:=db.Put([]byte(queue_name), buf)
						if err != nil {
					log.Fatalln("db.Get(), err:", err)
					}
				} 
				w.Header().Set("Pos", putpos)
				w.Write([]byte("MQ_PUT_OK"))
			} else {
				w.Write([]byte("MQ_PUT_END"))
			}
		} else if opt == "get" {
			getnamechan <- name
			getpos := <-getposchan

			if getpos == "0" {
				w.Write([]byte("MQ_GET_END"))
			} else {
				queue_name := name + getpos
				v, err := db.Get([]byte(queue_name))
				if err == nil {
					w.Header().Set("Pos", getpos)
					w.Write(v)
				} else {
					w.Write([]byte("MQ_GET_ERROR"))
				}
			}
		} else if opt == "status" {
			metadata := mq_read_metadata(name)
			maxqueue, _ := strconv.Atoi(metadata[0])
			putpos, _ := strconv.Atoi(metadata[1])
			getpos, _ := strconv.Atoi(metadata[2])
			var buffer bytes.Buffer
			var ungetnum float64
			var put_times, get_times string
			if putpos >= getpos {
				ungetnum = math.Abs(float64(putpos - getpos))
				put_times = "1st lap"
				get_times = "1st lap"
			} else if putpos < getpos {
				ungetnum = math.Abs(float64(maxqueue - getpos + putpos))
				put_times = "2nd lap"
				get_times = "1st lap"
			}

//			buf := fmt.Sprintf("MQ v%s\n", VERSION)
//			buf += fmt.Sprintf("------------------------------\n")
//			buf += fmt.Sprintf("Queue Name: %s\n", name)
//			buf += fmt.Sprintf("Maximum number of queues: %d\n", maxqueue)
//			buf += fmt.Sprintf("Put position of queue (%s): %d\n", put_times, putpos)
//			buf += fmt.Sprintf("Get position of queue (%s): %d\n", get_times, getpos)
//			buf += fmt.Sprintf("Number of unread queue: %g\n\n", ungetnum)
			buffer.WriteString(fmt.Sprintf("MQ v%s", VERSION))
			buffer.WriteString("\r\n----------------")
			buffer.WriteString(fmt.Sprintf("\r\n Queue Name: %s", name))
			buffer.WriteString(fmt.Sprintf("\r\n maximum number of queues: %d", maxqueue))
			buffer.WriteString(fmt.Sprintf("\r\n Put position of queue (%s): %d", put_times, putpos))
			buffer.WriteString(fmt.Sprintf("\r\n Get position of queue (%s): %d", get_times, getpos))
			buffer.WriteString(fmt.Sprintf("\r\n Number of unread queue: %g", ungetnum))

			w.Write([]byte(buffer.String()))
		} else if opt == "view" {
			v, err := db.Get([]byte(name+pos))
			if err == nil {
				w.Write([]byte(v))
			} else {
				w.Write([]byte("MQ_VIEW_ERROR"))
			}
		} else if opt == "reset" {
			maxqueue := strconv.Itoa(*default_maxqueue)
			db.Put([]byte(name+".maxqueue"), []byte(maxqueue))
			db.Put([]byte(name+".putpos"), []byte("0"))
			db.Put([]byte(name+".getpos"), []byte("0"))
			w.Write([]byte("MQ_RESET_OK"))
		} else if opt == "maxqueue" {
			maxqueue, _ := strconv.Atoi(num)
			if maxqueue > 0 && maxqueue <= 10000000 {
				db.Put([]byte(name+".maxqueue"), []byte(num))
				w.Write([]byte("MQ_MAXQUEUE_OK"))
			} else {
				w.Write([]byte("MQ_MAXQUEUE_CANCLE"))
			}
		}
	})

	n := negroni.New(negroni.NewRecovery())
	n.UseHandler(mux)
	n.Run(*ip + ":" + *port)
}
