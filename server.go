package blitzcache

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	cache    *Cache
	listener net.Listener
	addr     string
}

func NewServer(addr string, cache *Cache) *Server {
	return &Server{
		cache: cache,
		addr:  addr,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.listener = listener
	fmt.Printf("BlitzCache server listening on %s\n", s.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		args, err := s.readRESP(reader)
		if err != nil {
			if err != io.EOF {
				s.writeError(conn, err)
			}
			return
		}

		if len(args) == 0 {
			continue
		}

		cmd := strings.ToUpper(args[0])

		switch cmd {
		case "SET":
			s.handleSet(conn, args)
		case "GET":
			s.handleGet(conn, args)
		case "DEL":
			s.handleDel(conn, args)
		case "STATS":
			s.handleStats(conn)
		case "FLUSH":
			s.handleFlush(conn)
		case "PING":
			s.writeSimpleString(conn, "PONG")
		case "QUIT":
			return
		default:
			s.writeError(conn, fmt.Errorf("unknown command: %s", cmd))
		}
	}
}

func (s *Server) handleSet(conn net.Conn, args []string) {
	if len(args) < 3 {
		s.writeError(conn, fmt.Errorf("ERR wrong number of arguments"))
		return
	}

	key := args[1]
	value := []byte(args[2])
	ttl := 0 * time.Second

	if len(args) >= 5 && strings.ToUpper(args[3]) == "EX" {
		seconds, err := strconv.Atoi(args[4])
		if err == nil {
			ttl = time.Duration(seconds) * time.Second
		}
	}

	s.cache.Set(key, value, ttl)
	s.writeSimpleString(conn, "OK")
}

func (s *Server) handleGet(conn net.Conn, args []string) {
	if len(args) != 2 {
		s.writeError(conn, fmt.Errorf("ERR wrong number of arguments"))
		return
	}

	value, ok := s.cache.Get(args[1])
	if !ok {
		s.writeNull(conn)
		return
	}

	s.writeBulkString(conn, string(value))
}

func (s *Server) handleDel(conn net.Conn, args []string) {
	if len(args) != 2 {
		s.writeError(conn, fmt.Errorf("ERR wrong number of arguments"))
		return
	}

	deleted := s.cache.Delete(args[1])
	if deleted {
		s.writeInteger(conn, 1)
	} else {
		s.writeInteger(conn, 0)
	}
}

func (s *Server) handleStats(conn net.Conn) {
	stats := s.cache.Stats()
	result := fmt.Sprintf("hits:%d misses:%d sets:%d deletes:%d evictions:%d keys:%d",
		stats["hits"], stats["misses"], stats["sets"],
		stats["deletes"], stats["evictions"], stats["keys"])
	s.writeSimpleString(conn, result)
}

func (s *Server) handleFlush(conn net.Conn) {
	s.cache.Flush()
	s.writeSimpleString(conn, "OK")
}

func (s *Server) readRESP(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)

	if len(line) == 0 {
		return nil, fmt.Errorf("empty line")
	}

	if line[0] == '*' {
		count, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, err
		}

		args := make([]string, count)
		for i := 0; i < count; i++ {
			bulkLine, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}

			bulkLine = strings.TrimSpace(bulkLine)
			if bulkLine[0] != '$' {
				return nil, fmt.Errorf("expected bulk string")
			}

			length, err := strconv.Atoi(bulkLine[1:])
			if err != nil {
				return nil, err
			}

			data := make([]byte, length)
			_, err = io.ReadFull(reader, data)
			if err != nil {
				return nil, err
			}

			reader.ReadString('\n')
			args[i] = string(data)
		}

		return args, nil
	}

	return strings.Fields(line), nil
}

func (s *Server) writeSimpleString(conn net.Conn, str string) {
	conn.Write([]byte("+" + str + "\r\n"))
}

func (s *Server) writeError(conn net.Conn, err error) {
	conn.Write([]byte("-" + err.Error() + "\r\n"))
}

func (s *Server) writeBulkString(conn net.Conn, str string) {
	conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)))
}

func (s *Server) writeNull(conn net.Conn) {
	conn.Write([]byte("$-1\r\n"))
}

func (s *Server) writeInteger(conn net.Conn, n int) {
	conn.Write([]byte(fmt.Sprintf(":%d\r\n", n)))
}
