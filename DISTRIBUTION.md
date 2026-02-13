# Distribution Guide for QA Location Grouper

## Problem

Windows Defender flags the compiled `.exe` as a false positive virus detection. This is common with unsigned Go binaries, especially those built with Wails (which embed a WebView2 runtime).

## Strategies to Bypass Windows Defender False Positives

### 1. Windows Defender Folder Exclusion (Easiest - Requires IT Help)

Ask IT to add a **folder exclusion** in Windows Defender for a specific directory (e.g., `C:\Tools\qa-script\`). Anything placed in that folder won't be scanned.

**What to tell IT:**

> "We have an internal Go-based QA tool that Defender flags as a false positive. Can you add a folder exclusion for `C:\Tools\qa-script\` on 3 specific machines?"

This is a simple Group Policy or Defender settings change and is a very common request for internal tools.

### 2. Submit a False Positive Report to Microsoft

Submit the `.exe` to Microsoft for review at:

- <https://www.microsoft.com/en-us/wdsi/filesubmission>

They'll analyze it and if it's clean, they'll whitelist the hash. This can take a few days but is free. Note: you may need to resubmit after each new build since the hash changes.

### 3. Distribute the Web GUI Instead of the Desktop GUI

The project already has a **Web GUI** (`cmd/webgui/`) that runs as a local web server and opens in the user's default browser. This binary is much simpler than the Wails GUI (no embedded WebView2), which makes it **far less likely to be flagged** by Defender.

**Build it:**

```bash
make build-webgui-windows
```

**Distribute:** Just give users `qa-webgui-windows-amd64.exe` — they double-click it and it opens in their browser. No special runtime required.

**Why this helps:**
- Smaller, simpler binary (no embedded WebView2 or complex resource packing)
- Antivirus heuristics are less suspicious of plain HTTP server binaries
- Same user experience — browser-based UI works identically

### 4. Go Build Tweaks to Reduce False Positives

These build flags add metadata and clean up the binary, reducing AV false positive rates:

#### Add version info with `-ldflags`

```bash
GOOS=windows GOARCH=amd64 go build \
  -ldflags="-s -w -X main.version=1.0.0 -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -trimpath \
  -o bin/qa-webgui-windows-amd64.exe \
  ./cmd/webgui/
```

- `-trimpath` removes local file paths from the binary (removes "suspicious" build paths)
- `-s -w` strips debug symbols (smaller, cleaner binary)

#### Add a Windows manifest and icon via `.syso` resource

Use [go-winres](https://github.com/tc-hib/go-winres) to embed a proper Windows manifest, version info, and icon into the binary. This makes the `.exe` look like a legitimate application to both Windows and AV scanners.

```bash
# Install go-winres
go install github.com/tc-hib/go-winres@latest

# Initialize a winres directory with a manifest template
go-winres init

# Edit winres/winres.json to set your app name, version, company, etc.
# Then build — go-winres generates a .syso file that go build picks up automatically
go-winres make
go build -o bin/qa-webgui-windows-amd64.exe ./cmd/webgui/
```

#### Sign the binary (if your company has a code signing certificate)

A signed binary will almost never be flagged. If your company has a code signing certificate:

```powershell
signtool sign /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 /f cert.pfx /p <password> qa-webgui-windows-amd64.exe
```

Even a **self-signed certificate** installed on the 3 target machines can help, though a proper code signing cert is the gold standard.

### 5. Wrap in an MSI/MSIX Installer

Windows trusts installer packages more than loose `.exe` files. Tools like [WiX Toolset](https://wixtoolset.org/) (free) or [Inno Setup](https://jrsoftware.org/isinfo.php) (free) can wrap your binary into an installer. Combined with code signing, this virtually eliminates false positives.

### 6. Use GitHub Releases with Checksums

Attach the `.exe` to a [GitHub Release](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository) along with a SHA256 checksum file. This gives IT a way to verify the file is legitimate and untampered.

```bash
# Generate checksum
shasum -a 256 bin/qa-webgui-windows-amd64.exe > bin/checksums.txt

# Create a release (requires gh CLI)
gh release create v1.0.0 \
  bin/qa-webgui-windows-amd64.exe \
  bin/checksums.txt \
  --title "v1.0.0" \
  --notes "QA Location Grouper v1.0.0"
```

## Free Distribution Platforms (For a Small Team)

| Platform | Notes |
| --- | --- |
| **Private GitHub repo** | Free for small teams, great for versioning and releases |
| **Private GitLab repo** | Free, unlimited private repos |
| **Google Drive / OneDrive** | Simple shared folder, everyone grabs the latest |
| **SharePoint** | Likely already available if your company uses Microsoft 365 |

## Recommended Approach

For a team of 3 with no admin rights on the target machines:

1. **Ask IT for a folder exclusion** on the 3 machines — simplest permanent fix.
2. **Distribute the Web GUI binary** instead of the Wails GUI — simpler binary, less likely to trigger AV.
3. **Add a Windows manifest** with `go-winres` — makes the binary look legitimate to Windows.
4. **Use a private GitHub repo** with Releases to distribute — free version control and integrity verification.
5. As a nuclear option, **submit to Microsoft** for false positive review after each release.
