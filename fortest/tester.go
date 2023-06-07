package main

import (
	"flag"
	"io"
	"log"
	"net"
	"time"
)

// tcp сервер которыи по подключеному клиенту (по его часовому поясу) отдает время

// 3 порта - 8010 8020 8030 -> 8010 возвращает время вашингтона, 8020 возвращает время токио, 8030 возвращает время москвы

func main() {
	var port string

	flag.StringVar(&port, "port", "byaka", "boom")
	flag.Parse()

	log.Println(port)
	runServer(port)

}

func runServer(port string) {
	listen, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("run")

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println(listen.Addr().String())
		go Clock(conn, port)
	}
}

func Clock(con net.Conn, port string) {
	defer con.Close()
	var now string
	// проверить порт и выдать время

	for {
		switch port {
		case "8010":
			loc, _ := time.LoadLocation("Asia/Tokyo")
			now = time.Now().In(loc).Format("15:04:05\n")
			_, err := io.WriteString(con, now)
			if err != nil {
				return
			}
			time.Sleep(time.Second * 1)
		case "8020":
			loc, _ := time.LoadLocation("Europe/Berlin")
			now = time.Now().In(loc).Format("15:04:05\n")
			_, err := io.WriteString(con, now)
			if err != nil {
				return
			}
			time.Sleep(time.Second * 1)
		case "8030":
			loc, _ := time.LoadLocation("America/Juneau")
			now = time.Now().In(loc).Format("15:04:05\n")
			_, err := io.WriteString(con, now)
			if err != nil {
				return
			}
			time.Sleep(time.Second * 1)
		default:
			log.Panic()
		}

	}
}

// type RunTicker struct {
// 	F      func(ctx context.Context)
// 	ctx    context.Context
// 	cancel context.CancelFunc
// }

// func (r *RunTicker) Start() {
// 	go r.F(r.ctx)
// }

// func (r *RunTicker) Reset() {
// 	r.cancel()

// }

// func runFunc(s int, ss func()) (*RunTicker, error) {
// 	if s > 60 || s < 0 {
// 		return nil, fmt.Errorf("invalid second format")
// 	}
// 	ctx, cancel := context.WithCancel(context.Background())

// 	asd := RunTicker{
// 		F: func(ctx context.Context) {
// 			t := time.NewTicker(time.Second * 1)
// 			for {
// 				select {
// 				case res := <-t.C:
// 					if res.Second() == s {
// 						go ss()
// 						fmt.Println(res)
// 					}
// 				case <-ctx.Done():
// 					return
// 				}
// 			}
// 		},
// 		ctx:    ctx,
// 		cancel: cancel,
// 	}

// 	return &asd, fmt.Errorf("asdasd")
// }

// // каждую 20 секунду каждои минуты возвращать функ.
