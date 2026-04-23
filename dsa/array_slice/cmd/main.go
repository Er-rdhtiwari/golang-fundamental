package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

func main() {
	// Read input like: -nums=5,9,2,5,9
	numsArgs := flag.String("nums", "", "comma-seprated integers required")
	flag.Parse()

	if *numsArgs == "" {
		fmt.Println("please pass -nums")
	}

	// Split the single CLI string into smaller string parts.
	// Example: "5,9,2,5,9" -> ["5", "9", "2", "5", "9"]
	parts := strings.Split(*numsArgs, ",")

	// Create an empty integer slice with enough room for all values.
	nums := make([]int, 0, len(parts))

	for _, p := range parts {
		// Remove extra spaces, then convert each string into an integer.
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			fmt.Printf("invalid number: %q\n", p)
			return
		}

		// Add the converted number into the nums slice.
		nums = append(nums, n)
	}

	// Sum all numbers from the parsed input.
	sum := 0

	for _, num := range nums {
		sum = sum + num
	}

	fmt.Printf(" sum: %+v\n", sum)
}
