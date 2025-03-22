//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"syscall/js"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jung-kurt/gofpdf"
)

func redirectTo(link string) {
	_, err := url.ParseRequestURI(link)
	if err != nil {
		return
	}

	location := js.Global().Get("window").Get("location")
	if !location.IsNull() && !location.IsUndefined() {
		location.Set(
			"href",
			link,
		)
	}
}

func toSHA512(data string) string {
	hash := sha512.New()
	hash.Write([]byte(data))

	return hex.EncodeToString(hash.Sum(nil))
}

func numberWithCommas(f float64) string {
	s := fmt.Sprintf("%.2f", f)
	parts := strings.Split(s, ".")

	intPart := parts[0]
	decPart := parts[1]

	var result strings.Builder
	n := len(intPart)

	for i, c := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteRune(',')
		}

		result.WriteRune(c)
	}

	return result.String() + "." + decPart
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}

	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func generateTransactionPDF(transactions []Transaction) {
	usernameObject := document.Call(
		"getElementById",
		"card-holder",
	)

	if usernameObject.IsNull() || usernameObject.IsUndefined() {
		return
	}

	username := usernameObject.Get("innerHTML").String()
	pdf := gofpdf.New("P", "mm", "A4", "")
	pageCount := 1

	pdf.SetMargins(10, 10, 10)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "I", 8)
		now := time.Now().Format("02/01/2006 15:04:05 MST")
		pdf.CellFormat(
			0, 10,
			"Generated on "+now+". Page "+strconv.Itoa(pageCount)+".",
			"", 0,
			"C", false,
			0, "",
		)
	})
	pdf.AddPage()

	pageWidth, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	availableWidth := pageWidth - left - right

	colWidthID := availableWidth * 0.20
	colWidthAmount := availableWidth * 0.15
	colWidthCategory := availableWidth * 0.20
	colWidthCreatedAt := availableWidth * 0.30
	colWidthProcessed := availableWidth * 0.15

	headers := []string{
		"Transaction ID",
		"Amount",
		"Category",
		"Timestamp",
		"Processed",
	}

	colWidths := []float64{
		colWidthID,
		colWidthAmount,
		colWidthCategory,
		colWidthCreatedAt,
		colWidthProcessed,
	}

	printTableHeader := func() {
		pdf.SetFont("Helvetica", "B", 14)
		pdf.CellFormat(0, 10, "Ura Transaction Slip", "", 1, "C", false, 0, "")

		pdf.SetFont("Helvetica", "", 10)
		pdf.CellFormat(
			0, 4,
			"This official Ura transaction slip was generated for "+username+".",
			"", 1, "C",
			false, 0, "",
		)
		pdf.Ln(5)

		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetFillColor(200, 200, 200)

		for i, header := range headers {
			pdf.CellFormat(colWidths[i], 6, header, "1", 0, "C", true, 0, "")
		}

		pdf.Ln(-1)
		pdf.SetFont("Helvetica", "", 12)
	}
	printTableHeader()

	rowCount := 0
	for _, t := range transactions {
		if rowCount == 40 {
			pdf.AddPage()
			printTableHeader()

			rowCount = 0
			pageCount++
		}

		var fill bool
		if rowCount%2 == 1 {
			pdf.SetFillColor(240, 240, 240)
			fill = true
		} else {
			pdf.SetFillColor(255, 255, 255)
			fill = false
		}

		tid := t.TransactionID
		if len(tid) > 12 {
			tid = tid[:12]
		}

		var formattedCreatedAt string
		if parsedTime, err := time.Parse(time.RFC3339, t.CreatedAt); err == nil {
			formattedCreatedAt = parsedTime.Format("02/01/2006 15:04:05 MST")
		} else {
			formattedCreatedAt = t.CreatedAt
		}

		processedSymbol := "True"
		if t.Processed == 0 {
			processedSymbol = "False"
		}

		pdf.CellFormat(colWidths[0], 6, tid, "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[1], 6, fmt.Sprintf("%.2f", t.Amount), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[2], 6, capitalizeFirst(t.Category), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[3], 6, formattedCreatedAt, "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[4], 6, processedSymbol, "1", 0, "C", fill, 0, "")
		pdf.Ln(-1)

		rowCount++
	}

	pdf.SetTitle("Ura Transaction Slip", false)
	pdf.SetSubject("Official transaction slip", false)
	pdf.SetAuthor(username, false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return
	}

	jsUint8Array := js.Global().Get("Uint8Array")
	if jsUint8Array.IsNull() || jsUint8Array.IsUndefined() {
		return
	}

	data := buf.Bytes()
	uint8Array := jsUint8Array.New(len(data))
	js.CopyBytesToJS(uint8Array, data)

	blob := js.Global().Get("Blob").New(
		[]interface{}{uint8Array},
		map[string]interface{}{
			"type": "application/pdf",
		},
	)

	js.Global().Get("window").Call(
		"open",
		js.Global().Get("URL").Call("createObjectURL", blob),
	)
}
