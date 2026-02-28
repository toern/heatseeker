# HeatSeeker — Rust / egui Implementation

## Why Rust?

Rust was chosen as a reimplementation language for HeatSeeker for several reasons:

1. **Cross-platform by design** — Rust compiles natively on Windows, macOS, and Linux with
   no runtime dependencies. A single `cargo build` produces a self-contained binary.

2. **Performance** — Parsing binary IRG files and applying colormaps over large pixel arrays
   benefits from Rust's zero-cost abstractions. There is no garbage collector pause or
   interpreter overhead, which keeps the interactive viewer responsive even on large thermal
   images.

3. **Memory safety** — Rust's ownership model prevents the classes of memory bugs (buffer
   overflows, use-after-free) that commonly occur when manually decoding binary formats.

4. **Excellent ecosystem** — The `egui` / `eframe` GUI framework provides immediate-mode
   rendering that works on Windows, macOS, Linux, and even compiles to WebAssembly for
   browser deployment — a strict superset of the platforms PyQt5 supports.

## Why egui / eframe instead of Qt bindings?

| Concern              | PyQt5               | egui / eframe                        |
|----------------------|---------------------|--------------------------------------|
| Licensing            | GPL / commercial    | MIT / Apache-2.0                     |
| Binary size          | ~30 MB + Qt runtime | ~5–10 MB standalone                  |
| Cross-compilation    | Difficult           | Straightforward with `cross`         |
| WebAssembly support  | None                | Built-in (`eframe` compiles to wasm) |
| Installation         | `pip install`       | Single binary, zero dependencies     |

## Architecture

```
src/
├── main.rs   — GUI application built on eframe/egui
│              • Menu bar with file-open dialog (via `rfd`)
│              • Side panel with temperature-range sliders
│              • Central panel rendering thermal image as a texture
│              • Hover-to-inspect temperature readout
│
└── irg.rs    — Binary IRG format parser
               • Reads the TC004 header (emissivity, temperatures, distances)
               • Determines data-start offset (0x80 vs 0x100)
               • Extracts grayscale, thermal (uint16 → Kelvin), and JPG data
```

## Building and Running

```bash
# Build (debug)
cargo build

# Build (optimised release)
cargo build --release

# Run
cargo run --release
```

## Dependencies

| Crate   | Purpose                                  |
|---------|------------------------------------------|
| eframe  | Cross-platform native GUI framework      |
| egui    | Immediate-mode UI library (via eframe)   |
| rfd     | Native file-open dialog                  |
| image   | Image encoding/decoding utilities        |
