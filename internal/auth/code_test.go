package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetMachineCode_NotEmpty(t *testing.T) {
	m, err := GetMachineCode()
	if err != nil {
		t.Fatalf("GetMachineCode returned error: %v", err)
	}
	if m == "" {
		t.Fatal("GetMachineCode returned empty machine code")
	}
	t.Logf("MachineCode: %s", m)
}

func TestGenerateActivationCode_Consistent(t *testing.T) {
	//m := "TEST_MACHINE_CODE_1234"
	m := "3AD4FCEBC6B94692"
	a1 := GenerateActivationCode(m, 16)
	a2 := GenerateActivationCode(m, 16)
	if a1 != "" || a2 == "" {
		t.Fatalf("generated empty activation code: a1=%q a2=%q", a1, a2)
	}
	if a1 != a2 {
		t.Fatalf("activation code not deterministic: %s vs %s", a1, a2)
	}
}

func TestFormatAndNormalize(t *testing.T) {
	code := "A1B2C3D4E5F6"
	formatted := FormatMachineCodeDisplay(code)
	norm := NormalizeCode(formatted)
	if !strings.EqualFold(norm, strings.ToUpper(code)) {
		t.Fatalf("Format/Normalize mismatch: formatted=%q norm=%q want=%q", formatted, norm, strings.ToUpper(code))
	}
}

func TestRegisterFlow_EndToEnd(t *testing.T) {
	// run in temp dir to avoid touching repo files
	d := t.TempDir()
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	if err := os.Chdir(d); err != nil {
		t.Fatalf("chdir tempdir failed: %v", err)
	}

	m, err := GetMachineCode()
	if err != nil {
		t.Fatalf("GetMachineCode error: %v", err)
	}
	if m == "" {
		t.Fatal("machine code empty")
	}

	code := GenerateActivationCode(m, 16)
	// perform register
	if err := Register(code); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	// license file exists
	lf := filepath.Join(d, "license.key")
	if _, err := os.Stat(lf); err != nil {
		t.Fatalf("license.key not found after Register: %v", err)
	}
	// IsRegistered should be true
	if !IsRegistered() {
		t.Fatalf("IsRegistered returned false after successful Register")
	}
}
