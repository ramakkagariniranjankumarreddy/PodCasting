package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Function to send HTTP request
func registerRequest(url string, wg *sync.WaitGroup, clientNum int, username string) {
	url = url + "/register?username=" + username + "&password=" + username
	defer wg.Done()
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Client %d: Error making request: %v\n", clientNum, err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Client %d: Received %s\n", clientNum, resp.Status)
}

func testRegister(url string) {
	user := [5]string{"test1", "test2", "test3", "test4", "test5"}
	var wg sync.WaitGroup
	for i := 0; i < len(user); i++ {
		wg.Add(1)
		go registerRequest(url, &wg, i, user[i])
	}
	wg.Wait()

}

func main() {

	url := "http://192.168.29.153:8080"
	testRegister(url)

	fmt.Println("All requests completed!")
}
