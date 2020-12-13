package main

import (
	"fmt"
)

func squares(ch chan int) {
	for i := 0; i <= 9; i++ {
		fmt.Println("Sending Data")
		ch <- i * i
	}

	close(ch)
}

// func main() {
// 	ch := make(chan int, 10)

// 	go squares(ch)

// 	for val := range ch {
// 		fmt.Println("Receiving Data", val)
// 	}

// 	fmt.Println("Done")
// }
