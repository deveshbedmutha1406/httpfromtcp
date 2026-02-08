package main
import (
	"log"
	"fmt"
	"bytes"
	"io"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func(){
		defer f.Close()	// before you leave do this
		defer close(out)	// closing channel

		str := ""

		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				break
			}
			data = data[:n]	// read only n bytes which read.
			// when new line char then put it in chan.
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])
				data = data[i + 1:]
				out <- str
				str = ""
			}
			str += string(data)
		}
		if len(str) != 0 {
			out <- str
		}
	}()

	return out
}



func main(){
	listener, err := net.Listen("tcp", ":42069")
	if err != nil{
		log.Fatal("error", "error", err)
	}


	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error", "error", err)
		}

		for line := range getLinesChannel(conn) {
			fmt.Printf("read: %s\n", line) 
		}
	}

}



