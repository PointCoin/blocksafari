// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PointCoin/btcjson"
)

var (
	templates = template.Must(template.ParseGlob("includes/*.html"))
)

type displayBlockPage struct {
	Bits            string
	Difficulty      string
	Hash            string
	Height          int64
	MerkleRoot      string
	NextHash        string
	Nonce           uint32
	PreviousHash    string
	Size            string
	Timestamp       string
	Txs             []blockPageTx
	CoinbaseMessage string
}

type displayMainPage struct {
	Blocks       []blockPager
	PreviousPage int64
	NextPage     int64
}

type blockPager struct {
	DisplayHash     string
	Hash            string
	Height          int64
	Size            string
	Timestamp       string
	Txs             int
	TotalBTC        string
	CoinbaseMessage string
}

type ScoreItem struct {
	Name   string
	Number int
}

type displayTxPage struct {
	Hash   string
	Vin    []btcjson.Vin
	Vout   []btcjson.Vout
	BtcOut string
}

// ErrMsg struct to hold the string from an error for display.
type ErrMsg struct {
	ErrMsg string
}

type blockPageTx struct {
	DisplayHash string
	Hash        string
	Vin         []btcjson.Vin
	Vout        []btcjson.Vout
}

func printBlock(w http.ResponseWriter, block *btcjson.BlockResult, trans []btcjson.TxRawResult) {
	tmpTime := time.Unix(block.Time, 0)
	txs := make([]blockPageTx, len(trans))
	for i, tran := range trans {
		txs[i] = blockPageTx{
			DisplayHash: fmt.Sprintf("%s", tran.Txid)[:10],
			Hash:        tran.Txid,
			Vin:         tran.Vin,
			Vout:        tran.Vout,
		}
	}

	msg := getCoinbaseMsg(block.RawTx[0])

	b := &displayBlockPage{
		Bits:            block.Bits,
		Difficulty:      fmt.Sprintf("%f", block.Difficulty),
		Hash:            block.Hash,
		Height:          block.Height,
		MerkleRoot:      block.MerkleRoot,
		NextHash:        block.NextHash,
		Nonce:           block.Nonce,
		PreviousHash:    block.PreviousHash,
		Size:            fmt.Sprintf("%0.3f", float64(block.Size)/1000.00),
		Timestamp:       fmt.Sprintf("%s", tmpTime.String()[:19]),
		Txs:             txs,
		CoinbaseMessage: msg,
	}
	err := templates.ExecuteTemplate(w, "block.html", b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func printContentType(w http.ResponseWriter, ctype string) {
	w.Header().Set("Content-type", ctype)
}

func printErrorPage(w http.ResponseWriter, errstr string) {
	e := &ErrMsg{
		ErrMsg: errstr,
	}

	printHTMLHeader(w, "Error")
	err := templates.ExecuteTemplate(w, "error.html", e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	printHTMLFooter(w)
}

func printHTMLFooter(w http.ResponseWriter) {
	err := templates.ExecuteTemplate(w, "footer.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func printHTMLHeader(w http.ResponseWriter, title string) {
	printContentType(w, "text/html")

	err := templates.ExecuteTemplate(w, "header.html", title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getCoinbaseMsg(coinbaseTx btcjson.TxRawResult) string {
	b, err := hex.DecodeString(coinbaseTx.Vin[0].Coinbase)
	if err != nil {
		b = []byte("ERROR")
	}

	var msg string
	// Set the CoinbaseMessage
	if len(b) > 9 {
		msg = string(b[9:])
	} else {
		msg = string(b)
	}
	return msg
}

type ByNum []ScoreItem

func (this ByNum) Len() int {
	return len(this)
}
func (this ByNum) Less(i, j int) bool {
	return this[i].Number > this[j].Number
}
func (this ByNum) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func printScores(w http.ResponseWriter, blocks []*btcjson.BlockResult) {
	var ScoreList []ScoreItem
	for _, block := range blocks {

		// Get the coinbase transaction
		flag := 0
		msg := getCoinbaseMsg(block.RawTx[0])
		if !strconv.IsPrint(rune(msg[0])) {
			msg = msg[1:]
		}
		for i, entry := range ScoreList {
			if entry.Name == msg {
				ScoreList[i].Number++
				flag = 1
			}
		}
		if flag == 0 {
			ScoreList = append(ScoreList, ScoreItem{Name: msg, Number: 1})
		}
	}

	sort.Sort(ByNum(ScoreList))

	err := templates.ExecuteTemplate(w, "scores.html", ScoreList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func produceBlockPager(blocks []*btcjson.BlockResult) []blockPager {
	display := make([]blockPager, len(blocks))
	for i, block := range blocks {
		var totalBtc float64

		// Get the coinbase transaction
		msg := getCoinbaseMsg(block.RawTx[0])

		for _, tx := range block.RawTx {

			for _, v := range tx.Vout {

				totalBtc += v.Value
			}
		}
		tmpTime := time.Unix(block.Time, 0)

		display[i] = blockPager{
			DisplayHash:     fmt.Sprintf("%s", strings.TrimLeft(block.Hash, "0"))[:10],
			Hash:            block.Hash,
			Height:          block.Height,
			Size:            fmt.Sprintf("%0.3f", float64(block.Size)/1000.00),
			Timestamp:       fmt.Sprintf("%s", tmpTime.String()[:19]),
			Txs:             len(block.RawTx),
			TotalBTC:        fmt.Sprintf("%.5f", totalBtc),
			CoinbaseMessage: msg,
		}
	}
	return display
}

func printBlockPager(w http.ResponseWriter, blocks []*btcjson.BlockResult) {

	display := produceBlockPager(blocks)
	page := displayMainPage{
		Blocks:       display,
		PreviousPage: display[len(display)-1].Height,
		NextPage:     display[0].Height + 29,
	}
	err := templates.ExecuteTemplate(w, "pager.html", page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func printMainBlock(w http.ResponseWriter, blocks []*btcjson.BlockResult) {

	display := produceBlockPager(blocks)
	page := displayMainPage{
		Blocks:       display,
		PreviousPage: display[len(display)-1].Height,
	}
	err := templates.ExecuteTemplate(w, "mainblock.html", page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func printTx(w http.ResponseWriter, tx *btcjson.TxRawResult) {
	var totalBtc float64
	for _, v := range tx.Vout {
		totalBtc += v.Value
	}
	display := &displayTxPage{
		Hash:   tx.Txid,
		Vin:    tx.Vin,
		Vout:   tx.Vout,
		BtcOut: fmt.Sprintf("%.8f", totalBtc),
	}
	err := templates.ExecuteTemplate(w, "tx.html", display)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
