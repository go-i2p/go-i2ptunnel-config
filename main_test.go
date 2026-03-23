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

// TestMain_StdinDryRun checks that --in-format yaml --dry-run - reads from stdin
// and prints converted output without writing any file.
func TestMain_StdinDryRun(t *testing.T) {
	yamlContent := "tunnels:\n  stdinTunnel:\n    name: stdinTunnel\n    type: httpclient\n    port: 4444\n"

	// Replace stdin with a pipe pre-loaded with YAML content.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	if _, err := w.WriteString(yamlContent); err != nil {
		t.Fatalf("write to pipe: %v", err)
	}
	w.Close()

	origStdin := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = origStdin
		r.Close()
	}()

	// Capture stdout.
	origStdout := os.Stdout
	ro, rw, _ := os.Pipe()
	os.Stdout = rw

	app := newApp()
	runErr := app.Run([]string{
		"go-i2ptunnel-config",
		"--in-format", "yaml",
		"--dry-run",
		"-",
	})

	rw.Close()
	os.Stdout = origStdout
	var buf strings.Builder
	tmp := make([]byte, 4096)
	for {
		n, readErr := ro.Read(tmp)
		buf.Write(tmp[:n])
		if readErr != nil {
			break
		}
	}
	ro.Close()

	if runErr != nil {
		t.Fatalf("unexpected error: %v", runErr)
	}
	if !strings.Contains(buf.String(), "stdinTunnel") {
		t.Errorf("stdin dry-run output did not contain 'stdinTunnel'; got:\n%s", buf.String())
	}
}

// TestMain_SplitDryRun checks that --split --dry-run on a 3-section INI file
// prints output for all three tunnels without writing any files.
func TestMain_SplitDryRun(t *testing.T) {
	dir := t.TempDir()
	iniContent := "[tunnel-a]\ntype = client\nport = 1111\n\n[tunnel-b]\ntype = client\nport = 2222\n\n[tunnel-c]\ntype = client\nport = 3333\n"
	src := filepath.Join(dir, "multi.conf")
	if err := os.WriteFile(src, []byte(iniContent), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Capture stdout.
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := newApp()
	runErr := app.Run([]string{
		"go-i2ptunnel-config",
		"--split",
		"--dry-run",
		src,
	})

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
	output := buf.String()
	for _, name := range []string{"tunnel-a", "tunnel-b", "tunnel-c"} {
		if !strings.Contains(output, name) {
			t.Errorf("split dry-run output missing tunnel '%s'; got:\n%s", name, output)
		}
	}
}

// TestMain_ListTunnels checks that --list-tunnels on a 3-section INI file
// prints all three tunnel names without error and without writing files.
func TestMain_ListTunnels(t *testing.T) {
	dir := t.TempDir()
	iniContent := "[alpha]\ntype = client\nport = 1111\n\n[beta]\ntype = server\ndestination = example.b32.i2p\n\n[gamma]\ntype = client\nport = 3333\n"
	src := filepath.Join(dir, "tunnels.conf")
	if err := os.WriteFile(src, []byte(iniContent), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Capture stdout.
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	app := newApp()
	runErr := app.Run([]string{
		"go-i2ptunnel-config",
		"--list-tunnels",
		src,
	})

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
	output := buf.String()
	for _, name := range []string{"alpha", "beta", "gamma"} {
		if !strings.Contains(output, name) {
			t.Errorf("--list-tunnels output missing tunnel '%s'; got:\n%s", name, output)
		}
	}

	// No output files should have been created in the temp dir.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() != "tunnels.conf" {
			t.Errorf("unexpected file created: %s", e.Name())
		}
	}
}
