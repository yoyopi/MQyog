# MQyog
[![wercker status](https://app.wercker.com/status/d19e73df9017e9c60bebd47368c5f2bd/s "wercker status")](https://app.wercker.com/project/bykey/d19e73df9017e9c60bebd47368c5f2bd)

MQyog is a simple HTTP message queue written in Go with Tokyo Cabinet.

Feature
======

* Very simple, less than 300 lines Go code.
* Very fast, more than 10000 requests/sec.
* High concurrency, support the tens of thousands of concurrent connections.
* Multiple queue.
* Low memory consumption, mass data storage, storage dozens of GB of data takes less than 100MB of physical memory buffer.
* Convenient to change the maximum queue length of per-queue.
* Queue status view.
* Be able to view the contents of the specified queue ID.
* Multi-Character encoding support.

Install 
======
  ```
ulimit -SHn 65535

wget http://httpsqs.googlecode.com/files/libevent-2.0.12-stable.tar.gz
tar zxvf libevent-2.0.12-stable.tar.gz
cd libevent-2.0.12-stable/
./configure --prefix=/usr/local/libevent-2.0.12-stable/
make
make install
cd ../

wget http://httpsqs.googlecode.com/files/tokyocabinet-1.4.47.tar.gz
tar zxvf tokyocabinet-1.4.47.tar.gz
cd tokyocabinet-1.4.47/
./configure --prefix=/usr/local/tokyocabinet-1.4.47/ 
make
make install
cd ../

tar -xzf go1.4.2.linux-amd64.tar.gz -C /usr/local/
which pkg-config
export PKG_CONFIG_PATH=/usr/local/tokyocabinet/lib/pkgconfig/
export GOPATH=/project/golang/
export PATH=$PATH:/usr/local/go/bin

go get github.com/codegangsta/negroni


cd /project/golang/src/mq
go build
go install
/project/golang/bin/mq
    -auth="": auth password to access httpmq
    -cpu=1: cpu number for httpmq
    -ip="0.0.0.0": ip address to listen on
    -maxqueue=1000000: the max queue length
    -port="9091": port to listen on
    -verbose=true: output log
  ```

1. PUT text message into a queue

  HTTP GET protocol (Using curl for example):
  ```
  curl "http://host:port/?name=your_queue_name&opt=put&data=url_encoded_text_message"
  ```
  HTTP POST protocol (Using curl for example):
  ```
  curl -d "url_encoded_text_message" "http://host:port/?name=your_queue_name&opt=put"
  ```

2. GET text message from a queue

  HTTP GET protocol (Using curl for example):
  ```
  curl "http://host:port/?charset=utf-8&name=your_queue_name&opt=get"
  ```

3. View queue status

  HTTP GET protocol (Using curl for example):
  ```
  curl "http://host:port/?name=your_queue_name&opt=status"
  ```
4. View queue details

  HTTP GET protocol (Using curl for example):
  ```
  curl "http://host:port/?name=your_queue_name&opt=view&pos=1"
  ```
5. Reset queue

  HTTP GET protocol (Using curl for example):
  ```
  curl "http://host:port/?name=your_queue_name&opt=reset&pos=1"
  ```

Benchmark
========

Test machine:
  ```
  CPU:    2  AMD Athlon(tm) II X2 245 Processor
  Memory: Size: 2048 MB
          Locator: DIMM0
          Range Size: 2 GB
          Size: 2048 MB
          Locator: DIMM1
          Range Size: 2 GB
          Size: No Module Installed
          Locator: DIMM2
          Size: No Module Installed
          Locator: DIMM3
  ```


###PUT queue:

    ab -k -c 1000 -n 10000 "http://127.0.0.1:9091/?name=yog&opt=put&data=aaaaaaaaaaaaaaaaaaaaaaaa"
    This is ApacheBench, Version 2.3 <$Revision: 655654 $>
    Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
    Licensed to The Apache Software Foundation, http://www.apache.org/
    
    Benchmarking 127.0.0.1 (be patient)
    Completed 1000 requests
    Completed 2000 requests
    Completed 3000 requests
    Completed 4000 requests
    Completed 5000 requests
    Completed 6000 requests
    Completed 7000 requests
    Completed 8000 requests
    Completed 9000 requests
    Completed 10000 requests
    Finished 10000 requests
    
    
    Server Software:        
    Server Hostname:        127.0.0.1
    Server Port:            9091
    
    Document Path:          /?name=yog&opt=put&data=aaaaaaaaaaaaaaaaa
    Document Length:        13 bytes
    
    Concurrency Level:      1000
    Time taken for tests:   0.771 seconds
    Complete requests:      10000
    Failed requests:        0
    Write errors:           0
    Keep-Alive requests:    10000
    Total transferred:      1640000 bytes
    HTML transferred:       130000 bytes
    Requests per second:    12964.69 [#/sec] (mean)
    Time per request:       77.133 [ms] (mean)
    Time per request:       0.077 [ms] (mean, across all concurrent requests)
    Transfer rate:          2076.38 [Kbytes/sec] received
    
    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        0    2   7.7      0      41
    Processing:     0   70  74.9     73     473
    Waiting:        0   70  74.9     73     473
    Total:          0   72  75.8     76     473
    
    Percentage of the requests served within a certain time (ms)
      50%     76
      66%     91
      75%     98
      80%    110
      90%    183
      95%    216
      98%    272
      99%    310
     100%    473 (longest request)

###GET queue:

    ab -k -c 1000 -n 10000 "http://127.0.0.1:9091/?name=yog&opt=get"                                                                                                   [system]
    This is ApacheBench, Version 2.3 <$Revision: 655654 $>
    Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
    Licensed to The Apache Software Foundation, http://www.apache.org/
    
    Benchmarking 127.0.0.1 (be patient)
    Completed 1000 requests
    Completed 2000 requests
    Completed 3000 requests
    Completed 4000 requests
    Completed 5000 requests
    Completed 6000 requests
    Completed 7000 requests
    Completed 8000 requests
    Completed 9000 requests
    Completed 10000 requests
    Finished 10000 requests
    
    
    Server Software:        
    Server Hostname:        127.0.0.1
    Server Port:            9091
    
    Document Path:          /?name=yog&opt=get
    Document Length:        512 bytes
    
    Concurrency Level:      1000
    Time taken for tests:   0.703 seconds
    Complete requests:      10000
    Failed requests:        0
    Write errors:           0
    Keep-Alive requests:    10000
    Total transferred:      6640000 bytes
    HTML transferred:       5120000 bytes
    Requests per second:    14227.83 [#/sec] (mean)
    Time per request:       70.285 [ms] (mean)
    Time per request:       0.070 [ms] (mean, across all concurrent requests)
    Transfer rate:          9225.86 [Kbytes/sec] received
    
    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        0    1   5.3      0      33
    Processing:     0   49  61.2     20     449
    Waiting:        0   49  61.2     20     449
    Total:          0   50  62.0     22     471
    
    Percentage of the requests served within a certain time (ms)
      50%     22
      66%     67
      75%     87
      80%    105
      90%    128
      95%    161
      98%    224
      99%    240
     100%    471 (longest request)
