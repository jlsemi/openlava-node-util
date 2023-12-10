package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	n := rand.Intn(10)
	fmt.Printf("worker sleep %v seconds\n", n)
	time.Sleep((time.Duration(n) + 1) * time.Second)
}
