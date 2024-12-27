# HFinder 🚀

HFinder is a lightweight Go-based tool for scanning CIDR ranges, extracting hostnames, and performing efficient network reconnaissance. 

## Features 🌟
- **Stdin/Stdout Support:** Seamlessly process input from stdin and output results to stdout for flexible integrations.

- **Hostname Extraction:** Extract valid hostnames from HTML responses.
- **Silent Mode:** Suppress unnecessary output for clean and focused results.
- **File Support:** Provide a list of CIDR ranges via a file for batch processing.
- **Automatic Cleanup:** Deletes cache files after processing to save space.

## Installation 📦
Install HFinder using Go:
```bash
go install github.com/0xm1d/Hfinder@latest
```
Make sure `GOPATH` is set up correctly and the binary is available in your `PATH`.

## Usage 🛠️
### Scan a Single CIDR Range
```bash
HFinder -r 192.168.0.0/24
```

### Process a File with Multiple CIDR Ranges
```bash
HFinder -l cidr_list.txt
```

### Enable Silent Mode
```bash
HFinder -r 192.168.0.0/24 -silent
```

## License 📜
HFinder is licensed under the GNU General Public License v3.0. See the `LICENSE` file for more details.

## Contributing 🤝
Contributions are welcome! Feel free to open an issue or submit a pull request.

## Contact 📧
For questions or support, reach out on our [X](https://x.com/0xM1D_).
