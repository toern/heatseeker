# HeatSeeker — Go / Fyne Implementation

## Why Go?

Go was chosen as a reimplementation language for HeatSeeker for several reasons:

1. **Simple cross-compilation** — Go produces statically-linked binaries and its
   toolchain can cross-compile for any supported OS/architecture with a single
   `GOOS=windows go build` command — no additional toolchains required.

2. **Fast compilation** — The entire project compiles in seconds, making the
   development cycle much faster than C++ or Rust for a project of this size.

3. **Low learning curve** — Go's small specification means contributors can be
   productive quickly. The standard library covers binary parsing, image handling,
   and file I/O without external dependencies.

4. **Excellent concurrency** — If future work adds real-time camera streaming or
   batch processing, goroutines make concurrent work trivial.

## Why Fyne?

| Concern              | PyQt5               | Fyne (Go)                            |
|----------------------|---------------------|--------------------------------------|
| Licensing            | GPL / commercial    | BSD-3-Clause                         |
| Platforms            | Win/Mac/Linux       | Win/Mac/Linux/BSD/iOS/Android        |
| Binary size          | ~30 MB + Qt runtime | ~15 MB standalone                    |
| Cross-compilation    | Difficult           | `GOOS=... go build`                  |
| Installation         | `pip install`       | Single binary, zero dependencies     |
| Mobile support       | None                | Built-in (iOS, Android)              |

Fyne was chosen over alternatives because:
- **Native look and feel** on each platform via custom rendering
- **Material Design** inspired widgets that look modern out of the box
- **Active community** and regular releases
- **No CGo for core logic** — only the OpenGL/X11 backend uses CGo, and
  Fyne abstracts that away

## Architecture

```
go/
├── main.go       — GUI application built on Fyne
│                  • Main menu with file-open dialog
│                  • Side panel with temperature-range sliders
│                  • Central canvas rendering the thermal image
│                  • IRG header info display
│
├── irg.go        — Binary IRG format parser
│                  • Reads TC004 header fields
│                  • Determines data-start offset (0x80 vs 0x100)
│                  • Extracts grayscale and thermal (uint16 → Kelvin) data
│                  • Temperature conversion (Kelvin → Fahrenheit)
│
└── irg_test.go   — Unit tests
                   • Colormap boundary & clamping tests
                   • Temperature conversion accuracy tests
                   • Min/max detection tests
                   • Image rendering tests
```

## Building and Running

```bash
# Build
go build -o heatseeker .

# Run
go run .

# Cross-compile for Windows
GOOS=windows GOARCH=amd64 go build -o heatseeker.exe .

# Run tests
go test ./... -v
```

## Dependencies

| Module             | Purpose                                 |
|--------------------|-----------------------------------------|
| fyne.io/fyne/v2   | Cross-platform GUI framework            |
| (stdlib)           | Binary parsing, image, file I/O         |
