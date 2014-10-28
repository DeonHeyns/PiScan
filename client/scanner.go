// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// This is a fully-functional (but simple) PiScanner application. So far, all
// it does is define a function which takes the scanned barcode result and
// prints it to stdout. But in time, this binary will grow to do much more...

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/Banrai/PiScan/scanner"
	"github.com/Banrai/PiScan/server/commerce/amazon"
	"github.com/Banrai/PiScan/server/database/barcodes"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
)

const ( // barcodes db lookups will be via internal api to a remote server, eventually
	barcodeDBUser   = "pod"
	barcodeDBServer = "127.0.0.1"
	barcodeDBPort   = "3306"
)

func main() {
	var device, dbUser, dbHost, dbPort string
	flag.StringVar(&device, "device", scanner.SCANNER_DEVICE, fmt.Sprintf("The '/dev/input/event' device associated with your scanner (defaults to '%s')", scanner.SCANNER_DEVICE))
	flag.StringVar(&dbUser, "dbUser", barcodeDBUser, fmt.Sprintf("The barcodes database user (defaults to '%s')", barcodeDBUser))
	flag.StringVar(&dbHost, "dbHost", barcodeDBServer, fmt.Sprintf("The barcodes database server (defaults to '%s')", barcodeDBServer))
	flag.StringVar(&dbPort, "dbPort", barcodeDBPort, fmt.Sprintf("The barcodes database port (defaults to '%s')", barcodeDBPort))
	flag.Parse()

	// eventually, make all this via an internal api call to a remote server, but for now...
	// connect to the barcodes database and make all the prepared statements available to scanner functions
	db, err := sql.Open("mysql",
		strings.Join([]string{dbUser, "@tcp(", dbHost, ":", dbPort, ")/product_open_data"}, ""))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	gtin, err := db.Prepare(barcodes.GTIN_LOOKUP)
	if err != nil {
		log.Fatal(err)
	}
	defer gtin.Close()

	brand, err := db.Prepare(barcodes.BRAND_LOOKUP)
	if err != nil {
		log.Fatal(err)
	}
	defer brand.Close()

	asinLookup, err := db.Prepare(barcodes.ASIN_LOOKUP)
	if err != nil {
		log.Fatal(err)
	}
	defer asinLookup.Close()

	asinInsert, err := db.Prepare(barcodes.ASIN_INSERT)
	if err != nil {
		log.Fatal(err)
	}
	defer asinInsert.Close()

	printFn := func(barcode string) {
		// print the barcode returned by the scanner to stdout
		fmt.Println(fmt.Sprintf("barcode: %s", barcode))

		// and, as a glimpse into the future...
		// lookup the barcode on Amazon's API
		// and print the (json) result to stdout
		// (in the future, this will be handled more elegantly/correctly)
		js, err := amazon.Lookup(barcode, asinLookup, asinInsert) // use both the barcodes database, and the Amazon API
		if err != nil {
			fmt.Println(fmt.Sprintf("Amazon lookup error: %s", err))
		} else {
			fmt.Println(fmt.Sprintf("Amazon result: %s", js))
		}
	}
	errorFn := func(e error) {
		log.Fatal(e)
	}
	scanner.ScanForever(device, printFn, errorFn)
}