package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	i2pconv "github.com/go-i2p/go-i2ptunnel-config/i2pconv"
	"github.com/urfave/cli/v2"
)

// newApp builds the same cli.App that main() creates, so tests exercise flag
// parsing and routing through ConvertCommand without spawning a subprocess.
func newApp() *cli.App {
	return &cli.App{
		Name: "go-i2ptunnel-config",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "in-format", Aliases: []string{"if"}},
			&cli.StringFlag{Name: "out-format", Aliases: []string{"of"}, Value: "yaml"},
			&cli.StringFlag{Name: "output", Aliases: []string{"o"}},
			&cli.BoolFlag{Name: "validate"},
			&cli.BoolFlag{Name: "strict"},
			&cli.BoolFlag{Name: "dry-run"},
			&cli.BoolFlag{Name: "batch"},
			&cli.BoolFlag{Name: "sam"},
			&cli.StringFlag{Name: "keystore"},
			&cli.BoolFlag{Name: "split"},
			&cli.BoolFlag{Name: "list-tunnels"},
		},
		Action: i2pconv.ConvertCommand,
	}
}

// TestMain_NoArgs checks that calling the tool without arguments returns an error.
func TestMain_NoArgs(t *testing.T) {
	app := newApp()
	err := app.Run([]string{"go-i2ptunnel-config"})
	if err == nil {
		t.Fatal("expected error when no arguments provided, got nil")
	}
}

// TestMain_BatchOutputConflict checks that --batch and --output together return an error.
func TestMain_BatchOutputConflict(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a.yaml")
	content := "tunnels:\n  myTunnel:\n    name: myTunnel\n    type: httpclient\n    port: 4444\n"
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	app := newApp()
	err := app.Run([]string{
		"go-i2ptunnel-config",
		"--batch",
		"--output", filepath.Join(dir, "out.yaml"),
		f,
	})
	if err == nil {
		t.Fatal("expected error for --batch + --output conflict, got nil")
	}
}

// TestMain_EndToEndConversion writes a properties temp file and checks that the
// YAML output file is created at the auto-generated path.
func TestMain_EndToEndConversion(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "tunnel.properties")
	content := "tunnel.0.name=myTunnel\ntunnel.0.type=httpclient\ntunnel.0.listenPort=4444\n"
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	app := newApp()
	if err := app.Run([]string{"go-i2ptunnel-config", src}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(dir, "tunnel.yaml")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Errorf("expected output file %s was not created", expected)
	}
}

// TestMain_Validate checks that --validate succeeds for a valid file and fails
// for an invalid one without writing any output.
func TestMain_Validate(t *testing.T) {
	dir := t.TempDir()

	valid := filepath.Join(dir, "valid.yaml")
	os.WriteFile(valid, []byte("tunnels:\n  t:\n    name: t\n    type: httpclient\n    port: 4444\n"), 0o644)

	invalid := filepath.Join(dir, "invalid.yaml")
	// A tunnel with type httpclient must have port > 0; omit port to trigger validation failure.
	os.WriteFile(invalid, []byte("tunnels:\n  t:\n    name: t\n    type: httpclient\n"), 0o644)

	app := newApp()
	if err := app.Run([]string{"go-i2ptunnel-config", "--validate", valid}); err != nil {
		t.Errorf("validate on valid file: unexpected error: %v", err)
	}

	if err := app.Run([]string{"go-i2ptunnel-config", "--validate", invalid}); err == nil {
		t.Error("validate on invalid file: expected error, got nil")
	}

	// No output files should have been produced.
	matches, _ := filepath.Glob(filepath.Join(dir, "*.yaml"))
	// The two input .yaml files are expected; no additional ones should appear.
	for _, m := range matches {
		base := filepath.Base(m)
		if base != "valid.yaml" && base != "invalid.yaml" {
			t.Errorf("unexpected output file created: %s", m)
		}
	}
}

// TestMain_DryRunStdout checks that --dry-run produces output on stdout without
// writing any file.
func TestMain_DryRunStdout(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "tunnel.properties")
	content := "tunnel.0.name=myTunnel\ntunnel.0.type=httpclient\ntunnel.0.listenPort=4444\n"
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Capture stdout.
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := newApp()
	runErr := app.Run([]string{"go-i2ptunnel-config", "--dry-run", src})

	w.Close()
	os.Stdout = origStdout
	var buf strings.Builder
	tmp := make([]byte, 4096)
	for {
		n, err := r.Read(tmp)
		buf.Write(tmp[:n])
		if err != nil {
			break
		}
	}
	r.Close()

	if runErr != nil {
		t.Fatalf("unexpected error: %v", runErr)
	}
	if !strings.Contains(buf.String(), "myTunnel") {
		t.Errorf("dry-run stdout did not contain 'myTunnel'; got:\n%s", buf.String())
	}

	// No output file should be written.
	expected := filepath.Join(dir, "tunnel.yaml")
	if _, err := os.Stat(expected); err == nil {
		t.Errorf("dry-run must not write %s", expected)
	}
}
