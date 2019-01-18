package godol

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	pb "gopkg.in/cheggaaa/pb.v1"
)

// Progress is progress bar for display purpose
type progress struct {
	pool *pb.Pool
	bars []*pb.ProgressBar
}

// Godol is a module for downloading a file concurrently
type Godol struct {
	url         string
	destination string
	workers     int
	fileName    string
	wg          sync.WaitGroup
	file        *os.File
	totalSize   int64
	progress
}

// Option to give to Godol
type Option func(*Godol)

// New creates an instance of Godol with provided options
func New(options ...Option) *Godol {
	g := &Godol{}

	for _, option := range options {
		option(g)
	}

	return g
}

// WithURL option
func WithURL(url string) Option {
	return func(g *Godol) {
		g.url = url
	}
}

// WithDestination option
func WithDestination(dest string) Option {
	return func(g *Godol) {
		g.destination = dest
	}
}

// WithWorker option
func WithWorker(n int) Option {
	return func(g *Godol) {
		g.workers = n
	}
}

// WithFilename option
func WithFilename(name string) Option {
	return func(g *Godol) {
		g.fileName = name
	}
}

// Start starts the task
func (g *Godol) Start() error {
	err := g.checkURLHeader()
	if err != nil {
		return err
	}

	g.file, err = os.Create(fmt.Sprintf("%s/%s", g.destination, g.fileName))
	if err != nil {
		return err
	}

	fmt.Printf("File will be saved in   : %s \n", g.destination)

	defer g.file.Close()

	subBytes := g.totalSize / int64(g.workers)

	var start, end int64

	for i := 0; i < g.workers; i++ {

		bar := pb.New(0).Prefix(fmt.Sprintf("worker %d  0%% ", i+1))
		bar.ShowSpeed = true
		bar.SetMaxWidth(100)
		bar.SetUnits(pb.U_BYTES_DEC)
		bar.SetRefreshRate(time.Second)
		bar.ShowPercent = true

		g.bars = append(g.bars, bar)

		if i == g.workers {
			end = g.totalSize
		} else {
			end = start + subBytes
		}

		g.wg.Add(1)

		go g.spawnWorker(start, end, i)

		start = end

	}
	g.pool, err = pb.StartPool(g.bars...)
	if err != nil {
		return err
	}

	g.wg.Wait()
	g.pool.Stop()
	return err
}

func (g *Godol) checkURLHeader() (err error) {
	res, err := http.Head(g.url)
	if err != nil {
		return err
	}

	if g.fileName == "" {
		g.fileName = path.Base(res.Request.URL.Path)
	}

	header := res.Header
	acceptRanges, supported := header["Accept-Ranges"]
	if !supported {
		return fmt.Errorf("the link %s doesn't support concurrent download", g.url)
	} else if supported && acceptRanges[0] != "bytes" {
		return fmt.Errorf("the link %s doesn't have supported type for concurrent download", g.url)
	}
	size, err := strconv.ParseInt(header["Content-Length"][0], 10, 64)

	if err != nil {
		return err
	}

	g.totalSize = size

	fmt.Printf("File name \t\t: %s \n", g.fileName)
	fmt.Printf("File size \t\t: %d bytes\n", g.totalSize)

	return err

}

func (g *Godol) spawnWorker(start, end int64, i int) {
	body, size, err := g.getPartialBody(start, end)
	if err != nil {
		log.Fatalf("an error occurred in worker %d; error: %s\n", i, err.Error())
	}

	defer body.Close()
	defer g.bars[i].Finish()
	defer g.wg.Done()

	g.bars[i].Total = size

	buf := make([]byte, 4*1024)
	var written int64
	flag := map[int64]bool{}

	for {
		nRead, err := body.Read(buf)
		if nRead > 0 {
			nWrite, err := g.file.WriteAt(buf[0:nRead], start)
			if err != nil {
				fmt.Println(err.Error())
				log.Fatalf("an error occurred in worker %d; error: %s\n", i+1, err.Error())
			}

			if nRead != nWrite {
				log.Fatalf("an error occurred in worker %d; error: %s\n", i+1, err.Error())
			}

			start = int64(nWrite) + start
			if nWrite > 0 {
				written += int64(nWrite)
			}

			g.bars[i].Set64(written)

			percentage := int64(float32(written) / float32(size) * 100)

			_, flagged := flag[percentage]
			if !flagged {
				flag[percentage] = true
				g.bars[i].Prefix(fmt.Sprintf("worker %d %d%% ", i+1, percentage))
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				if size != written {
					log.Fatalf("an error occurred in worker %d; error: %s\n", i+1, err.Error())
				}
				break
			}
			log.Fatalf("an error occurred in worker %d; error: %s\n", i+1, err.Error())
		}
	}

}

func (g *Godol) getPartialBody(start, end int64) (io.ReadCloser, int64, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, g.url, nil)
	if err != nil {
		return nil, 0, err
	}

	rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end-1)
	req.Header.Add("Range", rangeHeader)
	res, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	size, err := strconv.ParseInt(res.Header["Content-Length"][0], 10, 64)
	return res.Body, size, err
}
