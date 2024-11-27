# Wallet With Single Threaded Event Loop - Proof of Concept

This proof of concept is intended to show you how single threaded event loop could solve concurrency problem that happen in a online wallet usecases.

## Requirements

- User can be created and have a wallet.
- User can top up their wallet
- User can transfer their money from their wallet to another wallet
- User can see their top 5 incoming and outgoing wallet mutations

## Concept

![Single Threaded Event Loop](https://github.com/user-attachments/assets/0661d101-2368-4be2-8b1d-dd802d4e1da3)

## Benchmark

**DB package benchmark**

```
Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^BenchmarkTransaction$ github.com/insomnius/wallet-event-loop/db

goos: linux
goarch: amd64
pkg: github.com/insomnius/wallet-event-loop/db
cpu: 12th Gen Intel(R) Core(TM) i5-12400F
BenchmarkTransaction-12    	   48580	     22459 ns/op	   40989 B/op	     182 allocs/op
PASS
ok  	github.com/insomnius/wallet-event-loop/db	1.345s
```

**Transfer process benchmark**

```
Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^BenchmarkTransfer$ github.com/insomnius/wallet-event-loop/agregation

goos: linux
goarch: amd64
pkg: github.com/insomnius/wallet-event-loop/agregation
cpu: 12th Gen Intel(R) Core(TM) i5-12400F
BenchmarkTransfer-12    	  296427	      3563 ns/op	   10936 B/op	      31 allocs/op
PASS
ok  	github.com/insomnius/wallet-event-loop/agregation	1.103s
```

## Test Results

![Test Results](https://private-user-images.githubusercontent.com/20650401/390493721-647d1ed2-4b91-4154-b63d-d76952daeb70.png?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTUiLCJleHAiOjE3MzI3MjUzMjEsIm5iZiI6MTczMjcyNTAyMSwicGF0aCI6Ii8yMDY1MDQwMS8zOTA0OTM3MjEtNjQ3ZDFlZDItNGI5MS00MTU0LWI2M2QtZDc2OTUyZGFlYjcwLnBuZz9YLUFtei1BbGdvcml0aG09QVdTNC1ITUFDLVNIQTI1NiZYLUFtei1DcmVkZW50aWFsPUFLSUFWQ09EWUxTQTUzUFFLNFpBJTJGMjAyNDExMjclMkZ1cy1lYXN0LTElMkZzMyUyRmF3czRfcmVxdWVzdCZYLUFtei1EYXRlPTIwMjQxMTI3VDE2MzAyMVomWC1BbXotRXhwaXJlcz0zMDAmWC1BbXotU2lnbmF0dXJlPTljMmNkZTE1YTYzZDI4Y2QyM2ZiYzE5ZDA4ZDIyOGY2ZDI4YjVkY2ZiOGQ1NzE3MGIyY2M5ODlhYzA2M2NmZjYmWC1BbXotU2lnbmVkSGVhZGVycz1ob3N0In0.je9mjDKsaVimJrJ8Fv5lhuBNDv2Km4hYAfqK-okY5NQ)