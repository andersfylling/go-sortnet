module github.com/andersfylling/go-sortnet

go 1.19

require (
	github.com/cheggaaa/pb/v3 v3.1.0
	github.com/kelindar/bitmap v1.4.1
	github.com/kyroy/kdtree v0.0.0-20200419114247-70830f883f1d
	golang.org/x/sync v0.1.0
)

require (
	github.com/VividCortex/ewma v1.1.1 // indirect
	github.com/fatih/color v1.10.0 // indirect
	github.com/kelindar/simd v1.1.2 // indirect
	github.com/klauspost/cpuid/v2 v2.0.12 // indirect
	github.com/kyroy/priority-queue v0.0.0-20180327160706-6e21825e7e0c // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	golang.org/x/sys v0.0.0-20220503163025-988cb79eb6c6 // indirect
)

//replace github.com/kyroy/kdtree v0.0.0-20200419114247-70830f883f1d => ../kdtree
