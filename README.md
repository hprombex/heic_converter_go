# HEIC Converter

HEIC Converter is a command-line utility for converting HEIC images to JPEG or PNG formats. The tool supports batch processing of files in a directory, allows specifying output quality, and optionally deletes the original HEIC files after conversion.

## Features
- Convert individual HEIC files or all HEIC files in a directory.
- Output formats: JPEG or PNG.
- Adjustable JPEG quality (1-100).
- Optionally delete the original HEIC files after conversion.

## Requirements

- Go 1.18 or later
- [libheif](https://github.com/strukturag/libheif)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/hprombex/heic-converter.git
   cd heic-converter
   ```
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Build the executable:
   ```bash
   go build -o heic-converter
   ```

## Usage

Run the HEIC Converter with the following command-line options:

```bash
./heic-converter [options]
```

### Options

- `--input_file` (string): Path to a single HEIC file to be converted.
- `--input_dir` (string): Path to a directory containing HEIC files.
- `--output_path` (string): Path to the output file or directory.
- `--delete` (boolean): Delete the original file after conversion. Default: `false`.
- `--format` (string): Output image format (`jpeg` or `png`). Default: `jpeg`.
- `--quality` (int): Quality of the output JPEG image (1-100). Default: `80`.

### Examples

#### Convert a Single HEIC File
```bash
./heic-converter --input_file input.heic --output_path output.jpg --format jpeg --quality 90
```

#### Convert All HEIC Files in a Directory
```bash
./heic-converter --input_dir ./images --output_path ./converted --format png
```

#### Convert and Delete Original Files
```bash
./heic-converter --input_dir ./images --output_path ./converted --delete
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Author

hprombex

Feel free to contribute or suggest improvements!

