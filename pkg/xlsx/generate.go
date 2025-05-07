package xlsx

import (
	"fmt"
	"log"
	"os"

	"github.com/xuri/excelize/v2"
)

type ResponseData struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	FileURL string `json:"file_url,omitempty"`
}

func ProcessXlsx() ResponseData {
	f := excelize.NewFile()
	sheet := "Sheet1"

	// Write header
	headers := []string{"ID", "Name", "Email", "Phone", "Address", "Note"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Write mock data
	for row := 2; row <= 100000; row++ {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), row-1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("User %d", row-1))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("user%d@example.com", row-1))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("081-000-%04d", row-1))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), "123 Mockingbird Lane, Springfield, USA")
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), "This is a note field with some repeated text to increase file size.")
	}

	filename := "mock_data_5mb.xlsx"
	filepath := fmt.Sprintf("public/%s", filename)

	// Ensure the public directory exists
	if _, err := os.Stat("public"); os.IsNotExist(err) {
		if err := os.Mkdir("public", 0755); err != nil {
			log.Println("Failed to create public directory:", err)
			return ResponseData{false, "Server error", ""}
		}
	}

	// Save to file
	if err := f.SaveAs(filepath); err != nil {
		log.Println("Error saving file:", err)
		return ResponseData{false, "Download failed", ""}
	}

	apiUrl := "http://localhost:8080/"
	if apiUrl == "" {
		log.Println("API_URL is not set")
		return ResponseData{false, "API_URL not configured", ""}
	}

	return ResponseData{
		Status:  true,
		Message: "File generated successfully",
		FileURL: fmt.Sprintf("%spublic/%s", apiUrl, filename), // Construct the file URL
	}
}
