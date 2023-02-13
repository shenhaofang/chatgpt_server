package zipkin

import (
	"os"
	"sync"
	"time"
)

type RotateWriter struct {
	lock     sync.Mutex
	filename string // should be set to the actual filename
	fp       *os.File
	ch       chan []byte
	curTime  string
}

const (
	MAX_CHAN_SIZE = 10000
	TIME_FORMAT   = "2006010215"
)

// Make a new RotateWriter. Return nil if error occurs during setup.
func NewRotaterWriter(filename string) *RotateWriter {
	w := &RotateWriter{filename: filename}
	w.ch = make(chan []byte, MAX_CHAN_SIZE)

	curTime := time.Now().Format(TIME_FORMAT)
	w.curTime = curTime

	var err error
	w.fp, err = os.Create(w.filename)
	if err != nil {
		return nil
	}
	go w.doRotate()
	go w.doWrite()
	return w
}

func (w *RotateWriter) doWrite() {
	for output := range w.ch {
		w.lock.Lock()
		w.fp.Write(output)
		w.lock.Unlock()
	}
}

// Write satisfies the io.Writer interface.
func (w *RotateWriter) Write(output []byte) (int, error) {
	if len(w.ch) < MAX_CHAN_SIZE {
		w.ch <- output
	}
	return len(output), nil
}

func (w *RotateWriter) doRotate() {
	ticker := time.NewTicker(2 * time.Second)
	for now := range ticker.C {
		curTime := now.Format(TIME_FORMAT)
		if w.curTime != curTime {
			w.rotate()
		}
	}
}

// Perform the actual act of rotating and reopening file.
func (w *RotateWriter) rotate() (err error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	// Close existing file if open
	if w.fp != nil {
		err = w.fp.Close()
		w.fp = nil
		if err != nil {
			return
		}
	}
	// Rename dest file if it already exists
	_, err = os.Stat(w.filename)
	if err == nil {
		curTime := time.Now().Format(TIME_FORMAT)

		err = os.Rename(w.filename, w.filename+"."+curTime)
		w.curTime = curTime
		if err != nil {
			return
		}
	}

	// Create a file.
	w.fp, err = os.Create(w.filename)
	return
}
