# ICU - Internal Catalog Utility

A CLI tool for fetching and querying satellite catalog data (TLE and SATCAT) from spacebook.com.

## Installation

Build the binary:
```bash
go build -o icu
```

Install to your system:
```bash
# Linux/macOS
sudo mv icu /usr/local/bin/

# Or for user-only install
mkdir -p ~/bin
mv icu ~/bin/
export PATH="$HOME/bin:$PATH"  # Add to ~/.bashrc or ~/.zshrc
```

Verify installation:
```bash
icu --help
```

## Usage

### Fetch catalog data

```bash
icu fetch
```

### Get satellite by NORAD ID

```bash
icu get --norad 25544
```

Output (3-line TLE format):
```
0 ISS (ZARYA)
1 25544U 98067A   26036.24398336  .00012313  00000-0  23569-3 0  9993
2 25544  51.6316 232.7490 0011153  66.4623 293.7536 15.48407263551308
```

### Get satellite by name

```bash
icu get --name "ISS (ZARYA)"
```

### Verbose output

```bash
icu get --norad 25544 --verbose
```

### Search for satellites

```bash
# Search by partial name
icu search --name "starlink"

# Search with filters
icu search --name "ISS" --type "payload"

# Limit results
icu search --name "starlink" --limit 100

# Show detailed results
icu search --name "starlink" --verbose
```

### View catalog statistics

```bash
icu stats
```
