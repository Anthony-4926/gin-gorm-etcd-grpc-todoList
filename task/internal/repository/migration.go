package repository

import "fmt"

func Migration() {
	err := DB.Set("gorm:table_options", "charset=utf8mb4").
		AutoMigrate(&Task{})
	if err != nil {
		fmt.Println("migration err", err)
	}
}
