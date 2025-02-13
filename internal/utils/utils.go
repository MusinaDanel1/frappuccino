package utils

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
)

var (
	Port = flag.String("port", "8080", "Port to listen on")
	Help = flag.Bool("help", false, "Show help Definition")
)

const (
	PINK  = "\033[38;5;206m"
	TEAL  = "\033[38;5;32m"
	RESET = "\033[0m"
)

func SetupDirectory(dir string) error {
	if dir == "internal" || dir == "models" || dir == "cmd" || dir == "..." || dir == ".." {
		return errors.ErrUnsupported
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create di: %v", err)
		}
	}

	createFileIfNotExist(fmt.Sprintf("%s/menu_items.json", dir), "[]")
	createFileIfNotExist(fmt.Sprintf("%s/inventory.json", dir), "[]")
	createFileIfNotExist(fmt.Sprintf("%s/orders.json", dir), "[]")
	return nil
}

func createFileIfNotExist(filePath, defaultContent string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Failed fo create a file %s: %v\n", filePath, err)
		}
		defer file.Close()

		_, err = file.WriteString(defaultContent)
		if err != nil {
			fmt.Printf("failed to write into file %s: %v\n", filePath, err)
		}
	}
}

func CheckPort(port string) bool {
	portNum, err := strconv.Atoi(port)
	return err == nil && portNum > 1024 && portNum <= 65535
}

// ParsePaginationParams парсит параметры page и pageSize, задавая значения по умолчанию
func ParsePaginationParams(queryParams url.Values) (int, int) {
	page, err := strconv.Atoi(queryParams.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(queryParams.Get("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	return page, pageSize
}

func PrintHelp() {
	fmt.Println("$ ./frappuccino --help" +
		"\nCoffee Shop Management System" +
		"\n" +
		"\nUsage:" +
		"\n  frappuccino [--port <N>] [--dir <S>]" +
		"\n  frappuccino --help" +
		"\n" +
		"\nOptions:" +
		"\n  --help       Show this screen." +
		"\n  --port N     Port number." +
		"  --dir S      Path to the data directory.")
}
