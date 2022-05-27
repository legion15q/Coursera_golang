package main

import (
	"sort"
	"strconv"
	"sync"
)

//var start = time.Now()

func ExecutePipeline(hashSignJobs ...job) {
	in := make(chan interface{})
	defer close(in)
	wg := new(sync.WaitGroup)
	for _, job_func := range hashSignJobs {
		wg.Add(1)
		out := make(chan interface{})
		go func(in, out chan interface{}, fn job) {
			defer close(out)
			defer wg.Done()
			fn(in, out)
		}(in, out, job_func)
		in = out //Здесь in является указателем (как и out) на канал out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := new(sync.WaitGroup)
	for val := range in {
		wg.Add(1)
		//fmt.Println("начало цикла")
		val_str := strconv.Itoa(val.(int))
		md_5 := DataSignerMd5(val_str)
		go func(out_chan chan interface{}, val_str string, md_5 string) {
			defer wg.Done()
			var crc_32, crc_32_md_5 string
			//fmt.Println("горутина заблочена")
			wg_local := new(sync.WaitGroup)
			wg_local.Add(2)
			go func() {
				defer wg_local.Done()
				crc_32 = DataSignerCrc32(val_str)
			}()
			go func() {
				defer wg_local.Done()
				crc_32_md_5 = DataSignerCrc32(md_5)
			}()
			wg_local.Wait()
			result := crc_32 + "~" + crc_32_md_5
			//fmt.Println("получен результат =", result, "за", time.Since(start))
			out_chan <- result

		}(out, val_str, md_5)
		//fmt.Println("конец цикла")
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := new(sync.WaitGroup)
	for val := range in {
		wg.Add(1)
		go func(out_chan chan interface{}, val_ string) {
			defer wg.Done()
			map_ := make(map[int]string, 6)
			mu := new(sync.Mutex)
			wg_local := new(sync.WaitGroup)
			for i := 0; i < 6; i++ {
				wg_local.Add(1)
				go func(index int) {
					defer wg_local.Done()
					res := DataSignerCrc32(strconv.Itoa(index) + val_)
					mu.Lock()
					map_[index] = res
					mu.Unlock()
				}(i)
			}
			wg_local.Wait()
			var result string
			for i := 0; i < 6; i++ {
				result += map_[i]
			}
			//fmt.Println(result)
			out <- result
		}(out, val.(string))
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var result string
	var arr sorted_arr
	for val := range in {
		//fmt.Println(val)
		arr = append(arr, val.(string))
	}
	sort.Sort(arr)
	for i := 0; i < len(arr)-1; i++ {
		result = result + string(arr[i]) + "_"
	}
	result += arr[arr.Len()-1]
	out <- result
}

type sorted_arr []string

func (c sorted_arr) Len() int {
	return len(c)
}

func (c sorted_arr) Less(i, j int) bool {
	return c[i] < c[j]
}

func (c sorted_arr) Swap(i, j int) {
	temp := c[j]
	c[j] = c[i]
	c[i] = temp
}
