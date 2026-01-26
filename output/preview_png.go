package output

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ErrBrowserUnavailable indicates Chrome/Chromium isn't available to generate a PNG preview.
var ErrBrowserUnavailable = errors.New("browser unavailable for png preview")

// DefaultPNGPreviewPath returns a sibling .png path for an .xlsx path.
// If xlsxPath doesn't end with an extension, it appends ".png".
func DefaultPNGPreviewPath(xlsxPath string) string {
	ext := filepath.Ext(xlsxPath)
	if ext == "" {
		return xlsxPath + ".png"
	}
	return strings.TrimSuffix(xlsxPath, ext) + ".png"
}

// IsBrowserUnavailable reports whether err is caused by lack of a browser for PNG rendering.
func IsBrowserUnavailable(err error) bool {
	return errors.Is(err, ErrBrowserUnavailable)
}

func hasChromeLikeBinary() bool {
	candidates := []string{
		"google-chrome",
		"google-chrome-stable",
		"chromium",
		"chromium-browser",
		"chrome",
		"msedge",
		"microsoft-edge",
	}
	for _, name := range candidates {
		if _, err := exec.LookPath(name); err == nil {
			return true
		}
	}
	return false
}

// WritePNGPreview writes a PNG "print screen" style preview of the final output.
// It renders the same layout as the XLSX into HTML and uses a headless browser to screenshot it.
//
// Note: requires Chrome/Chromium available on the system. If not present, returns ErrBrowserUnavailable.
func WritePNGPreview(filePath string, data *OutputData) error {
	if !hasChromeLikeBinary() {
		return fmt.Errorf("%w: no Chrome/Chromium found on PATH", ErrBrowserUnavailable)
	}

	tmpDir, err := os.MkdirTemp("", "qa-preview-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	htmlPath := filepath.Join(tmpDir, "preview.html")
	if err := WriteHTMLPreview(htmlPath, data); err != nil {
		return fmt.Errorf("failed to write intermediate html preview: %w", err)
	}

	absHTML, err := filepath.Abs(htmlPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute html path: %w", err)
	}

	u := url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(absHTML),
	}

	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.WindowSize(1400, 900),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Avoid hangs.
	ctx, cancelTimeout := context.WithTimeout(ctx, 15*time.Second)
	defer cancelTimeout()

	var png []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(u.String()),
		chromedp.WaitReady("table.sheet", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, err := page.CaptureScreenshot().
				WithFormat(page.CaptureScreenshotFormatPng).
				WithCaptureBeyondViewport(true).
				Do(ctx)
			if err != nil {
				return err
			}
			png = buf
			return nil
		}),
	); err != nil {
		// Most common cause: browser can't start in this environment.
		return fmt.Errorf("%w: %v", ErrBrowserUnavailable, err)
	}

	if err := os.WriteFile(filePath, png, 0644); err != nil {
		return fmt.Errorf("failed to write png preview: %w", err)
	}
	return nil
}
