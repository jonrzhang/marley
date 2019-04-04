package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/jonrzhang/marley/scanfile/queue"
)

var path string

var wg sync.WaitGroup

var q = queue.NewQueue()

func checkFName(name string) {
	log.Printf("Name: %s", name)
}

func scanDir(dir string, fn chan string) {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for _, f := range files {
			// wg.Add(1)
			n := f.Name()
			fn <- n
			// fmt.Printf(n)
			q.Push(n)
		}
		close(fn)
	}()
}

func main() {

	// path = "/Users/zhangrong/Downloads/Images/megvii_feature"
	path = "/Users/zhangrong/Downloads/Images/Cropped"

	fn := make(chan string)
	scanDir(path, fn)

	for {
		// v := <-fn
		v, ok := <-fn
		if ok == false {
			break
		}

		// fmt.Println("Received ", v)
		fmt.Println("Received ", v, ok)

		if q.Size() > 0 {
			n := q.Pop()

			fmt.Println("File Name ", n)

		}
	}

	// num := 1
	// for _, f := range files {
	// wg.Add(1)
	// n := f.Name()
	// fname <- n
	// fmt.Printf(n)
	// q.Push(n)
	// num++
	// }

	// scanDir()

	// wg.Wait()

	// close(fname)

	// fmt.Println(fmt.Sprintf("Num: %d", num))
}
