package p2p

import (
	"bufio"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/network/p2p"
	"io"
	"net"
	"sync"
)

type Link struct {
	Connection          net.Conn
	Reader              *bufio.Reader
	Writer              *bufio.Writer
	RawMessageProcessor p2p.RawMessageProcessor
	LinkClosed          p2p.LinkClosed
	ConnectOut          bool

	senderLock sync.Mutex
}

func (l *Link) RemoteAddr() net.Addr {
	return l.Connection.RemoteAddr()
}

func (l *Link) Start() {

	l.Reader = bufio.NewReader(l.Connection)
	l.Writer = bufio.NewWriter(l.Connection)

	go l.StartReceiving()
}

func (l *Link) StartReceiving() {
	defer func() {
		l.Close()
		l.LinkClosed(l)
	}()

	for {
		msgPrefix := make([]byte, p2p.MESSAGE_PREFIX_LEN, p2p.MESSAGE_PREFIX_LEN)
		data := make([]byte, p2p.Max_buffer_size, p2p.Max_buffer_size)

		n, err := io.ReadFull(l.Reader, msgPrefix) //l.Reader.Read(msgPrefix[:])
		if n != p2p.MESSAGE_PREFIX_LEN {
			log.Log.Println("get msg prefix failed: need ", p2p.MESSAGE_PREFIX_LEN, " bytes, got ", n, " bytes. ", err)
		}

		if err != nil {
			if err == io.EOF {
				log.Log.Println("disconnect remote: ", err)
				break
			}
			log.Log.Println("got an network error: ", err)
			continue
		}

		prefix := p2p.MessagePrefix{}
		err = prefix.FromBytes(msgPrefix)
		if err != nil {
			log.Log.Warn("unknown message: ", err.Error())
			unreadSize := l.Reader.Size()
			_, _ = l.Reader.Discard(unreadSize)
			continue
		}

		if prefix.Size > p2p.MAX_MESSAGE_LEN {
			_, _ = l.Reader.Discard(int(prefix.Size))
			continue
		}

		n, err = io.ReadFull(l.Reader, data[:prefix.Size])
		//n, err := l.Reader.Read(data[:prefix.Size])

		if int32(n) != prefix.Size {
			log.Log.Println("error message: need ", prefix.Size, "bytes bug got ", n, " bytes")
			//log.Log.Println("error ", err.Error())
			return
		}

		if err != nil {
			if err == io.EOF {
				log.Log.Println("disconnect remote: ", err)
				break
			}
			log.Log.Println("got an network error: ", err)
			continue
		}

		go l.RawMessageProcessor(data[:prefix.Size], l)
	}
}

func (l *Link) SendMessage(msg p2p.Message) (n int, err error) {
	data, err := msg.ToRawMessage()
	if err != nil {
		return
	}

	return l.SendData(data)
}

func (l *Link) SendData(data []byte) (n int, err error) {
	l.senderLock.Lock()
	defer l.senderLock.Unlock()

	n, err = l.Writer.Write(data)
	if err != nil {
		log.Log.Warn("got an error when write to buffer: ", err.Error())
		return
	}

	if n != len(data) {
		log.Log.Warn("not write the complete data to buffer: ", n, " bytes written but need: ", len(data), " bytes")
	}

	err = l.Writer.Flush()

	if n == 0 {
		log.Log.Println("sent 0 bytes to ", l.RemoteAddr())
		if err != nil {
			log.Log.Warn(" and got an error: ", err.Error())
		}
	}
	return
}

func (l *Link) Close() {
	l.Connection.Close()
}
