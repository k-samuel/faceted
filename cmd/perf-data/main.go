package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	outputDir = "./tests/data"
	dataFile  = "data.json"
)

var (
	colors = []string{"red", "green", "blue", "yellow", "black", "white"}
	brands = []string{
		"Nike",
		"H&M",
		"Zara",
		"Adidas",
		"Louis Vuitton",
		"Cartier",
		"Hermes",
		"Gucci",
		"Uniqlo",
		"Rolex",
		"Coach",
		"Victoria's Secret",
		"Chow Tai Fook",
		"Tiffany & Co.",
		"Burberry",
		"Christian Dior",
		"Polo Ralph Lauren",
		"Prada",
		"Under Armour",
		"Armani",
		"Puma",
		"Ray-Ban",
	}
	warehouses = []int{1, 10, 23, 345, 43, 5476, 34, 675, 34, 24, 789, 45, 65, 34, 54, 511, 512, 520}
	types      = []string{"normal", "middle", "good"}
)

func main() {

	sz := flag.Int("size", 100000, "dataset size")
	flag.Parse()
	resultSize := *sz

	rand.Seed(time.Now().UnixNano())

	info, err := os.Stat(fmt.Sprintf("%s", outputDir))
	if os.IsNotExist(err) || !info.IsDir() {
		panic(fmt.Sprintf("%s", outputDir) + " is not exists")
	}

	filePath := fmt.Sprintf("%s/%d/%s", outputDir, resultSize, dataFile)

	// Create directory if not exists
	dir := fmt.Sprintf("%s/%d", outputDir, resultSize)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	// Open file for writing
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Printf("Generating %d records to %s...\n", resultSize, filePath)

	startTime := time.Now()

	// Generate records
	for i := 1; i <= resultSize; i++ {
		// Random warehouse count (0 to len(warehouses))
		countWh := rand.Intn(len(warehouses) + 1)

		whSet := make(map[int]bool)
		// Get unique warehouses
		for j := 0; j < countWh; j++ {
			whSet[warehouses[rand.Intn(len(warehouses))]] = true
		}

		wh := make([]int, 0, len(whSet))
		for w := range whSet {
			wh = append(wh, w)
		}

		rec := map[string]interface{}{
			"id":         i,
			"color":      colors[rand.Intn(len(colors))],
			"back_color": colors[rand.Intn(len(colors))],
			"size":       rand.Intn(50-34+1) + 34,
			"brand":      brands[rand.Intn(len(brands))],
			"price":      rand.Intn(10000-1000+1) + 1000,
			"discount":   rand.Intn(11),
			"combined":   rand.Intn(2),
			"quantity":   rand.Intn(101),
			"warehouse":  wh,
			"type":       types[rand.Intn(len(types))],
		}

		// Write to file as JSON line
		line := fmt.Sprintf(`{"id":%d,"color":"%s","back_color":"%s","size":%d,"brand":"%s","price":%d,"discount":%d,"combined":%d,"quantity":%d,"warehouse":[%s],"type":"%s"}`,
			rec["id"].(int),
			rec["color"].(string),
			rec["back_color"].(string),
			rec["size"].(int),
			rec["brand"].(string),
			rec["price"].(int),
			rec["discount"].(int),
			rec["combined"].(int),
			rec["quantity"].(int),
			formatWarehouse(wh),
			rec["type"].(string))
		_, err := file.WriteString(line + "\n")
		if err != nil {
			fmt.Printf("Error writing record %d: %v\n", i, err)
			return
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("Generated %d records in %.3f seconds\n", resultSize, elapsed.Seconds())
}

func formatWarehouse(wh []int) string {
	if len(wh) == 0 {
		return ""
	}
	result := fmt.Sprintf("%d", wh[0])
	for i := 1; i < len(wh); i++ {
		result += fmt.Sprintf(",%d", wh[i])
	}
	return result
}
