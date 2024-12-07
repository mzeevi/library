package data

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
	"os"
	"strings"
	"sync"
)

const (
	errCSVWriterNotInitialized = "CSV writer is not initialized"
)

const (
	CSVOutputFormat   = TransactionOutputType("csv")
	EXLAMOutputFormat = TransactionOutputType("xlam")
	XLSMOutputFormat  = TransactionOutputType("xlsm")
	XLSXOutputFormat  = TransactionOutputType("xlsx")
	XLTMOutputFormat  = TransactionOutputType("xltm")
	XLTXOutputFormat  = TransactionOutputType("xltx")
)

const (
	excelSheetName   = "transactions"
	nonExistentSheet = -1
)

type TransactionOutputType string

type TransactionsOutput interface {
	// CreateWriter initializes a writer and creates or opens the specified file.
	CreateWriter(filename, format string) error

	// WriteRecord writes a single record to an output file.
	WriteRecord(record []string) error

	// CloseWriter closes the file writer.
	CloseWriter() error
}

type CSVTransactionOutput struct {
	file   *os.File
	mutex  *sync.Mutex
	writer *csv.Writer
}

type ExcelTransactionOutput struct {
	f         *excelize.File
	filename  string
	sheetName string
}

// addFormatSuffix ensures the given filename ends with the appropriate file format suffix.
// If the suffix is not present, it appends the format as a suffix to the filename.
func addFormatSuffix(filename, format string) string {
	if !strings.HasSuffix(filename, format) {
		return fmt.Sprintf("%s.%s", filename, format)
	}

	return ""
}

func (c *CSVTransactionOutput) CreateWriter(filename, format string) error {
	f, err := os.Create(addFormatSuffix(filename, format))
	if err != nil {
		return err
	}

	c.file = f
	c.writer = csv.NewWriter(f)
	c.mutex = &sync.Mutex{}

	return nil
}

func (c *CSVTransactionOutput) WriteRecord(record []string) error {
	if c.writer == nil {
		return errors.New(errCSVWriterNotInitialized)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.writer.Write(record)
}

func (c *CSVTransactionOutput) CloseWriter() error {
	if c.writer != nil {
		c.mutex.Lock()
		c.writer.Flush()
		c.mutex.Unlock()
	}

	if c.file != nil {
		return c.file.Close()
	}
	return nil
}

func (c *ExcelTransactionOutput) CreateWriter(filename, format string) error {
	var err error

	normalizedFilename := addFormatSuffix(filename, format)
	if c.f, err = excelize.OpenFile(normalizedFilename); err != nil {
		c.f = excelize.NewFile()
	}

	i, err := c.f.GetSheetIndex(excelSheetName)
	if err != nil {
		return err
	}

	if i == nonExistentSheet {
		_, err = c.f.NewSheet(excelSheetName)
		if err != nil {
			return err
		}
	}

	c.sheetName = excelSheetName
	c.filename = normalizedFilename

	return nil
}

func (c *ExcelTransactionOutput) WriteRecord(record []string) error {
	rows, err := c.f.GetRows(c.sheetName)
	if err != nil {
		return err
	}
	nextRow := len(rows) + 1

	for colIndex, value := range record {
		cell := fmt.Sprintf("%s%d", string(rune('A'+colIndex)), nextRow)
		if err = c.f.SetCellValue(c.sheetName, cell, value); err != nil {
			return err
		}
	}

	if err = c.f.SaveAs(c.filename); err != nil {
		return err
	}

	return nil
}

func (c *ExcelTransactionOutput) CloseWriter() error {
	if err := c.f.Close(); err != nil {
		return err
	}

	return nil
}
