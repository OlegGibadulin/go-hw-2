package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	wg.Add(len(jobs))
	in := make(chan interface{})
	defer close(in)
	for _, jb := range jobs {
		out := make(chan interface{})
		go func(in, out chan interface{}, jb job) {
			defer wg.Done()
			defer close(out)
			jb(in, out)
		}(in, out, jb)
		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for i := range in {
		wg.Add(1)
		data := strconv.Itoa(i.(int))
		md5 := DataSignerMd5(data)
		go func() {
			defer wg.Done()
			wg := &sync.WaitGroup{}
			wg.Add(2)
			var leftHash, rightHash string
			go func() {
				defer wg.Done()
				leftHash = DataSignerCrc32(data)
			}()
			go func() {
				defer wg.Done()
				rightHash = DataSignerCrc32(md5)
			}()
			wg.Wait()
			res := leftHash + "~" + rightHash
			out <- res
		}()
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for i := range in {
		wg.Add(1)
		data := i.(string)
		go func() {
			defer wg.Done()
			wg := &sync.WaitGroup{}
			wg.Add(6)
			hashs := make([]string, 6)
			for th := 0; th < 6; th++ {
				go func(th int) {
					defer wg.Done()
					thData := strconv.Itoa(th) + data
					hashs[th] = DataSignerCrc32(thData)
				}(th)
			}
			wg.Wait()
			res := strings.Join(hashs, "")
			out <- res
		}()
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var results []string
	for i := range in {
		results = append(results, i.(string))
	}
	sort.Strings(results)
	res := strings.Join(results, "_")
	out <- res
}
