package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	flagOverwrite = flag.Bool("overwrite", false, "Overwrite the CSV file with formatted CSV")
	flagJSON      = flag.Bool("json", false, "Output JSON")
	flagMarkdown  = flag.Bool("markdown", false, "Output Markdown table")
)

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func readCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	// Trim spaces in every cell
	for i := range rows {
		for j := range rows[i] {
			rows[i][j] = strings.TrimSpace(rows[i][j])
		}
	}

	return rows, nil
}

func writeCSV(path string, rows [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	err = w.WriteAll(rows)
	if err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func formatPretty(rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}
	cols := len(rows[0])
	widths := make([]int, cols)

	for _, row := range rows {
		for i := 0; i < cols; i++ {
			if i < len(row) && len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	var b strings.Builder
	for _, row := range rows {
		for i := 0; i < cols; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			fmt.Fprintf(&b, "%-*s", widths[i], cell)
			if i != cols-1 {
				b.WriteString("   ")
			}
		}
		b.WriteByte('\n')
	}

	return b.String()
}

func formatMarkdown(rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}
	cols := len(rows[0])
	widths := make([]int, cols)

	for _, row := range rows {
		for i := 0; i < cols; i++ {
			if i < len(row) && len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	var b strings.Builder

	// header
	for i, cell := range rows[0] {
		fmt.Fprintf(&b, "| %-*s ", widths[i], cell)
	}
	b.WriteString("|\n")

	// separator
	for i := 0; i < cols; i++ {
		b.WriteString("| ")
		b.WriteString(strings.Repeat("-", widths[i]))
		b.WriteString(" ")
	}
	b.WriteString("|\n")

	// body
	for _, row := range rows[1:] {
		for i, cell := range row {
			fmt.Fprintf(&b, "| %-*s ", widths[i], cell)
		}
		b.WriteString("|\n")
	}

	return b.String()
}

func jsonObjects(rows [][]string) ([]map[string]string, error) {
	if len(rows) == 0 {
		return nil, nil
	}

	headers := rows[0]
	var result []map[string]string

	for _, row := range rows[1:] {
		obj := make(map[string]string)
		for i, key := range headers {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			obj[key] = val
		}
		result = append(result, obj)
	}

	return result, nil
}

func main() {
	flag.Parse()
	files := flag.Args()

	if len(files) == 0 {
		fmt.Println("Usage: csvfmt [--overwrite] [--json] [--markdown] file1.csv [...]")
		os.Exit(1)
	}

	for _, path := range files {
		rows, err := readCSV(path)
		must(err)

		// JSON: array of objects using header as keys
		if *flagJSON {
			objs, err := jsonObjects(rows)
			must(err)
			enc, _ := json.MarshalIndent(objs, "", "  ")
			fmt.Println(string(enc))
			continue
		}

		// Markdown
		if *flagMarkdown {
			fmt.Println(formatMarkdown(rows))
			continue
		}

		// Pretty
		out := formatPretty(rows)

		// Overwrite CSV (still valid CSV, just trimmed)
		if *flagOverwrite {
			err := writeCSV(path, rows)
			must(err)
		}

		fmt.Print(out)
	}
}
